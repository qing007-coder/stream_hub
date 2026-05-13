package main

import (
	"fmt"
	"github.com/IBM/sarama"
	"stream_hub/internal/im"
	"stream_hub/internal/infra"
	"stream_hub/internal/security"
	"stream_hub/pkg/config"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/mq"
	"sync"
)

func main() {
	commonConf, err := config.NewCommonConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	imConf, err := config.NewIMConfig()
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

	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	consumer, err := sarama.NewConsumerGroup(
		[]string{fmt.Sprintf("%s:%s", commonConf.Kafka.Addr, commonConf.Kafka.Port)},
		constant.IMConsumerGroupID,
		cfg,
	)
	if err != nil {
		fmt.Println("Failed to create kafka consumer:", err)
		return
	}
	defer consumer.Close()

	auth := security.NewAuth(commonConf)
	router := im.NewIMRouter(base, auth, imConf, producer.Producer())

	go func() {
		im.StartKafkaConsumer(consumer, router.Hub())
	}()

	go func() {
		for {
			select {
			case err := <-producer.Producer().Errors():
				fmt.Println("Kafka producer error:", err)
			case <-producer.Producer().Successes():
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := router.Run(); err != nil {
			fmt.Println("err:", err)
		}
	}()

	wg.Wait()
}