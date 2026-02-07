package interaction

import (
	"context"
	"errors"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"
)

type Wrapper struct {
}

func NewWrapper() *Wrapper {
	c := new(Wrapper)
	return c
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
