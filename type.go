package glog

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type OutputFormat uint8

func (of OutputFormat) String() string {
	switch of {
	case OutputFormatJSON:
		return "json"
	case OutputFormatTEXT:
		return "text"
	default:
		return "json"
	}
}

func (of OutputFormat) toDigit(v string) (OutputFormat, error) {
	switch v {
	case "json":
		return OutputFormatJSON, nil
	case "text":
		return OutputFormatTEXT, nil
	default:
		return 0, fmt.Errorf("unknown output format '%s'", v)
	}
}

const (
	OutputFormatJSON OutputFormat = iota
	OutputFormatTEXT
)

func ParseOutputFormat(format string) (OutputFormat, error) {
	var f OutputFormat
	return f.toDigit(strings.ToLower(format))
}

const (
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
	LevelDebug = slog.LevelDebug

	TimeKey    = slog.TimeKey
	LevelKey   = slog.LevelKey
	SourceKey  = slog.SourceKey
	MessageKey = slog.MessageKey

	KindTime     = slog.KindTime
	KindGroup    = slog.KindGroup
	KindString   = slog.KindString
	KindInt64    = slog.KindInt64
	KindUint64   = slog.KindUint64
	KindFloat64  = slog.KindFloat64
	KindBool     = slog.KindBool
	KindDuration = slog.KindDuration
	KindAny      = slog.KindAny
)

type (
	Logger         = slog.Logger
	Attr           = slog.Attr
	Level          = slog.Level
	Leveler        = slog.Leveler
	Handler        = slog.Handler
	Record         = slog.Record
	Value          = slog.Value
	HandlerOptions = slog.HandlerOptions
	LogValuer      = slog.LogValuer
	Source         = slog.Source
)

var (
	NewTextHandler = slog.NewTextHandler
	NewJSONHandler = slog.NewJSONHandler
	New            = slog.New
	SetDefault     = slog.SetDefault
	GetDefault     = slog.Default

	StringAttr   = slog.String
	BoolAttr     = slog.Bool
	Float64Attr  = slog.Float64
	AnyAttr      = slog.Any
	DurationAttr = slog.Duration
	IntAttr      = slog.Int
	Int64Attr    = slog.Int64
	Uint64Attr   = slog.Uint64
	Time         = slog.Time
	Any          = slog.Any
	String       = slog.String

	GroupValue  = slog.GroupValue
	StringValue = slog.StringValue
	Group       = slog.Group
)

func Float32Attr(key string, val float32) Attr {
	return slog.Float64(key, float64(val))
}

func UInt32Attr(key string, val uint32) Attr {
	return slog.Int(key, int(val))
}

func Int32Attr(key string, val int32) Attr {
	return slog.Int(key, int(val))
}

func TimeAttr(key string, time time.Time) Attr {
	return slog.String(key, time.String())
}

func ErrAttr(err error) Attr {
	return slog.String("error", err.Error())
}

var Err = ErrAttr
