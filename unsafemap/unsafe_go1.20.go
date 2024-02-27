//go:build go1.20

package unsafemap

import (
	"unsafe"
)

type Slice struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func ptrToBytes(ptr uintptr, n int) []byte {
	return *(*[]byte)(unsafe.Pointer(&Slice{Data: unsafe.Pointer(ptr), Len: n, Cap: n}))
}

func bytesToPoint(data []byte) *byte {
	return unsafe.SliceData(data)
}
