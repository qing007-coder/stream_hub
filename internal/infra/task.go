package infra

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"log"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/errors"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/model/infra"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/mq"
)

type TaskSender struct {
	producer      sarama.AsyncProducer
	db            *DB
	maxRetryCount int
}

func NewTaskSender(conf *config.CommonConfig) (*TaskSender, error) {
	producer, err := mq.NewKafkaProducer(conf)
	if err != nil {
		return nil, err
	}
	db, err := NewMysql(conf)
	if err != nil {
		return nil, err
	}

	taskSender := new(TaskSender)
	taskSender.producer = producer.Producer()
	taskSender.db = db

	go taskSender.Listen()

	return taskSender, nil
}

func (t *TaskSender) SendTask(message infra.TaskMessage) error {
	data, err := json.Marshal(&message)
	if err != nil {
		return err
	}

	task := storage.Task{
		Type:   message.Type,
		BizID:  message.BizID,
		Status: constant.TaskPending,
		//Payload: message.Payload,
	}

	t.db.Create(&task)
	message.TaskID = task.ID
	t.producer.Input() <- &sarama.ProducerMessage{
		Topic: constant.TaskTopic,
		Key:   sarama.StringEncoder(message.Type),
		Value: sarama.ByteEncoder(data),
	}

	return nil
}

func (t *TaskSender) RetrySendTask(data []byte) error {
	var task storage.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return err
	}

	t.db.Where("id = ?", task.ID).First(&task)
	if task.RetryCount > t.maxRetryCount {
		return errors.MaxRetryCount
	}

	t.db.Model(&storage.Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"retry_count": gorm.Expr("retry_count + 1"),
	})

	t.producer.Input() <- &sarama.ProducerMessage{
		Topic: constant.TaskTopic,
		Key:   sarama.StringEncoder(task.Type),
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
			data, err_ := err.Msg.Value.Encode()
			if err_ != nil {
				log.Println("error:", err_)
			}
			if err := t.RetrySendTask(data); err != nil {
				log.Println("error:", err)
			}
		}
	}
}
