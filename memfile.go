package mmap

import (
	"io"
	"os"
	"unsafe"
)

const (
	pageSize = 4096
)

type MemFile struct {
	owner  bool
	id     int
	data   []byte
	rdOnly bool
	off    int
}

// Read implements the io.Reader interface.
func (f *MemFile) Read(p []byte) (int, error) {
	if f == nil {
		return 0, os.ErrInvalid
	}

	if f.off >= len(f.data) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.off:])
	f.off += n
	return n, nil
}

// Write implements the io.Writer interface.
func (f *MemFile) Write(p []byte) (int, error) {
	if f == nil {
		return 0, os.ErrInvalid
	}

	if f.rdOnly {
		return 0, errBadFD
	}
	if f.off >= len(f.data) {
		return 0, io.ErrShortWrite
	}
	n := copy(f.data[f.off:], p)
	f.off += n
	if len(p) > n {
		return n, io.ErrShortWrite
	}
	return n, nil
}

func (f *MemFile) ID() int {
	return f.id
}

func OpenMemFile(id int, prot int, size int) (*MemFile, error) {
	return createMem(id, prot, size)
}

func OpenMemFileS(id int, prot int) (*MemFile, error) {
	return createMem(id, prot, pageSize)
}

func PtrToBytes(ptr uintptr, n int) []byte {
	return ptrToBytes(ptr, n)
}

func BytesToPtr(data []byte) uintptr {
	return uintptr(unsafe.Pointer(&data[:1][0]))
}

func BytesToPoint(data []byte) *byte {
	return bytesToPtr(data)
}

func PointToBytes(ptr *byte, n int) []byte {
	return unsafe.Slice(ptr, n)
}

func CopyBytesToPtr(dst uintptr, src []byte) {
	copy(PtrToBytes(dst, len(src)), src)
}
