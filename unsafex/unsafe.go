package unsafex

import (
	"unsafe"
)

// PtrToBytes converts a pointer to a byte slice of length n
func PtrToBytes(ptr uintptr, n int) []byte {
	return ptrToBytes(ptr, n)
}

// BytesToPtr converts a byte slice to a pointer
func BytesToPtr(data []byte) uintptr {
	return uintptr(unsafe.Pointer(&data[0]))
}

// BytesToPoint converts a byte slice to a pointer to a byte
func BytesToPoint(data []byte) *byte {
	return bytesToPoint(data)
}

// PointToBytes converts a pointer to a byte slice of length n
func PointToBytes(ptr *byte, n int) []byte {
	return unsafe.Slice(ptr, n)
}

// CopyBytesToPtr copies a byte slice to a pointer
func CopyBytesToPtr(dst uintptr, src []byte) {
	copy(PtrToBytes(dst, len(src)), src)
}
