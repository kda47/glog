package glog

import (
	"context"
	"testing"
)

func TestLoggerInContext(t *testing.T) {
	logger := NewDiscardLogger()

	newCtx := ContextWithLogger(context.Background(), nil)

	if newCtx != context.Background() {
		t.Errorf("it was expected that the context would be returned unchanged")
	}

	if LoggerFromContext(nil) != Default() {
		t.Errorf("expected default logger")
	}

	ctxWithLogger := ContextWithLogger(context.Background(), logger)

	if LoggerFromContext(ctxWithLogger) != logger {
		t.Errorf("wrong logger from context, expected same logger that was putted to context")
	}

	loggerFromCtx := LoggerFromContext(context.Background())

	if loggerFromCtx == logger {
		t.Errorf("wrong logger from context, expected default logger")
	}

	if loggerFromCtx != Default() {
		t.Errorf("wrong logger from context, expected default logger")
	}

}
