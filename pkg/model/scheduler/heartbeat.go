package scheduler

type HeartbeatPayload struct {
	NodeID    string   `json:"node_id"`
	WorkersID []string `json:"workers_id"`
}
