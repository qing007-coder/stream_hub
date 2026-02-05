package infra

type TaskMessage struct {
	TaskID    string `json:"task_id"`
	Type      string `json:"type"`
	BizID     string `json:"biz_id"`
	Payload   string `json:"payload"`
	NextRunAt int64  `json:"next_run_at"` // unix timestamp
}
