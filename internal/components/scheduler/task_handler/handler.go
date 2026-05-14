package task_handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"stream_hub/internal/infra"
	"stream_hub/pkg/mq"
	"stream_hub/pkg/email"
	"stream_hub/pkg/model/config"
	"time"

	"github.com/IBM/sarama"
)

type TaskHandler struct {
	client       *http.Client
	email        *email.Client
	imProducer   sarama.AsyncProducer
	*infra.Base
	taskSender   *infra.TaskSender // 用来做DAG 任务依赖链
	mediaPrefix  string
	auditServer  string
}

func NewTaskHandler(conf *config.CommonConfig, schedulerConf *config.SchedulerConfig, base *infra.Base) (*TaskHandler, error) {
	email := email.NewClient(conf)
	taskSender, err := infra.NewTaskSender(conf)
	if err != nil {
		return nil, err
	}
	imProducerClient, err := mq.NewKafkaProducer(conf)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case err := <-imProducerClient.Producer().Errors():
				log.Println("scheduler kafka producer error:", err)
			case <-imProducerClient.Producer().Successes():
			}
		}
	}()
	return &TaskHandler{
		client:      &http.Client{
			Timeout: 5 * time.Minute,
		},
		email:       email,
		imProducer:  imProducerClient.Producer(),
		Base:        base,
		taskSender:  taskSender,
		mediaPrefix: schedulerConf.Audit.MediaPrefix,
		auditServer: schedulerConf.Audit.ServerAddr,
	}, nil
}

func (t *TaskHandler) Get(url string) ([]byte, error) {
	resp, err := t.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (t *TaskHandler) PostJSON(url string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := t.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
