package main

import (
	"fmt"
	"stream_hub/internal/components/logger"
	"stream_hub/pkg/config"
)

func main() {
	commnoConf, err := config.NewCommonConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	loggerConf, err := config.NewLoggerConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	server, err := logger.NewServer(commnoConf, loggerConf)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	server.Start()
}
