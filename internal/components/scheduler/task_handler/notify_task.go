package task_handler

import (
	"context"
	"encoding/json"

	infra_ "stream_hub/pkg/model/infra"
)

func (t *TaskHandler) NotifyTask(ctx context.Context, task *infra_.TaskMessage) error {
	var req NotifyPayload
	if len(task.Payload.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(task.Payload.Data, &req); err != nil {
		return err
	}
	return t.SendNotify(ctx, &req)
}
