package mmap

import (
	"log/slog"
	"os"
)

var debug = false

func init() {
	if os.Getenv("GO_MMAP_DEBUG") != "" {
		debug = true
	}
}

func Log() *slog.Logger {
	return slog.Default()
}
