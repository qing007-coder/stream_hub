package mq

import (
	"fmt"
	"time"

	"stream_hub/pkg/model/config"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	producer sarama.AsyncProducer
}

func NewKafkaProducer(config *config.CommonConfig) (*KafkaProducer, error) {
	conf := sarama.NewConfig()
	conf.Producer.Return.Successes = true
	conf.Producer.Return.Errors = true
	conf.Producer.RequiredAcks = sarama.WaitForLocal
	conf.Producer.Flush.Messages = 500
	conf.Producer.Flush.Frequency = 100 * time.Millisecond
	conf.Producer.MaxMessageBytes = 1000000

	if config.Kafka.Username != "" && config.Kafka.Password != "" {
		conf.Net.SASL.Enable = true
		conf.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		conf.Net.SASL.User = config.Kafka.Username
		conf.Net.SASL.Password = config.Kafka.Password
	}

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
