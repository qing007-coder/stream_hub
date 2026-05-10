package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"stream_hub/internal/infra"
	"stream_hub/internal/security"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
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
		origin := c.Request.Header.Get("Origin")

		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
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
				ctx.Abort()
				return
			}

			utils.UnAuthorizationRequest(ctx, "token invalid")
			ctx.Abort()
			return
		}

		if claims.Role != "admin" {
			utils.ForbiddenRequest(ctx, "permission denied")
			ctx.Abort()
			return
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Set("role", claims.Role)
	}
}

func (m *Middleware) AdminOperationLog() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if ctx.Writer.Status() >= 400 {
			return
		}

		adminID := ctx.GetString("user_id")
		adminEmail := ""

		if adminID != "" {
			var admin storage.Admin
			if err := m.DB.Where("id = ?", adminID).First(&admin).Error; err == nil {
				adminEmail = admin.Email
			}
		}

		action := getAdminAction(ctx.Request.Method, ctx.Request.URL.Path)
		targetType, targetID := getTargetInfo(ctx)

		logEntry := storage.AdminLogEntry{
			EventTime:  float64(time.Now().UnixMicro()) / 1000000.0,
			Level:      "info",
			AdminID:    adminID,
			AdminEmail: adminEmail,
			IP:         ctx.ClientIP(),
			Action:     action,
			TargetType: targetType,
			TargetID:   targetID,
			Detail:     fmt.Sprintf("%s %s", ctx.Request.Method, ctx.Request.URL.Path),
			Result:     "success",
			Module:     constant.Admin,
		}

		if ctx.Writer.Status() >= 400 {
			logEntry.Level = "error"
			logEntry.Result = "failed"
		}

		logs := []storage.AdminLogEntry{logEntry}
		if err := m.Clickhouse.BatchInsertStruct(context.Background(), constant.StorageAdminLog, logs); err != nil {
			m.Logger.Error(
				"admin operation log failed",
				adminID,
				"",
				"",
				ctx.Request.Method,
				ctx.Request.URL.Path,
				constant.Admin,
				int16(ctx.Writer.Status()),
				0,
			)
		}
	}
}

func getAdminAction(method, path string) string {
	switch method + " " + path {
	case "POST /admin/auth/login":
		return "admin_login"
	case "POST /admin/auth/logout":
		return "admin_logout"
	case "POST /admin/auth/refresh":
		return "refresh_token"
	case "GET /admin/stats":
		return "view_dashboard"
	case "GET /admin/users/list":
		return "list_users"
	case "GET /admin/users/detail":
		return "view_user_detail"
	case "PUT /admin/users/status":
		return "update_user_status"
	case "GET /admin/videos/list":
		return "list_videos"
	case "GET /admin/videos/detail":
		return "view_video_detail"
	case "PUT /admin/videos/update":
		return "update_video"
	case "GET /admin/videos/interaction":
		return "view_video_interaction"
	case "GET /admin/videos/pending-audit":
		return "list_pending_audit"
	case "POST /admin/videos/audit":
		return "audit_video"
	default:
		return "unknown_action"
	}
}

func getTargetInfo(ctx *gin.Context) (targetType, targetID string) {
	path := ctx.Request.URL.Path

	switch {
	case contains(path, "/users/detail/"):
		targetType = "user"
		targetID = ctx.Param("id")
	case contains(path, "/videos/detail/"):
		targetType = "video"
		targetID = ctx.Param("id")
	case contains(path, "/videos/update/"):
		targetType = "video"
		targetID = ctx.Param("id")
	case contains(path, "/videos/interaction/"):
		targetType = "video"
		targetID = ctx.Param("id")
	case contains(path, "/videos/audit/"):
		targetType = "video"
		targetID = ctx.Param("id")
	default:
		targetType = ""
		targetID = ""
	}

	return
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
