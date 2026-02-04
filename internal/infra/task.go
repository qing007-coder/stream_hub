package infra

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"log"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/model/infra"
	"stream_hub/pkg/mq"
)

type TaskSender struct {
	producer sarama.AsyncProducer
}

func NewTaskSender(conf *config.CommonConfig) (*TaskSender, error) {
	producer, err := mq.NewKafkaProducer(conf)
	if err != nil {
		return nil, err
	}

	taskSender := new(TaskSender)
	taskSender.producer = producer.Producer()

	go taskSender.Listen()

	return taskSender, nil
}

func (t *TaskSender) SendTask(message infra.TaskMessage) error {
	data, err := json.Marshal(&message)
	if err != nil {
		return err
	}

	t.producer.Input() <- &sarama.ProducerMessage{
		Topic: constant.TaskTopic,
		Value: sarama.ByteEncoder(data),
	}

	return nil
}

func (t *TaskSender) Listen() {
	for {
		select {
		case _ = <-t.producer.Successes():
		case err := <-t.producer.Errors():
			log.Println("error:", err)
		}
	}
}
