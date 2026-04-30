package task_handler

import (
	"context"
	"encoding/json"
	"errors"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/api"
	infra_ "stream_hub/pkg/model/infra"
	"time"
)

func (t *TaskHandler) CalculateFeed(ctx context.Context, task *infra_.TaskMessage) error {
	var req api.CalculateFeedReq
	data, err := t.PostJSON("http://127.0.0.1:8088/calculate_feed", req)
	if err != nil {
		return err
	}

	var resp api.CalculateFeedResp 
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	if !resp.Status {
		return errors.New(resp.Message)
	}

	t.taskSender.SendDelayTask(infra_.TaskMessage{
		Type: constant.TaskCalculateFeed,
		BizID: "",
		Priority: "critical",
		Payload: infra_.TaskPayload{
			Operator: "",
			Action: "",
			Source: constant.Scheduler,
			Data: nil,
		},
	}, time.Hour * 2)

	return nil
}

