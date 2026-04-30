package task_handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/api"
	infra_ "stream_hub/pkg/model/infra"
	"stream_hub/pkg/model/storage"
)

func (t *TaskHandler) AuditVideo(ctx context.Context, task *infra_.TaskMessage) error {
	var media storage.MediaModel
	if err := t.DB.Where("id = ?", task.BizID).First(&media).Error; err != nil {
		return err
	}
	req := api.MachineAudit{
		TargetURL: fmt.Sprintf("http://192.168.233.128:9100%s", media.M3u8Url),
	}
	data, err := t.PostJSON("http://127.0.0.1:8088/audit", req)
	if err != nil {
		return err
	}

	var resp api.AuditResult
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	status := 0

	switch resp.Status {
	case constant.MachineAuditPass:
		status = constant.VideoStatusMachinePassed
	case constant.MachineAuditReject:
		status = constant.VideoStatusMachineFailed
	case constant.MachineAuditSuspicious:
		status = constant.VideoStatusManualChecking

	default:
		return errors.New("unknown status")
	}

	return t.DB.Model(&storage.VideoModel{}).Where("id = ?", task.BizID).Updates(map[string]interface{}{
		"audit_status": status,
		"video_meta":   resp,
	}).Error
}
