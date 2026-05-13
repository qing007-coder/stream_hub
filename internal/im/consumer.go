package im

import (
	"context"
	"encoding/json"
	"log"
	"stream_hub/pkg/constant"
	im_api "stream_hub/pkg/model/api"

	"github.com/IBM/sarama"
)

type Consumer struct {
	ready chan bool
	hub   *Hub
}

func NewConsumer(hub *Hub) *Consumer {
	return &Consumer{
		ready: make(chan bool),
		hub:   hub,
	}
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("message channel was closed")
				return nil
			}

			log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)

			var kafkaMsg im_api.IMPushMessage
			if err := json.Unmarshal(message.Value, &kafkaMsg); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			err := c.hub.SendMessageToUser(kafkaMsg.ReceiverID, kafkaMsg)
			if err != nil {
				log.Printf("Failed to send message to user %s: %v", kafkaMsg.ReceiverID, err)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func StartKafkaConsumer(consumer sarama.ConsumerGroup, hub *Hub) {
	ctx := context.Background()
	consumerInstance := NewConsumer(hub)

	topics := []string{constant.IMMessageTopic}

	for {
		if err := consumer.Consume(ctx, topics, consumerInstance); err != nil {
			log.Printf("Error from consumer: %v", err)
		}

		if ctx.Err() != nil {
			return
		}
	}
}
