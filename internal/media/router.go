package media

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"stream_hub/internal/infra"
	"stream_hub/pkg/model/config"
)

type MediaRouter struct {
	router *gin.Engine
	media  *MediaApi
	port   int
}

func NewMediaRouter(base *infra.Base, conf *config.MediaConfig) *MediaRouter {
	r := new(MediaRouter)
	r.media = NewMediaApi(base)
	r.port = conf.Port
	r.init()

	return r
}

func (r *MediaRouter) init() {
	r.router = gin.Default()
	r.router.POST("upload_image", r.media.UploadImage)
}

func (r *MediaRouter) Run() error {
	return r.router.Run(fmt.Sprintf(":%d", r.port))
}
