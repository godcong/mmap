package mmap

import (
	"syscall"
)

const (
	EINVAL          = syscall.EINVAL
	defaultPageSize = 4096
)

type Handle = syscall.Handle

var (
	pageSize uint32
)

func init() {
	pageSize = defaultPageSize
}

func PageSize() uint32 {
	return pageSize
}

func InitPageSize(size uint32) {
	pageSize = size
}
