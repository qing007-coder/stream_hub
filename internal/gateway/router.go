package gateway

import (
	"fmt"
	"github.com/gin-gonic/gin"
	f "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
	_ "stream_hub/docs_api"
	"stream_hub/internal/infra"
	"stream_hub/internal/security"
	"stream_hub/pkg/model/config"
)

// @title Gateway API
// @version 1.0
// @description Gateway API for Stream Hub
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token

type GatewayRouter struct {
	router     *gin.Engine
	gateway    *Gateway
	middleware *Middleware
	port       int
}

func NewGatewayRouter(base *infra.Base, auth *security.Auth, ratelimiter *infra.Ratelimter, commonConf *config.CommonConfig, conf *config.GatewayConfig) *GatewayRouter {
	srv := NewService(commonConf)

	router := &GatewayRouter{
		port:       conf.Port,
		gateway:    NewGateway(base, srv, conf),
		middleware: NewMiddleware(base, ratelimiter, auth),
	}

	router.init()

	return router
}

func (r *GatewayRouter) init() {
	r.router = gin.Default()
	api := r.router.Group("/api")
	api.Use(r.middleware.Cors(), r.middleware.Ratelimit(), r.middleware.LogToStorage())
	{
		// Video API
		video := api.Group("/video")
		{
			// 所有接口都需要认证
			video.POST("/create", r.middleware.Auth(), r.gateway.CreateVideo)
			video.GET("/get/:video_id", r.middleware.Auth(), r.gateway.GetVideo)
			video.PUT("/update/:video_id", r.middleware.Auth(), r.gateway.UpdateVideo)
			video.DELETE("/delete/:video_id", r.middleware.Auth(), r.gateway.DeleteVideo)
			video.GET("/list/:user_id", r.middleware.Auth(), r.gateway.ListUserPublishedVideos)
			video.GET("/my/list", r.middleware.Auth(), r.gateway.ListMyVideos)
		}

		// Interaction API
		interaction := api.Group("/interaction")
		{
			// 所有接口都需要认证
			interaction.POST("/like/:video_id", r.middleware.Auth(), r.gateway.CreateLike)
			interaction.DELETE("/like/:video_id", r.middleware.Auth(), r.gateway.DeleteLike)
			interaction.GET("/like/:video_id", r.middleware.Auth(), r.gateway.IsLike)
			interaction.GET("/likes/:video_id", r.middleware.Auth(), r.gateway.ListLikes)

			interaction.POST("/favorite/:video_id", r.middleware.Auth(), r.gateway.CreateFavorite)
			interaction.DELETE("/favorite/:video_id", r.middleware.Auth(), r.gateway.DeleteFavorite)
			interaction.GET("/favorite/:video_id", r.middleware.Auth(), r.gateway.IsFavorite)

			interaction.POST("/follow/:target_user_id", r.middleware.Auth(), r.gateway.CreateFollow)
			interaction.DELETE("/follow/:target_user_id", r.middleware.Auth(), r.gateway.DeleteFollow)
			interaction.GET("/follow/:target_user_id", r.middleware.Auth(), r.gateway.IsFollow)
			interaction.GET("/followers/:user_id", r.middleware.Auth(), r.gateway.ListFollowers)
			interaction.GET("/followings/:user_id", r.middleware.Auth(), r.gateway.ListFollowings)

			interaction.POST("/comment", r.middleware.Auth(), r.gateway.CreateComment)
			interaction.DELETE("/comment/:comment_id", r.middleware.Auth(), r.gateway.DeleteComment)
			interaction.GET("/comments/:video_id", r.middleware.Auth(), r.gateway.ListComments)
		}
	}

	// Swagger 路由
	r.router.GET("/swagger/*any", ginswagger.WrapHandler(f.Handler))
}

func (r *GatewayRouter) Run() error {
	return r.router.Run(fmt.Sprintf(":%d", r.port))
}
