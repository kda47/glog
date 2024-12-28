package glog

import (
	"context"
	"testing"
)

func TestDiscardLogger(t *testing.T) {
	ctx := context.Background()
	logger := NewDiscardLogger()
	handler := logger.Handler()

	for _, lvl := range []Level{LevelDebug, LevelError, LevelWarn, LevelInfo} {
		if handler.Enabled(ctx, lvl) {
			t.Errorf("Enabled(ctx, '%s') was return true, but expected false", lvl.String())
		}
	}
	if handler.Handle(ctx, Record{}) != nil {
		t.Error("expected nil")
	}
	if handler.WithGroup("group1") != handler {
		t.Error("expected return handler object")
	}
	if handler.WithAttrs([]Attr{StringAttr("a", "b")}) != handler {
		t.Error("expected return handler object")
	}
}
