package main

import (
	"fmt"
	"stream_hub/internal/infra"
	"stream_hub/internal/interaction"
	"stream_hub/pkg/config"
)

func main() {
	commonConf, err := config.NewCommonConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	interactionConf, err := config.NewInteractionConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	base, err := infra.NewBase(commonConf)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	server, err := interaction.NewServer(base, commonConf, interactionConf)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	if err := server.Run(); err != nil {
		fmt.Println("err:", err)
		return
	}
}
