package mmap

import "syscall"

const (
	EINVAL = syscall.EINVAL
	ENOENT = syscall.ENOENT
)

func dummyCloser() error { return nil }
