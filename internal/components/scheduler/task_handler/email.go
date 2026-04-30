package task_handler

import (
	"context"
	infra_ "stream_hub/pkg/model/infra"
	"time"
)

func (t *TaskHandler) EmailHandler(ctx context.Context, task *infra_.TaskMessage) error {
	code, err := t.email.SendVerificationCode(task.BizID)
	if err != nil {
		return err
	}

	if err := t.Redis.Set(ctx, task.BizID, code, time.Minute*10); err != nil {
		return err
	}
	if err := t.Redis.Set(ctx, task.BizID+".send", 1, time.Minute); err != nil {
		return err
	}

	return nil
}
