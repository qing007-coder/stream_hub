package core

import (
	"context"
	infra_ "stream_hub/pkg/model/infra"
	"sync"
)

type ServeMux struct {
	mu  sync.RWMutex
	mux map[string]func(context.Context, *infra_.TaskMessage) error
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		mux: make(map[string]func(context.Context, *infra_.TaskMessage) error),
	}
}

func (s *ServeMux) HandleFunc(pattern string, handler func(context.Context, *infra_.TaskMessage) error) {
	s.mu.Lock()
	s.mux[pattern] = handler

	s.mu.Unlock()
}

func (s *ServeMux) Execute(ctx context.Context, pattern string, task *infra_.TaskMessage) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	handler := s.mux[pattern]
	return handler(ctx, task)
}
