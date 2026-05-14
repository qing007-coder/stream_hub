package mq

import (
	"fmt"

	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	consumerGroup sarama.ConsumerGroup
}

func NewKafkaConsumerGroup(conf *config.CommonConfig) (*KafkaConsumer, error) {
	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	if conf.Kafka.Username != "" && conf.Kafka.Password != "" {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		cfg.Net.SASL.User = conf.Kafka.Username
		cfg.Net.SASL.Password = conf.Kafka.Password
	}

	consumer, err := sarama.NewConsumerGroup([]string{fmt.Sprintf("%s:%s", conf.Kafka.Addr, conf.Kafka.Port)}, constant.ConsumerGroupID, cfg)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{consumerGroup: consumer}, nil
}

func (kc *KafkaConsumer) Consumer() sarama.ConsumerGroup {
	return kc.consumerGroup
}
