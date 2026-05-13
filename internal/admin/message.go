package admin

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/IBM/sarama"

	"stream_hub/internal/imnotify"
	"stream_hub/internal/infra"
	"stream_hub/pkg/constant"
	api_model "stream_hub/pkg/model/api"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"
)

type MessageController struct {
	*infra.Base
	imProducer sarama.AsyncProducer
}

func NewMessageController(base *infra.Base, producer sarama.AsyncProducer) *MessageController {
	return &MessageController{
		Base:       base,
		imProducer: producer,
	}
}

type SendSystemMessageReq struct {
	TargetType string `json:"target_type" binding:"required,oneof=all user"`
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content" binding:"required"`
}

func (m *MessageController) GetSystemMessages(ctx *gin.Context) {
	var req api_model.GetUserListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(ctx, "invalid param")
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	userID := strings.TrimSpace(ctx.Query("user_id"))
	content := strings.TrimSpace(ctx.Query("content"))
	offset := (req.Page - 1) * req.Size

	query := m.DB.Model(&storage.MessageModel{}).Where("sender_id = ?", constant.IMSystemSenderID)
	if userID != "" {
		query = query.Where("receiver_id = ?", userID)
	}
	if content != "" {
		query = query.Where("content LIKE ?", "%"+content+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		utils.BadRequest(ctx, fmt.Sprintf("count error: %s", err.Error()))
		return
	}

	var messages []storage.MessageModel
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.Size).Find(&messages).Error; err != nil {
		utils.BadRequest(ctx, fmt.Sprintf("query error: %s", err.Error()))
		return
	}

	list := make([]gin.H, 0, len(messages))
	for _, msg := range messages {
		list = append(list, gin.H{
			"id":           msg.ID,
			"sender_id":    msg.SenderID,
			"receiver_id":  msg.ReceiverID,
			"content":      msg.Content,
			"content_type": msg.ContentType,
			"status":       msg.Status,
			"created_at":   msg.CreatedAt,
		})
	}

	utils.StatusOK(ctx, gin.H{
		"total": total,
		"list":  list,
	}, "query successfully")
}

func (m *MessageController) SendSystemMessage(ctx *gin.Context) {
	var req SendSystemMessageReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, "invalid request body")
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	req.ReceiverID = strings.TrimSpace(req.ReceiverID)
	if req.Content == "" {
		utils.BadRequest(ctx, "content is required")
		return
	}
	if req.TargetType == "user" && req.ReceiverID == "" {
		utils.BadRequest(ctx, "receiver_id is required")
		return
	}

	receiverIDs, err := m.resolveMessageReceivers(req.TargetType, req.ReceiverID)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	for _, receiverID := range receiverIDs {
		// Send immediately via IM flow (persist + Kafka) for both single and all-user targets.
		err := imnotify.SendSystemToUser(context.Background(), m.DB, m.imProducer, receiverID, req.Content)
		if err != nil {
			utils.BadRequest(ctx, fmt.Sprintf("send message failed: %s", err.Error()))
			return
		}
	}

	utils.StatusOK(ctx, gin.H{
		"target_type": req.TargetType,
		"sent_count":  len(receiverIDs),
	}, "send successfully")
}

func (m *MessageController) resolveMessageReceivers(targetType, receiverID string) ([]string, error) {
	if targetType == "user" {
		var count int64
		if err := m.DB.Model(&storage.User{}).Where("id = ?", receiverID).Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, fmt.Errorf("user not found")
		}
		return []string{receiverID}, nil
	}

	var users []storage.User
	if err := m.DB.Select("id").Find(&users).Error; err != nil {
		return nil, err
	}

	receiverIDs := make([]string, 0, len(users))
	for _, user := range users {
		receiverIDs = append(receiverIDs, user.ID)
	}
	return receiverIDs, nil
}
