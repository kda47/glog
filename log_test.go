package glog

import (
	"context"
	"log/slog"
	"os"
	"regexp"
	"testing"
)

func TestNewLogger(t *testing.T) {
	ctx := context.Background()

	logger := NewLogger()

	if logger == nil {
		t.Errorf("logger should not be nil")
	}

	if Default() != logger {
		t.Errorf("logger should be default logger")
	}

	if slog.Default() != Default() {
		t.Errorf("logger should be default logger")
	}

	if logger != Default() {
		t.Errorf("logger should be default logger")
	}

	logger = NewLogger(WithOutputFormat(OutputFormatJSON))
	if slog.Default() != logger {
		t.Errorf("logger should be default logger")
	}

	_ = NewLogger(WithOutputFormat(OutputFormatTEXT))
	if slog.Default() == logger {
		t.Errorf("logger should NOT be default logger")
	}

	logger = NewLogger(WithLevel("info"))
	if logger.Handler().Enabled(ctx, LevelDebug) {
		t.Errorf("logger should NOT be enabled for debug level")
	}

	logger = NewLogger(WithLevel("warn"))
	if logger.Handler().Enabled(ctx, LevelDebug) {
		t.Errorf("logger should NOT be enabled for debug level")
	}

	logger = NewLogger(WithLevel("debug"))
	enabled := []Level{LevelDebug, LevelInfo, LevelWarn, LevelError}
	for _, level := range enabled {
		if !logger.Handler().Enabled(ctx, level) {
			t.Errorf("logger should be enabled for all levels")
		}
	}

	logger = NewLogger(WithLevel("abcdef"))
	if !logger.Handler().Enabled(ctx, LevelInfo) {
		t.Errorf("logger should be enabled for info level")
	}

	if logger.Handler().Enabled(ctx, LevelDebug) {
		t.Errorf("logger should NOT be enabled for info level")
	}

	logger = NewLogger(WithAddSource(true))
	if slog.Default() != logger {
		t.Errorf("logger should be default logger")
	}

	logger = NewLogger(WithSetDefault(false))
	if slog.Default() == logger {
		t.Errorf("logger should NOT be default logger")
	}

	if logger == Default() {
		t.Errorf("logger should NOT be default logger")
	}

	if L(ctx) == logger {
		t.Errorf("logger should NOT be from context logger")
	}

	logger = NewLogger(WithAddSource(false), WithOutputFormat(OutputFormatJSON))
	ctx = ContextWithLogger(ctx, logger)
	WithAttrs(ctx, String("attr1", "abc"))

	if L(ctx) != logger {
		t.Errorf("logger should be from context logger")
	}

	h := NewDiscardHandler()
	logger = NewLogger(WithCustomHandler(h))
	WithDefaultAttrs(logger, String("attr1", "abc"))

	if logger.Handler() != h {
		t.Errorf("logger should be have DiscardHandler")
	}

	logger = NewLogger(WithCustomHandler(h))
	WithDefaultAttrs(logger, String("attr1", "abc"))

	if logger.Handler() != h {
		t.Errorf("logger should be have DiscardHandler")
	}

	file, err := os.CreateTemp("", "log")
	if err != nil {
		t.Errorf("on create tempfile, expected no error, got %s", err.Error())
	}
	defer os.Remove(file.Name())
	logger = NewLogger(WithOutputFilePath(file.Name()), WithAddSource(false))
	logger.Info("test")

	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Errorf("on read tempfile data error: %s", err.Error())
	}

	matched, _ := regexp.MatchString(`.*"level":"INFO","msg":"test"}`, string(data))
	if !matched {
		t.Errorf("wrong logs output data")
	}

	file1, err := os.CreateTemp("", "log")
	if err != nil {
		t.Errorf("on create tempfile, expected no error, got %s", err.Error())
	}
	defer os.Remove(file1.Name())

	err = os.Chmod(file1.Name(), 0444)
	if err != nil {
		t.Errorf("cant change temp file permissions")
	}

	fatal := false
	logFatalf = func(_ string, _ ...any) {
		fatal = true
	}
	logger = NewLogger(WithOutputFilePath(file1.Name()), WithAddSource(false))
	if fatal == false {
		t.Errorf("expected fatal exit")
	}
}
