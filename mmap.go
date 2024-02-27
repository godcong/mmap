package mmap

import "syscall"

const (
	// defaultPageSize = 4096
	EINVAL = syscall.EINVAL
	ENOENT = syscall.ENOENT
)

func dummyCloser() error { return nil }
