//go:build !go1.20

package mmap

import (
	"reflect"
	"unsafe"
)

func ptrToBytes(ptr uintptr, n int) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: ptr,
		Len:  n,
		Cap:  n,
	}))
}

func bytesToPoint(data []byte) *byte {
	return (*byte)(unsafe.Pointer(&data[:1][0]))
}
