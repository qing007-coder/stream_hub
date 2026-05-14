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
		TargetURL: fmt.Sprintf("%s%s", t.mediaPrefix, media.M3u8Url),
	}
	data, err := t.PostJSON(t.auditServer+"/audit", req)
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

	if err := t.DB.Model(&storage.MediaModel{}).Where("id = ?", task.BizID).Updates(map[string]interface{}{
		"audit_status": status,
		"metadata":     json.RawMessage(data),
	}).Error; err != nil {
		return err
	}

	var video storage.VideoModel
	if err := t.DB.Select("author_id", "title").Where("id = ?", media.VideoID).First(&video).Error; err == nil && video.AuthorID != "" {
		notifyContent := "你的视频《" + video.Title + "》机审未通过"
		if status == constant.VideoStatusMachinePassed {
			notifyContent = "你的视频《" + video.Title + "》机审通过"
		}
		_ = t.SendNotify(ctx, &NotifyPayload{
			ReceiverID:  video.AuthorID,
			SenderID:    constant.IMSystemSenderID,
			Content:     notifyContent,
			ContentType: constant.IMContentTypeSystem,
		})
	}

	return nil
}
