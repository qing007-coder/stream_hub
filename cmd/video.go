package main

import (
	"fmt"
	"stream_hub/internal/infra"
	"stream_hub/internal/video"
	"stream_hub/pkg/config"
)

func main() {
	commonConf, err := config.NewCommonConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	videoConf, err := config.NewVideoConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	base, err := infra.NewBase(commonConf)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	server, err := video.NewServer(base, commonConf, videoConf)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	if err := server.Run(); err != nil {
		fmt.Println("err:", err)
		return
	}
}
