//go:build !go1.21

package mmap

import "golang.org/x/exp/slog"

func Log() *slog.Logger {
	return slog.Default()
}
