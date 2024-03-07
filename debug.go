package mmap

import (
	"io"
	"log"
	"log/slog"
	"os"
)

var debug = true

func init() {
	if os.Getenv("GO_MMAP_DEBUG") != "" {
		debug = true
	}

	if !debug {
		log.SetOutput(io.Discard)
	}
}

func Log() *slog.Logger {
	return slog.Default()
}
