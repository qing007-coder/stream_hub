package media

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"stream_hub/internal/infra"
	"stream_hub/internal/security"
	"stream_hub/pkg/model/config"
)

type MediaRouter struct {
	router     *gin.Engine
	media      *MediaApi
	middleware *Middleware
	port       int
}

func NewMediaRouter(base *infra.Base, conf *config.MediaConfig, auth *security.Auth) *MediaRouter {
	r := new(MediaRouter)
	r.media = NewMediaApi(base)
	r.middleware = NewMiddleware(base, auth)
	r.port = conf.Port
	r.init()

	return r
}

func (r *MediaRouter) init() {
	r.router = gin.Default()
	media := r.router.Group("/media").Use(r.middleware.Cors(), r.middleware.LogToStorage(), r.middleware.Auth())
	{
		media.POST("upload_image", r.media.UploadImage)
	}
}

func (r *MediaRouter) Run() error {
	return r.router.Run(fmt.Sprintf(":%d", r.port))
}
