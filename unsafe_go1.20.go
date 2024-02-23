//go:build go1.20

package mmap

import (
	"fmt"
	"unsafe"
)

func ptrToBytes(ptr uintptr, n int) []byte {
	// return []byte(syscall.BytePtrToString(p))
	// return unsafe.Slice(unsafe.Pointer(ptr), n)
	// return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
	// 	Data: ptr,
	// 	Len:  n,
	// 	Cap:  n,
	// }))
	fmt.Println("num", n)
	return *(*[]byte)(unsafe.Pointer(&ptr))
}

func bytesToPoint(data []byte) *byte {
	return unsafe.SliceData(data)
}
