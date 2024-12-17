package glog

import (
	"context"
)

func NewDiscardLogger() *Logger {
	return New(NewDiscardHandler())
}

type DiscardHandler struct{}

func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}

func (h *DiscardHandler) Handle(_ context.Context, _ Record) error {
	return nil
}

func (h *DiscardHandler) WithAttrs(_ []Attr) Handler {
	return h
}

func (h *DiscardHandler) WithGroup(_ string) Handler {
	return h
}

func (h *DiscardHandler) Enabled(_ context.Context, _ Level) bool {
	return false
}
