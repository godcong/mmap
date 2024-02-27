//go:build go1.21

package mmap

import "log/slog"

func Log() *slog.Logger {
	return slog.Default()
}
