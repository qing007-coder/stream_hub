package interaction

import (
	"encoding/json"

	"stream_hub/pkg/constant"
	infra_model "stream_hub/pkg/model/infra"
	"stream_hub/pkg/model/storage"
)

type notifyTaskPayload struct {
	ReceiverID  string `json:"receiver_id"`
	SenderID    string `json:"sender_id"`
	Content     string `json:"content"`
	ContentType int8   `json:"content_type"`
}

func (l *Like) notifyVideoAuthor(videoID, actorID, content string) {
	var video storage.VideoModel
	if err := l.DB.Select("author_id").Where("id = ?", videoID).First(&video).Error; err != nil {
		return
	}
	if video.AuthorID == "" || video.AuthorID == actorID {
		return
	}

	data, err := json.Marshal(notifyTaskPayload{
		ReceiverID:  video.AuthorID,
		SenderID:    actorID,
		Content:     content,
		ContentType: constant.IMContentTypeSystem,
	})
	if err != nil {
		return
	}

	_ = l.TaskSender.SendTask(infra_model.TaskMessage{
		Type:     constant.TaskSendNotify,
		BizID:    videoID,
		Priority: "normal",
		Payload: infra_model.TaskPayload{
			Operator: actorID,
			Action:   "notify_like",
			Source:   constant.Interaction,
			Data:     data,
		},
	})
}

func (f *Favourite) notifyVideoAuthor(videoID, actorID, content string) {
	var video storage.VideoModel
	if err := f.DB.Select("author_id").Where("id = ?", videoID).First(&video).Error; err != nil {
		return
	}
	if video.AuthorID == "" || video.AuthorID == actorID {
		return
	}

	data, err := json.Marshal(notifyTaskPayload{
		ReceiverID:  video.AuthorID,
		SenderID:    actorID,
		Content:     content,
		ContentType: constant.IMContentTypeSystem,
	})
	if err != nil {
		return
	}

	_ = f.TaskSender.SendTask(infra_model.TaskMessage{
		Type:     constant.TaskSendNotify,
		BizID:    videoID,
		Priority: "normal",
		Payload: infra_model.TaskPayload{
			Operator: actorID,
			Action:   "notify_favorite",
			Source:   constant.Interaction,
			Data:     data,
		},
	})
}

func (c *Comment) notifyVideoAuthor(videoID, actorID, content string) {
	var video storage.VideoModel
	if err := c.DB.Select("author_id").Where("id = ?", videoID).First(&video).Error; err != nil {
		return
	}
	if video.AuthorID == "" || video.AuthorID == actorID {
		return
	}

	data, err := json.Marshal(notifyTaskPayload{
		ReceiverID:  video.AuthorID,
		SenderID:    actorID,
		Content:     content,
		ContentType: constant.IMContentTypeSystem,
	})
	if err != nil {
		return
	}

	_ = c.TaskSender.SendTask(infra_model.TaskMessage{
		Type:     constant.TaskSendNotify,
		BizID:    videoID,
		Priority: "normal",
		Payload: infra_model.TaskPayload{
			Operator: actorID,
			Action:   "notify_comment",
			Source:   constant.Interaction,
			Data:     data,
		},
	})
}
