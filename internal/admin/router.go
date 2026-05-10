package admin

import (
	"fmt"
	"stream_hub/internal/infra"
	"stream_hub/internal/security"
	"stream_hub/pkg/model/config"

	"github.com/gin-gonic/gin"
)

type AdminRouter struct {
	router     *gin.Engine
	middleware *Middleware
	auth       *AuthController
	user       *UserController
	video      *VideoController
	task       *TaskController
	api        *ApiController
	port       int
}

func NewAdminRouter(base *infra.Base, auth *security.Auth, conf *config.AdminConfig) *AdminRouter {
	r := new(AdminRouter)
	r.middleware = NewMiddleware(base, auth)
	r.auth = NewAuthController(base, auth)
	r.user = NewUserController(base)
	r.video = NewVideoController(base)
	r.task = NewTaskController(base)
	r.api = NewApiController(base)
	r.port = conf.Port
	r.init()

	return r
}

func (r *AdminRouter) init() {
	r.router = gin.Default()
	r.router.Use(r.middleware.Cors())

	r.router.GET("/health", r.api.HealthCheck)

	authGroup := r.router.Group("/admin/auth")
	{
		authGroup.POST("/login", r.auth.Login)
		authGroup.POST("/refresh", r.auth.RefreshToken)
		authGroup.POST("/logout", r.auth.Logout)
	}

	admin := r.router.Group("/admin")
	admin.Use(r.middleware.Auth(), r.middleware.AdminOperationLog())
	{
		admin.GET("/stats", r.api.GetDashboardStats)

		users := admin.Group("/users")
		{
			users.GET("/list", r.user.GetUserList)
			users.GET("/detail/:id", r.user.GetUserDetail)
			users.PUT("/status", r.user.UpdateUserStatus)
		}

		admins := admin.Group("/admins")
		{
			admins.GET("/list", r.user.GetAdminList)
			admins.POST("/create", r.user.CreateAdmin)
			admins.PUT("/status/:id", r.user.UpdateAdminStatus)
			admins.DELETE("/delete/:id", r.user.DeleteAdmin)
		}

		videos := admin.Group("/videos")
		{
			videos.GET("/list", r.video.GetVideoList)
			videos.GET("/detail/:id", r.video.GetVideoDetail)
			videos.PUT("/update/:id", r.video.UpdateVideo)
			videos.GET("/interaction/:id", r.video.GetVideoInteractionDetail)
			videos.GET("/pending-audit", r.video.GetPendingAuditList)
			videos.POST("/audit/:id", r.video.AuditVideo)
		}

		tasks := admin.Group("/tasks")
		{
			tasks.GET("/list", r.task.GetTaskList)
			tasks.GET("/detail/:id", r.task.GetTaskDetail)
			tasks.PUT("/status", r.task.UpdateTaskStatus)
		}
	}
}

func (r *AdminRouter) Run() error {
	return r.router.Run(fmt.Sprintf(":%d", r.port))
}
