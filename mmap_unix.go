//go:build !windows

package mmap

import "syscall"

const (
	PROT_NONE      = 0x0
	PROT_READ      = 0x1
	PROT_WRITE     = 0x2
	PROT_EXEC      = 0x4
	PROT_GROWSDOWN = 0x1000000
	PROT_GROWSUP   = 0x2000000

	MAP_32BIT      = 0x40
	MAP_ANON       = 0x20
	MAP_ANONYMOUS  = 0x20
	MAP_DENYWRITE  = 0x800
	MAP_EXECUTABLE = 0x1000
	MAP_FILE       = 0x0
	MAP_FIXED      = 0x10
	MAP_GROWSDOWN  = 0x100
	MAP_HUGETLB    = 0x40000
	MAP_LOCKED     = 0x2000
	MAP_NONBLOCK   = 0x10000
	MAP_NORESERVE  = 0x4000
	MAP_POPULATE   = 0x8000
	MAP_PRIVATE    = 0x2
	MAP_SHARED     = 0x1
	MAP_STACK      = 0x20000
	MAP_TYPE       = 0xf
)

func Mmap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	return syscall.Mmap(fd, offset, length, prot, flags)
}

func Munmap(b []byte) (err error) {
	return syscall.Munmap(b)
}
