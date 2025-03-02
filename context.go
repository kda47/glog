package glog

import (
	"context"
)

type contextLoggerKey struct{}

// ContextWithLogger puts logger to context
func ContextWithLogger(ctx context.Context, logger *Logger) context.Context {
	if logger == nil {
		return ctx
	}
	return context.WithValue(ctx, contextLoggerKey{}, logger)
}

// LoggerFromContext returns logger from context
func LoggerFromContext(ctx context.Context) *Logger {
	if ctx == nil {
		return GetDefault()
	}
	if logger, ok := ctx.Value(contextLoggerKey{}).(*Logger); ok && logger != nil {
		return logger
	}

	return GetDefault()
}
