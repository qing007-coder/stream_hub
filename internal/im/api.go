package im

import (
	"encoding/json"
	"stream_hub/internal/infra"
	"stream_hub/pkg/constant"
	im_api "stream_hub/pkg/model/api"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

type IMApi struct {
	*infra.Base
	producer sarama.AsyncProducer
}

func NewIMApi(base *infra.Base, producer sarama.AsyncProducer) *IMApi {
	return &IMApi{
		Base:     base,
		producer: producer,
	}
}

type SendMessageReq struct {
	ReceiverID  string `json:"receiver_id" binding:"required"`
	Content     string `json:"content" binding:"required"`
	ContentType int8   `json:"content_type" binding:"required"`
}

type GetMessagesReq struct {
	TargetUserID string `json:"target_user_id" binding:"required"`
	Page         int    `json:"page" binding:"required"`
	Size         int    `json:"size" binding:"required"`
}

type GetConversationsReq struct {
	Page int `json:"page" binding:"required"`
	Size int `json:"size" binding:"required"`
}

type MessageResp struct {
	ID          string `json:"id"`
	SenderID    string `json:"sender_id"`
	ReceiverID  string `json:"receiver_id"`
	Content     string `json:"content"`
	ContentType int8   `json:"content_type"`
	Status      int8   `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type ConversationResp struct {
	ID            string `json:"id"`
	TargetUserID  string `json:"target_user_id"`
	LastMessageID string `json:"last_message_id"`
	UnreadCount   int    `json:"unread_count"`
	UpdatedAt     string `json:"updated_at"`
}

func (api *IMApi) SendMessage(ctx *gin.Context) {
	var req SendMessageReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, utils.MessageInvalidBody)
		return
	}

	senderID := ctx.GetString("user_id")

	message := storage.MessageModel{
		SenderID:    senderID,
		ReceiverID:  req.ReceiverID,
		Content:     req.Content,
		ContentType: req.ContentType,
		Status:      0,
	}

	if err := api.DB.Create(&message).Error; err != nil {
		utils.BadRequest(ctx, "send message failed")
		return
	}

	api.updateConversation(senderID, req.ReceiverID, message.ID, false)
	api.updateConversation(req.ReceiverID, senderID, message.ID, true)

	kafkaMsg := im_api.IMPushMessage{
		MessageID:   message.ID,
		SenderID:    message.SenderID,
		ReceiverID:  message.ReceiverID,
		Content:     message.Content,
		ContentType: message.ContentType,
		CreatedAt:   message.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	kafkaData, err := json.Marshal(kafkaMsg)
	if err != nil {
		api.Logger.Error("marshal kafka message failed", "", "", "", "", "", constant.IM, 500, 0)
	} else {
		api.producer.Input() <- &sarama.ProducerMessage{
			Topic: constant.IMMessageTopic,
			Value: sarama.ByteEncoder(kafkaData),
		}
	}

	utils.StatusOK(ctx, MessageResp{
		ID:          message.ID,
		SenderID:    message.SenderID,
		ReceiverID:  message.ReceiverID,
		Content:     message.Content,
		ContentType: message.ContentType,
		Status:      message.Status,
		CreatedAt:   message.CreatedAt.Format("2006-01-02 15:04:05"),
	}, "send message successfully")
}

func (api *IMApi) GetMessages(ctx *gin.Context) {
	var req GetMessagesReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, utils.MessageInvalidBody)
		return
	}

	userID := ctx.GetString("user_id")

	offset := (req.Page - 1) * req.Size

	var messages []storage.MessageModel
	err := api.DB.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, req.TargetUserID, req.TargetUserID, userID,
	).Order("created_at DESC").Offset(offset).Limit(req.Size).Find(&messages).Error

	if err != nil {
		utils.BadRequest(ctx, "get messages failed")
		return
	}

	api.DB.Model(&storage.MessageModel{}).Where(
		"receiver_id = ? AND sender_id = ? AND status = 0",
		userID, req.TargetUserID,
	).Update("status", 1)

	api.DB.Model(&storage.ConversationModel{}).Where(
		"user_id = ? AND target_user_id = ?",
		userID, req.TargetUserID,
	).Update("unread_count", 0)

	var resp []MessageResp
	for _, msg := range messages {
		resp = append(resp, MessageResp{
			ID:          msg.ID,
			SenderID:    msg.SenderID,
			ReceiverID:  msg.ReceiverID,
			Content:     msg.Content,
			ContentType: msg.ContentType,
			Status:      msg.Status,
			CreatedAt:   msg.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	utils.StatusOK(ctx, resp, "get messages successfully")
}

func (api *IMApi) GetConversations(ctx *gin.Context) {
	var req GetConversationsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, utils.MessageInvalidBody)
		return
	}

	userID := ctx.GetString("user_id")

	offset := (req.Page - 1) * req.Size

	var conversations []storage.ConversationModel
	err := api.DB.Where("user_id = ?", userID).Order("updated_at DESC").Offset(offset).Limit(req.Size).Find(&conversations).Error

	if err != nil {
		utils.BadRequest(ctx, "get conversations failed")
		return
	}

	var resp []ConversationResp
	for _, conv := range conversations {
		resp = append(resp, ConversationResp{
			ID:            conv.ID,
			TargetUserID:  conv.TargetUserID,
			LastMessageID: conv.LastMessageID,
			UnreadCount:   conv.UnreadCount,
			UpdatedAt:     conv.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	utils.StatusOK(ctx, resp, "get conversations successfully")
}

func (api *IMApi) updateConversation(userID, targetUserID, messageID string, increaseUnread bool) {
	var conversation storage.ConversationModel
	err := api.DB.Where("user_id = ? AND target_user_id = ?", userID, targetUserID).First(&conversation).Error

	if err != nil {
		conversation = storage.ConversationModel{
			UserID:        userID,
			TargetUserID:  targetUserID,
			LastMessageID: messageID,
			UnreadCount:   0,
		}
		if increaseUnread {
			conversation.UnreadCount = 1
		}
		api.DB.Create(&conversation)
	} else {
		unreadCount := conversation.UnreadCount
		if increaseUnread {
			unreadCount++
		}
		api.DB.Model(&conversation).Updates(map[string]interface{}{
			"last_message_id": messageID,
			"unread_count":    unreadCount,
		})
	}
}
