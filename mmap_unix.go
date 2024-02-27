//go:build !windows

package mmap

import (
	syscall "golang.org/x/sys/unix"
)

const (
	PROT_NONE      = syscall.PROT_NONE
	PROT_READ      = syscall.PROT_READ
	PROT_WRITE     = syscall.PROT_WRITE
	PROT_EXEC      = syscall.PROT_EXEC
	PROT_GROWSDOWN = syscall.PROT_GROWSDOWN
	PROT_GROWSUP   = syscall.PROT_GROWSUP

	// MAP_32BIT      = 0x40
	// MAP_ANON       = 0x20
	// MAP_ANONYMOUS  = 0x20
	// MAP_DENYWRITE  = 0x800
	// MAP_EXECUTABLE = 0x1000
	// MAP_FILE       = 0x0
	// MAP_FIXED      = 0x10
	// MAP_GROWSDOWN  = 0x100
	// MAP_HUGETLB    = 0x40000
	// MAP_LOCKED     = 0x2000
	// MAP_NONBLOCK   = 0x10000
	// MAP_NORESERVE  = 0x4000
	// MAP_POPULATE   = 0x8000
	// MAP_PRIVATE    = 0x2
	MAP_SHARED = syscall.MAP_SHARED
	// MAP_STACK      = 0x20000
	// MAP_TYPE       = 0xf
)

// Mmap maps length bytes of the file represented by the file descriptor fd into memory, starting at the byte offset.
//
// fd int - file descriptor
// offset int64 - byte offset
// length int - length in bytes
// prot int - memory protection
// flags int - mapping flags
// []byte, error - mapped data and an error
func Mmap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	return syscall.Mmap(fd, offset, length, prot, flags)
}

// Munmap unmaps the given byte slice.
//
// It takes a byte slice as a parameter and returns an error.
func Munmap(b []byte) (err error) {
	return syscall.Munmap(b)
}

// Mlock locks the given byte slice.
//
// It takes a byte slice as a parameter and returns an error.
func Mlock(b []byte) (err error) {
	return syscall.Mlock(b)
}
