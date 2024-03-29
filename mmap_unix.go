//go:build linux || darwin || freebsd

package mmap

import (
	syscall "golang.org/x/sys/unix"
)

const (
	PROT_NONE  = syscall.PROT_NONE
	PROT_READ  = syscall.PROT_READ
	PROT_WRITE = syscall.PROT_WRITE
	PROT_EXEC  = syscall.PROT_EXEC
	// PROT_GROWSDOWN = syscall.PROT_GROWSDOWN
	// PROT_GROWSUP   = syscall.PROT_GROWSUP

	MAP_SHARED = syscall.MAP_SHARED
)

// Mmap description of the Go function.
//
// Takes fd, offset, length, prot, and flags as parameters.
// Returns data []byte and err error.
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
