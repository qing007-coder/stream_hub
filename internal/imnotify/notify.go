package imnotify

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"

	"stream_hub/internal/infra"
	"stream_hub/pkg/constant"
	im_api "stream_hub/pkg/model/api"
	"stream_hub/pkg/model/storage"
)

// Request matches JSON from admin tasks and interaction notify payloads.
type Request struct {
	ReceiverID  string `json:"receiver_id"`
	SenderID    string `json:"sender_id"`
	Content     string `json:"content"`
	ContentType int8   `json:"content_type"`
}

// Send persists an IM message, updates conversations, and publishes to Kafka for the IM WebSocket consumer.
func Send(ctx context.Context, db *infra.DB, producer sarama.AsyncProducer, req *Request) error {
	if req == nil || req.ReceiverID == "" || req.Content == "" {
		return nil
	}

	senderID := req.SenderID
	if senderID == "" {
		senderID = constant.IMSystemSenderID
	}

	contentType := req.ContentType
	if contentType == 0 {
		contentType = constant.IMContentTypeSystem
	}

	message := storage.MessageModel{
		SenderID:    senderID,
		ReceiverID:  req.ReceiverID,
		Content:     req.Content,
		ContentType: contentType,
		Status:      0,
	}
	if err := db.WithContext(ctx).Create(&message).Error; err != nil {
		return err
	}

	updateConversation(ctx, db, req.ReceiverID, senderID, message.ID, true)
	if senderID != constant.IMSystemSenderID {
		updateConversation(ctx, db, senderID, req.ReceiverID, message.ID, false)
	}

	push := im_api.IMPushMessage{
		MessageID:   message.ID,
		SenderID:    senderID,
		ReceiverID:  req.ReceiverID,
		Content:     req.Content,
		ContentType: contentType,
		CreatedAt:   message.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	data, err := json.Marshal(push)
	if err != nil {
		return err
	}
	producer.Input() <- &sarama.ProducerMessage{
		Topic: constant.IMMessageTopic,
		Value: sarama.ByteEncoder(data),
	}

	return nil
}

// SendSystemToUser sends a system notification (admin / shortcut).
func SendSystemToUser(ctx context.Context, db *infra.DB, producer sarama.AsyncProducer, receiverID, content string) error {
	return Send(ctx, db, producer, &Request{
		ReceiverID:  receiverID,
		SenderID:    constant.IMSystemSenderID,
		Content:     content,
		ContentType: constant.IMContentTypeSystem,
	})
}

func updateConversation(ctx context.Context, db *infra.DB, userID, targetUserID, messageID string, increaseUnread bool) {
	var conversation storage.ConversationModel
	err := db.WithContext(ctx).Where("user_id = ? AND target_user_id = ?", userID, targetUserID).First(&conversation).Error

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
		db.WithContext(ctx).Create(&conversation)
		return
	}

	unreadCount := conversation.UnreadCount
	if increaseUnread {
		unreadCount++
	}
	db.WithContext(ctx).Model(&conversation).Updates(map[string]interface{}{
		"last_message_id": messageID,
		"unread_count":    unreadCount,
		"updated_at":      time.Now(),
	})
}
