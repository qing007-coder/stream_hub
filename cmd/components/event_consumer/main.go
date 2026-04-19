package main

import (
	"fmt"
	"stream_hub/internal/components/event_consumer"
	"stream_hub/pkg/config"
)

func main() {
	commnoConf, err := config.NewCommonConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	eventConsumerConfig, err := config.NewEventConsumerConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	server, err := event_consumer.NewServer(commnoConf, eventConsumerConfig)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	server.Start()
}
