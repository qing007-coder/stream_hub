package main

import (
	"fmt"
	"stream_hub/internal/infra"
	"stream_hub/internal/media"
	"stream_hub/internal/security"
	"stream_hub/pkg/config"
)

func main() {
	commonConf, err := config.NewCommonConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	mediaConf, err := config.NewMediaConfig()
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

	router := media.NewMediaRouter(base, mediaConf, auth)
	if err := router.Run(); err != nil {
		fmt.Println("err:", err)
		return
	}
}
