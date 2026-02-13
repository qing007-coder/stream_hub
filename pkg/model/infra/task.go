package infra

import "encoding/json"

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
