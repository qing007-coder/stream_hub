package api

type IMPushMessage struct {
	MessageID   string `json:"message_id"`
	SenderID    string `json:"sender_id"`
	ReceiverID  string `json:"receiver_id"`
	Content     string `json:"content"`
	ContentType int8   `json:"content_type"`
	CreatedAt   string `json:"created_at"`
}
