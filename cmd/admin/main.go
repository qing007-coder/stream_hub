package main

import (
	"fmt"
	"stream_hub/internal/admin"
	"stream_hub/internal/infra"
	"stream_hub/internal/security"
	"stream_hub/pkg/config"
	"stream_hub/pkg/mq"
)

func main() {
	commonConf, err := config.NewCommonConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	adminConf, err := config.NewAdminConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	base, err := infra.NewBase(commonConf)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	producer, err := mq.NewKafkaProducer(commonConf)
	if err != nil {
		fmt.Println("Failed to create kafka producer:", err)
		return
	}
	defer producer.Producer().Close()
	go func() {
		for {
			select {
			case err := <-producer.Producer().Errors():
				fmt.Println("Kafka producer error:", err)
			case <-producer.Producer().Successes():
			}
		}
	}()

	auth := security.NewAuth(commonConf)
	
	userController := admin.NewUserController(base)
	if err := userController.InitRootAdmin(); err != nil {
		fmt.Println("init root admin failed:", err)
		return
	}

	router := admin.NewAdminRouter(base, auth, adminConf, producer.Producer())

	fmt.Printf("Admin service starting on port %d...\n", adminConf.Port)
	if err := router.Run(); err != nil {
		fmt.Println("err:", err)
	}
}
