package glog

import (
	"context"
	"log/slog"
	"runtime"
	"testing"
	"time"

	"github.com/docker/go-units"
)

func TestMemoryStatisticLogging(t *testing.T) {
	var logRecords []Record

	logger := NewRecordsLogger(&logRecords)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	MemoryStatisticLogging(logger, LevelInfo)

	if count := len(logRecords); count != 1 {
		t.Errorf("excepted 1 log record, got %d", count)
	}
	err := checkLogRecord(
		logRecords[0],
		LevelInfo,
		"runtime MemStats",
		[]Attr{
			StringAttr("alloc", units.HumanSize(float64(m.Alloc))),
			StringAttr("total_alloc", units.HumanSize(float64(m.TotalAlloc))),
			StringAttr("sys", units.HumanSize(float64(m.Sys))),
			UInt32Attr("num_gc", m.NumGC),
		},
	)
	if err != nil {
		t.Errorf("check log record error: %s", err.Error())
	}
}

func TestPeriodicMemoryLogging(t *testing.T) {
	var logRecords []Record
	logger := NewRecordsLogger(&logRecords)

	// User logger
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		RunPeriodicMemoryLogging(ctx, logger, LevelInfo, 10*time.Millisecond, false)
	}()
	time.Sleep(15 * time.Millisecond)
	cancel()

	if count := len(logRecords); count != 1 {
		t.Errorf("excepted 1 log record, got %d", count)
	}
	err := checkLogRecord(
		logRecords[0],
		LevelInfo,
		"runtime MemStats",
		[]Attr{StringAttr("name", "memory_stat")},
	)
	if err != nil {
		t.Errorf("check log record error: %s", err.Error())
	}

	// Default logger
	logRecords = logRecords[:0]
	slog.SetDefault(logger)

	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		RunPeriodicMemoryLogging(ctx, nil, LevelInfo, 10*time.Millisecond, true)
	}()
	time.Sleep(15 * time.Millisecond)
	cancel()

	if count := len(logRecords); count != 1 {
		t.Errorf("excepted 1 log record, got %d", count)
	}
	err = checkLogRecord(
		logRecords[0],
		LevelInfo,
		"runtime MemStats",
		[]Attr{StringAttr("name", "memory_stat")},
	)
	if err != nil {
		t.Errorf("check log record error: %s", err.Error())
	}

	// StartPeriodicMemoryLogging goroutine
	logRecords = logRecords[:0]
	slog.SetDefault(logger)

	ctx, cancel = context.WithCancel(context.Background())
	go StartPeriodicMemoryLogging(ctx, nil, LevelInfo, 10*time.Millisecond, false)
	time.Sleep(15 * time.Millisecond)
	cancel()

	if count := len(logRecords); count != 1 {
		t.Errorf("excepted 1 log record, got %d", count)
	}
	err = checkLogRecord(
		logRecords[0],
		LevelInfo,
		"runtime MemStats",
		[]Attr{StringAttr("name", "memory_stat")},
	)
	if err != nil {
		t.Errorf("check log record error: %s", err.Error())
	}

}
