package video

import (
	"fmt"
	grpcc "github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/consul"
	grpcs "github.com/go-micro/plugins/v4/server/grpc"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"stream_hub/internal/infra"
	"stream_hub/internal/proto/video"
	"stream_hub/pkg/model/config"
)

type Server struct {
	srv     micro.Service
	video   *Video
	wrapper *Wrapper
	port    int
	name    string
}

func NewServer(base *infra.Base, commonConf *config.CommonConfig, videoConf *config.VideoConfig) (*Server, error) {
	s := &Server{
		port:    videoConf.Port,
		name:    videoConf.Name,
		video:   NewVideo(base),
		wrapper: NewWrapper(),
	}

	if err := s.init(commonConf); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) init(conf *config.CommonConfig) error {
	c := consul.NewRegistry(
		registry.Addrs(fmt.Sprintf("%s:%s", conf.Consul.Addr, conf.Consul.Port)),
	)

	s.srv = micro.NewService(
		micro.Server(grpcs.NewServer()),
		micro.Client(grpcc.NewClient()), // 使用 gRPC client
		micro.Name(s.name),
		micro.Version("latest"),
		micro.Registry(c),                      // 必须放底下哎，不然注册中心的优先级会变的
		micro.WrapHandler(s.wrapper.GetUserID), // 这个也是 顺序不能变
		micro.Address(fmt.Sprintf(":%d", s.port)),
	)

	s.srv.Init()

	return video.RegisterVideoServiceHandler(s.srv.Server(), s.video)
}

func (s *Server) Run() error {
	return s.srv.Run()
}
