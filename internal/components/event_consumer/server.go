package event_consumer

import (
	"context"
	"encoding/json"
	"log"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/mq"

	"github.com/IBM/sarama"
)

type Server struct {
	consumer sarama.ConsumerGroup
	workerNum int
	eventChan chan <- []storage.Event
	Workers    []*StorageWorker
}

func NewServer(commonConf *config.CommonConfig, eventConsumerConf *config.EventConsumerConfig) (*Server, error) {
	consumer, err := mq.NewKafkaConsumerGroup(commonConf)
	if err != nil {
		return nil, err
	}

	eventChan := make(chan []storage.Event,  1000)
	workers := make([]*StorageWorker, 0)
	for _ = range eventConsumerConf.WorkerNum {
		worker, err := NewStorageWorker(eventChan, commonConf)
		if err != nil {
			return nil, err
		}

		workers = append(workers, worker)
	}

	return &Server{
		consumer: consumer.Consumer(),
		workerNum: eventConsumerConf.WorkerNum,
		eventChan: eventChan,
		Workers: workers,
	}, nil
}

func (s *Server) Start() {
	ctx := context.Background()
	for _, worker := range s.Workers {
		worker.Start()
	}

	for {
		if err := s.consumer.Consume(ctx, []string{constant.EventTopic}, s); err != nil {
			log.Printf("consume error: %v\n", err)
		}
		
		if ctx.Err() != nil {
			return 
		}
	}
}

func (s *Server) HandleMessage(msg *sarama.ConsumerMessage) error {
	var events []storage.Event
	if err := json.Unmarshal(msg.Value, &events); err != nil {
		return err
	}

	s.eventChan <- events
	
	return nil
}

func (s *Server) Setup(sarama.ConsumerGroupSession) error { return nil }

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (s *Server) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (s *Server) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if err := s.HandleMessage(msg); err != nil {
			log.Println("err:", err)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}