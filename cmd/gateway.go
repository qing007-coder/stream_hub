package main

import (
	"fmt"
	"stream_hub/internal/gateway"
	"stream_hub/internal/infra"
	"stream_hub/internal/security"
	"stream_hub/pkg/config"
	"stream_hub/pkg/constant"
)

func main() {
	commonConf, err := config.NewCommonConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	gatewayConf, err := config.NewGatewayConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	base, err := infra.NewBase(commonConf)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	auth := security.NewAuth(commonConf)

	ratelimiter, err := infra.NewRatelimiter(base.Redis, constant.Gateway, commonConf)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	router := gateway.NewGatewayRouter(base, auth, ratelimiter, commonConf, gatewayConf)
	if err := router.Run(); err != nil {
		fmt.Println("err:", err)
		return
	}
}
