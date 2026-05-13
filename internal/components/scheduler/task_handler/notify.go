package task_handler

import (
	"context"

	"stream_hub/internal/imnotify"
)

func (t *TaskHandler) SendNotify(ctx context.Context, req *imnotify.Request) error {
	return imnotify.Send(ctx, t.DB, t.imProducer, req)
}

// NotifyPayload is an alias for unmarshalling JSON notify payloads (same fields as imnotify.Request).
type NotifyPayload = imnotify.Request
