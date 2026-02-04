package logger

import (
	"context"
	"github.com/IBM/sarama"
	"log"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/mq"
)

type Server struct {
	consumer        sarama.ConsumerGroup
	workerNum       int
	userBatchChan   chan<- []byte
	systemBatchChan chan<- []byte
	Workers         []*StorageWorker
}

func NewServer(conf *config.CommonConfig, loggerConf *config.LoggerConfig) (*Server, error) {
	consumer, err := mq.NewKafkaConsumerGroup(conf)
	if err != nil {
		return nil, err
	}

	userBatchChan := make(chan []byte, 10)
	systemBatchChan := make(chan []byte, 10)

	workers := make([]*StorageWorker, 0)
	for i := 0; i < loggerConf.WorkerNum; i++ {
		worker, err := NewStorageWorker(conf, userBatchChan, systemBatchChan)
		if err != nil {
			return nil, err
		}

		workers = append(workers, worker)
	}

	return &Server{
		consumer:        consumer.Consumer(),
		workerNum:       loggerConf.WorkerNum,
		userBatchChan:   userBatchChan,
		systemBatchChan: systemBatchChan,
		Workers:         workers,
	}, nil
}

func (s *Server) Start() {
	ctx := context.Background()

	for _, worker := range s.Workers {
		worker.Start()
	}

	for {
		if err := s.consumer.Consume(ctx, []string{constant.UserLogTopic, constant.SystemLogTopic}, s); err != nil {
			log.Printf("consume error: %v", err)
		}

		if ctx.Err() != nil {
			return
		}
	}
}

func (s *Server) HandleMessage(msg *sarama.ConsumerMessage) {
	switch msg.Topic {
	case constant.UserLogTopic:
		s.userBatchChan <- msg.Value
	case constant.SystemLogTopic:
		s.systemBatchChan <- msg.Value
	}
}

func (s *Server) Setup(sarama.ConsumerGroupSession) error { return nil }

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (s *Server) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (s *Server) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		s.HandleMessage(msg)
		session.MarkMessage(msg, "")
	}
	return nil
}
