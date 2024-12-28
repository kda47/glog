package glog

import (
	"context"
	"testing"
)

func TestLoggerInContext(t *testing.T) {
	logger := NewDiscardLogger()
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
