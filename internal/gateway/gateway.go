package gateway

import (
	"stream_hub/internal/infra"
	"stream_hub/internal/proto/interaction"
	"stream_hub/internal/proto/video"
	"stream_hub/pkg/model/config"
)

type Gateway struct {
	base              *infra.Base
	srv               *Service
	videoClient       video.VideoService
	interactionClient interaction.InteractionService
}

func NewGateway(base *infra.Base, srv *Service, conf *config.GatewayConfig) *Gateway {

	videoClient := video.NewVideoService(conf.Service.VideoService, srv.Client())
	interactionClient := interaction.NewInteractionService(conf.Service.InteractionService, srv.Client())

	return &Gateway{
		base:              base,
		videoClient:       videoClient,
		interactionClient: interactionClient,
		srv:               srv,
	}
}
