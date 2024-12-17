package glog

import (
	"context"
	"runtime"
	"time"

	"github.com/docker/go-units"
)

func RunPeriodicMemoryLogging(ctx context.Context, logger *Logger, level Level, interval time.Duration, gcCall bool) {
	if logger == nil {
		logger = NewLogger((WithLevel(level.String())))
	}
	logger = logger.With(StringAttr("name", "memory_stat"))

	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if gcCall {
				runtime.GC()
			}
			MemoryStatisticLogging(logger, level)
		}
		timer.Reset(interval)
	}
}

func StartPeriodicMemoryLogging(ctx context.Context, logger *Logger, level Level, interval time.Duration, gcCall bool) {
	go RunPeriodicMemoryLogging(ctx, logger, level, interval, gcCall)
}

func MemoryStatisticLogging(logger *Logger, level Level) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	logger.LogAttrs(
		context.Background(),
		level,
		"runtime MemStats",
		StringAttr("alloc", units.HumanSize(float64(m.Alloc))),
		StringAttr("total_alloc", units.HumanSize(float64(m.TotalAlloc))),
		StringAttr("sys", units.HumanSize(float64(m.Sys))),
		UInt32Attr("num_gc", m.NumGC),
	)
}
