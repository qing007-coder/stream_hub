package user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"stream_hub/internal/infra"
	"stream_hub/internal/security"
	"stream_hub/pkg/model/config"
)

type UserRouter struct {
	router     *gin.Engine
	user       *UserApi
	middleware *Middleware
	port       int
}

func NewUserRouter(base *infra.Base, auth *security.Auth, conf *config.UserConfig) *UserRouter {
	router := &UserRouter{
		port:       conf.Port,
		user:       NewUserApi(base, auth),
		middleware: NewMiddleware(base, auth),
	}

	router.init()

	return router
}

func (r *UserRouter) init() {
	r.router = gin.Default()
	r.router.Use(r.middleware.Cors(), r.middleware.LogToStorage())

	user := r.router.Group("/user")
	{
		// 登录 / 注册
		user.POST("/login", r.user.Login)
		user.POST("/register", r.user.Register)

		// token
		user.POST("/refresh", r.user.RefreshToken)
		user.POST("/logout", r.user.Logout)

		// 邮箱验证码
		user.POST("/send_verification_code", r.user.SendVerificationCode)

		user.PUT("/update_profile", r.middleware.Auth(), r.user.UpdateProfile)
		user.PUT("/password", r.middleware.Auth(), r.user.UpdatePassword)
		user.GET("/get_user_profile", r.middleware.Auth(), r.user.GetUserProfile)
	}
}

func (r *UserRouter) Run() error {
	return r.router.Run(fmt.Sprintf(":%d", r.port))
}
