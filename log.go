package glog

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

const (
	defaultLevel        = LevelInfo
	defaultAddSource    = true
	defaultOutputFormat = OutputFormatJSON
	defaultSetDefault   = true
	defaultLogFile      = ""
)

func NewLogger(opts ...LoggerOption) *Logger {
	config := &LoggerOptions{
		Level:         defaultLevel,
		AddSource:     defaultAddSource,
		OutputFormat:  defaultOutputFormat,
		SetDefault:    defaultSetDefault,
		LogFilePath:   defaultLogFile,
		CustomHandler: nil,
	}

	for _, opt := range opts {
		opt(config)
	}

	var logger *Logger

	if config.CustomHandler != nil {
		logger = New(config.CustomHandler)
	} else {
		options := &HandlerOptions{
			AddSource: config.AddSource,
			Level:     config.Level,
		}

		var logStream io.Writer = os.Stdout
		var isatty bool

		switch config.LogFilePath {
		case "", "stdout", "/dev/stdout":
			logStream = os.Stdout
			isatty = true
		default:
			f, err := os.OpenFile(config.LogFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				log.Fatalf("Error on opening logging file: %s\n", err.Error())
			}
			logStream = f
		}

		var handler slog.Handler

		switch config.OutputFormat {
		case OutputFormatJSON:
			handler = NewJSONHandler(logStream, options)
		case OutputFormatTEXT:
			opts := &tint.Options{
				Level:      options.Level,
				TimeFormat: time.DateTime,
				NoColor:    !isatty,
				AddSource:  options.AddSource,
			}
			handler = tint.NewHandler(logStream, opts)
		}

		logger = New(handler)
	}

	if config.SetDefault {
		SetDefault(logger)
	}

	return logger
}

type LoggerOptions struct {
	Level         Level
	AddSource     bool
	OutputFormat  OutputFormat
	SetDefault    bool
	LogFilePath   string
	CustomHandler Handler
}

type LoggerOption func(*LoggerOptions)

// WithLevel logger option sets the log level, if not set, the default level is Info
func WithLevel(level string) LoggerOption {
	return func(o *LoggerOptions) {
		var l Level
		if err := l.UnmarshalText([]byte(level)); err != nil {
			l = LevelInfo
		}

		o.Level = l
	}
}

// WithAddSource logger option sets the add source option, which will add source file and line number to the log record
func WithAddSource(addSource bool) LoggerOption {
	return func(o *LoggerOptions) {
		o.AddSource = addSource
	}
}

// WithOutputFormat set output format
func WithOutputFormat(format OutputFormat) LoggerOption {
	return func(o *LoggerOptions) {
		o.OutputFormat = format
	}
}

// WithSetDefault logger option sets the set default option, which will set the created logger as default logger
func WithSetDefault(setDefault bool) LoggerOption {
	return func(o *LoggerOptions) {
		o.SetDefault = setDefault
	}
}

// WithCustomHandler logger option sets the custom handler
func WithCustomHandler(handler Handler) LoggerOption {
	return func(o *LoggerOptions) {
		o.CustomHandler = handler
	}
}

// WithAttrs returns logger with attributes
func WithAttrs(ctx context.Context, attrs ...Attr) *Logger {
	logger := L(ctx)
	for _, attr := range attrs {
		logger = logger.With(attr)
	}

	return logger
}

// WithDefaultAttrs returns logger with default attributes
func WithDefaultAttrs(logger *Logger, attrs ...Attr) *Logger {
	for _, attr := range attrs {
		logger = logger.With(attr)
	}

	return logger
}

func L(ctx context.Context) *Logger {
	return LoggerFromContext(ctx)
}

// WithName returns logger with name attribute
func WithName(ctx context.Context, name string) *Logger {
	return L(ctx).With(StringAttr(NameKey, name))
}

func Default() *Logger {
	return GetDefault()
}
