package media

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"stream_hub/internal/infra"
	"stream_hub/internal/security"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/utils"
	"time"
)

type Middleware struct {
	auth *security.Auth
	*infra.Base
}

func NewMiddleware(base *infra.Base, auth *security.Auth) *Middleware {
	return &Middleware{
		auth: auth,
		Base: base,
	}
}

func (m *Middleware) Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func (m *Middleware) LogToStorage() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		ctx.Next()

		clientIp := ctx.ClientIP()
		reqMethod := ctx.Request.Method
		path := ctx.Request.URL.Path
		uid := ctx.GetString("user_id")
		traceID := ctx.GetString("trace_id")
		status := int16(ctx.Writer.Status())
		latency := time.Since(start).Microseconds()

		if status >= 400 && status < 500 {
			m.Logger.Warn(
				"client side error",
				clientIp,
				uid,
				traceID,
				reqMethod,
				path,
				constant.Media,
				status,
				latency,
			)
		} else if status >= 500 {
			m.Logger.Error(
				"server side error",
				clientIp,
				uid,
				traceID,
				reqMethod,
				path,
				constant.Media,
				status,
				latency,
			)
		}

		m.Logger.Info(
			"request processed",
			clientIp,
			uid,
			traceID,
			reqMethod,
			path,
			constant.Media,
			status,
			latency,
		)
	}
}

func (m *Middleware) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			utils.UnAuthorizationRequest(ctx, "need token")
			ctx.Abort()
			return
		}

		claims, err := m.auth.ParseToken(token)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				utils.UnAuthorizationRequest(ctx, "token expired")
				return
			}

			utils.UnAuthorizationRequest(ctx, "token invalid")
		}

		ctx.Set("user_id", claims.UserID)
	}
}
