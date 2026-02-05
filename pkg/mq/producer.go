package mq

import (
	"fmt"
	"github.com/IBM/sarama"
	"stream_hub/pkg/model/config"
	"time"
)

type KafkaProducer struct {
	producer sarama.AsyncProducer
}

func NewKafkaProducer(config *config.CommonConfig) (*KafkaProducer, error) {
	conf := sarama.NewConfig()
	conf.Producer.Return.Successes = true
	conf.Producer.Return.Errors = true
	conf.Producer.RequiredAcks = sarama.WaitForLocal
	conf.Producer.Flush.Messages = 500                     // 队列满 500 条再发
	conf.Producer.Flush.Frequency = 100 * time.Millisecond // 或者每 100ms 发一次
	conf.Producer.MaxMessageBytes = 1000000                // 单条消息最大字节数
	producer, err := sarama.NewAsyncProducer([]string{fmt.Sprintf("%s:%s", config.Kafka.Addr, config.Kafka.Port)}, conf)
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{
		producer: producer,
	}, nil
}

func (k *KafkaProducer) Producer() sarama.AsyncProducer {
	return k.producer
}
