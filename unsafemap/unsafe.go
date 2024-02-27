package unsafemap

import (
	"unsafe"
)

func PtrToBytes(ptr uintptr, n int) []byte {
	return ptrToBytes(ptr, n)
}

func BytesToPtr(data []byte) uintptr {
	return uintptr(unsafe.Pointer(&data[:1][0]))
}

func BytesToPoint(data []byte) *byte {
	return bytesToPoint(data)
}

func PointToBytes(ptr *byte, n int) []byte {
	return unsafe.Slice(ptr, n)
}

func CopyBytesToPtr(dst uintptr, src []byte) {
	copy(PtrToBytes(dst, len(src)), src)
}
