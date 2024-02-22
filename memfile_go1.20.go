//go:build go1.20

package mmap

import (
	"unsafe"
)

func ptrToBytes(ptr uintptr, n int) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(ptr)), n)
}

func bytesToPtr(data []byte) *byte {
	return unsafe.SliceData(data)
}
