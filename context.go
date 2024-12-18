package glog

import (
	"context"
)

type contextLoggerKey struct{}

// ContextWithLogger puts logger to context
func ContextWithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, contextLoggerKey{}, logger)
}

// LoggerFromContext returns logger from context
func LoggerFromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(contextLoggerKey{}).(*Logger); ok {
		return logger
	}

	return GetDefault()
}
