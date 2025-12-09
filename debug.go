package mmap

import (
	"log/slog"
	"os"
	"sync"
)

var (
	logger     *slog.Logger
	loggerOnce sync.Once
	debugMode  bool
)

func init() {
	debugMode = os.Getenv("GO_MMAP_DEBUG") != ""
}

// Log returns a structured logger for mmap operations
func Log() *slog.Logger {
	loggerOnce.Do(func() {
		opts := &slog.HandlerOptions{}
		if !debugMode {
			opts.Level = slog.LevelError // 只记录错误级别日志
		}

		handler := slog.NewTextHandler(os.Stdout, opts)
		logger = slog.New(handler)
	})
	return logger
}

// DebugLogEnabled returns whether debug logging is enabled
func DebugLogEnabled() bool {
	return debugMode
}
