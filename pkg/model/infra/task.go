package infra

import (
	"encoding/json"
	"errors"
	"strconv"
)

type TaskMessage struct {
	TaskID     string      `json:"task_id"`
	Type       string      `json:"type"`
	BizID      string      `json:"biz_id"`
	Priority   string      `json:"priority"`
	Payload    TaskPayload `json:"payload"`
	RetryCount int         `json:"retry_count"`
}

type TaskPayload struct {
	Operator string          `json:"operator"`
	Source   string          `json:"source"`
	Action   string          `json:"action"`
	Data     json.RawMessage `json:"data"`
}

func (t *TaskMessage) TransformByMap(data map[string]string) error {
	if data["task_id"] == "" {
		return errors.New("missing task_id")
	}

	t.TaskID = data["task_id"]
	t.Type = data["type"]
	t.BizID = data["biz_id"]
	t.Priority = data["priority"]
	retryCount, err := strconv.Atoi(data["retry_count"])
	if err != nil {
		return errors.New("invalid retry_count: " + err.Error())
	}
	t.RetryCount = retryCount

	return nil
}
