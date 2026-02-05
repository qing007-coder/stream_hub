package mq

import (
	"fmt"
	"github.com/IBM/sarama"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
)

type KafkaConsumer struct {
	consumerGroup sarama.ConsumerGroup
}

func NewKafkaConsumerGroup(conf *config.CommonConfig) (*KafkaConsumer, error) {
	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	consumer, err := sarama.NewConsumerGroup([]string{fmt.Sprintf("%s:%s", conf.Kafka.Addr, conf.Kafka.Port)}, constant.ConsumerGroupID, cfg)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{consumerGroup: consumer}, nil
}

func (kc *KafkaConsumer) Consumer() sarama.ConsumerGroup {
	return kc.consumerGroup
}
