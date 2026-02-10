package video

import (
	"context"
	"errors"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/eventmap"

	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"
)

type Wrapper struct {
}

func NewWrapper() *Wrapper {
	return new(Wrapper)
}

func (w *Wrapper) GetUserID(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, resp interface{}) error {
		md, ok := metadata.FromContext(ctx)
		if !ok {
			return errors.New("need id")
		}

		uid, ok := md.Get("user_id")
		if !ok {
			return errors.New("need id")
		}

		newCtx := context.WithValue(ctx, "user_id", uid)

		err := fn(newCtx, req, resp)

		// 处理完之后的可以做的
		return err
	}
}

func (w *Wrapper) SendEventField(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, resp interface{}) error {
		endpointKey := req.Endpoint()
		eventType, exists := eventmap.GRPCEndpointToEvent[endpointKey]
		if !exists {
			return fn(ctx, req, resp)
		}
		// 设置事件类型到上下文
		ctx = context.WithValue(ctx, "event_type", eventType)
		// 设置资源类型到上下文
		ctx = context.WithValue(ctx, "resource_type", constant.ResourceVideo)

		return fn(ctx, req, resp)
	}
}
