package gateway

import (
	"fmt"
	grpcc "github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/consul"
	grpcs "github.com/go-micro/plugins/v4/server/grpc"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/registry"
	"stream_hub/pkg/model/config"
)

type Service struct {
	srv micro.Service
}

func NewService(conf *config.CommonConfig) *Service {
	srv := new(Service)
	consulRegister := consul.NewRegistry(
		registry.Addrs(fmt.Sprintf("%s:%s", conf.Consul.Addr, conf.Consul.Port)),
	)
	srv.srv = micro.NewService(
		micro.Server(grpcs.NewServer()), // 使用 gRPC server
		micro.Client(grpcc.NewClient()), // 使用 gRPC client
		micro.Registry(consulRegister),
	)

	return srv
}

func (s *Service) Client() client.Client {
	return s.srv.Client()
}
