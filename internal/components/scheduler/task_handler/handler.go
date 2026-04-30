package task_handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"stream_hub/internal/infra"
	"stream_hub/pkg/email"
	"stream_hub/pkg/model/config"
	"time"
)

type TaskHandler struct {
	client *http.Client
	email  *email.Client
	*infra.Base
	taskSender *infra.TaskSender // 用来做DAG 任务依赖链
}

func NewTaskHandler(conf *config.CommonConfig, base *infra.Base) (*TaskHandler, error) {
	email := email.NewClient(conf)
	taskSender, err := infra.NewTaskSender(conf)
	if err != nil {
		return nil, err
	}
	return &TaskHandler{
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
		email: email,
		Base:  base,
		taskSender: taskSender,
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
