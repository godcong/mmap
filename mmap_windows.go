//go:build windows

package mmap

import (
	"os"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

const (
	PROT_NONE  = 0x0
	PROT_READ  = 0x1
	PROT_WRITE = 0x2
	PROT_EXEC  = 0x4
	PROT_COPY  = 0x8
	// PROT_GROWSDOWN = 0x1000000
	// PROT_GROWSUP   = 0x2000000
	//
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
	MAP_SHARED = 0x1
	// MAP_STACK      = 0x20000
	// MAP_TYPE       = 0xf
)

type active struct {
	data []byte
	fd   syscall.Handle
}

type mmapper struct {
	sync.Mutex
	active map[*byte]*active // active mappings; key is last byte in mapping
	mmap   func(addr, length uintptr, prot, flags, fd int, offset int64) (uintptr, error)
	munmap func(addr uintptr, length uintptr) error
}

var mapper = &mmapper{
	active: make(map[*byte]*active),
	mmap:   mmap,
	munmap: munmap,
}

func (m *mmapper) Mmap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	if length <= 0 {
		return nil, EINVAL
	}

	// Map the requested memory.
	addr, errno := m.mmap(0, uintptr(length), prot, flags, fd, offset)
	if errno != nil {
		return nil, errno
	}

	// Use unsafe to turn addr into a []byte.
	data = PtrToBytes(addr, length)

	// Register mapping in m and return it.
	p := &data[cap(data)-1]
	m.Lock()
	defer m.Unlock()
	m.active[p] = &active{data: data, fd: syscall.Handle(fd)}
	return data, nil
}

func (m *mmapper) Munmap(data []byte) (err error) {
	if len(data) == 0 || len(data) != cap(data) {
		return EINVAL
	}
	// Find the base of the mapping.
	p := &data[cap(data)-1]
	m.Lock()
	defer m.Unlock()
	b := m.active[p]
	if b == nil || &b.data[0] != &data[0] {
		return EINVAL
	}
	err = flush(b, data, uintptr(len(data)))
	if err != nil {
		return err
	}

	// Unmap the memory and update m.
	if errno := m.munmap(BytesToPtr(data), uintptr(len(b.data))); errno != nil {
		return errno
	}
	delete(m.active, p)
	return nil
}

// Mmap maps the contents of the file at the given path.
func Mmap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	return mapper.Mmap(fd, offset, length, prot, flags)
}

// Munmap unmaps the memory referenced by data.
func Munmap(data []byte) (err error) {
	return mapper.Munmap(data)
}

func mmap(addr uintptr, length uintptr, prot int, flags int, fd int, offset int64) (xaddr uintptr, err error) {
	flProtect := uint32(syscall.PAGE_READONLY)
	dwDesiredAccess := uint32(syscall.FILE_MAP_READ)
	// writable := false
	switch {
	case prot&PROT_COPY != 0:
		flProtect = syscall.PAGE_WRITECOPY
		dwDesiredAccess = syscall.FILE_MAP_COPY
		// writable = true
	case prot&PROT_WRITE != 0:
		flProtect = syscall.PAGE_READWRITE
		dwDesiredAccess = syscall.FILE_MAP_WRITE
		// writable = true
	}
	if prot&PROT_EXEC != 0 {
		flProtect <<= 4
		dwDesiredAccess |= syscall.FILE_MAP_EXECUTE
	}

	// The maximum size is the area of the file, starting from 0,
	// that we wish to allow to be mappable. It is the sum of
	// the length the user requested, plus the offset where that length
	// is starting from. This does not map the data into memory.
	maxSizeHigh := uint32((offset + int64(length)) >> 32)
	maxSizeLow := uint32((offset + int64(length)) & 0xFFFFFFFF)
	h, errno := syscall.CreateFileMapping(syscall.Handle(fd), makeInheritSa(), flProtect, maxSizeHigh, maxSizeLow, nil)
	if errno != nil {
		return xaddr, os.NewSyscallError("CreateFileMapping", errno)
	}
	runtime.SetFinalizer(&close{handle: h}, (*close).Close)
	// defer syscall.CloseHandle(h)
	// Actually map a view of the data into memory. The view's size
	// is the length the user requested.
	fileOffsetHigh := uint32(offset >> 32)
	fileOffsetLow := uint32(offset & 0xFFFFFFFF)
	ptr, errno := syscall.MapViewOfFile(h, dwDesiredAccess, fileOffsetHigh, fileOffsetLow, uintptr(length))
	if errno != nil {
		return xaddr, os.NewSyscallError("MapViewOfFile", errno)
	}
	return ptr, nil
}

func flush(active *active, data []byte, len uintptr) (err error) {
	errno := syscall.FlushViewOfFile(uintptr(unsafe.Pointer(&data[0])), len)
	if errno != nil {
		return os.NewSyscallError("FlushViewOfFile", errno)
	}
	errno = syscall.FlushFileBuffers(active.fd)
	return os.NewSyscallError("FlushFileBuffers", errno)
}

func munmap(addr uintptr, length uintptr) (err error) {
	return syscall.UnmapViewOfFile(addr)
}

func makeInheritSa() *syscall.SecurityAttributes {
	var sa syscall.SecurityAttributes
	sa.Length = uint32(unsafe.Sizeof(sa))
	sa.InheritHandle = 1
	return &sa
}

type close struct {
	handle syscall.Handle
}

func (c *close) Close() error {
	return syscall.CloseHandle(c.handle)
}
