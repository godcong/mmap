//go:build windows

package mmap

import (
	"fmt"
	"os"
	"sync"
	sys "syscall"
	"unsafe"

	"github.com/godcong/mmap/unsafex"
	syscall "golang.org/x/sys/windows"
)

const (
	PROT_NONE  = 0x0
	PROT_READ  = 0x1
	PROT_WRITE = 0x2
	PROT_EXEC  = 0x4
	PROT_COPY  = 0x8
	// PROT_GROWSDOWN = 0x1000000
	// PROT_GROWSUP   = 0x2000000

	MAP_SHARED = 0x1
)

type Handle = syscall.Handle

type active struct {
	data  []byte
	fd    uintptr
	close func() error
}

type mmapper struct {
	sync.Mutex
	active map[*byte]*active // active mappings; key is last byte in mapping
	mmap   func(addr, length uintptr, prot, flags, fd int, offset int64) (uintptr, uintptr, error)
	munmap func(addr uintptr, length uintptr) error
}

var (
	modkernel32          = syscall.NewLazySystemDLL("kernel32.dll")
	procOpenFileMappingW = modkernel32.NewProc("OpenFileMappingW")
)

var mapper = &mmapper{
	active: make(map[*byte]*active),
	mmap:   mmap,
	munmap: munmap,
}

// Mmap maps the requested memory.
//
// Parameters: fd int, offset int64, size int, prot int, flags int.
// Returns: data []byte, err error.
func (m *mmapper) Mmap(fd int, offset int64, size int, prot int, flags int) (data []byte, err error) {
	// Map the requested memory.
	handle, mapview, err := m.mmap(0, uintptr(size), prot, flags, fd, offset)
	if err != nil {
		return nil, err
	}

	var info syscall.MemoryBasicInformation
	err = syscall.VirtualQuery(mapview, &info, unsafe.Sizeof(info))
	if err != nil {
		return nil, os.NewSyscallError("VirtualQuery", err)
	}

	bufHdr := (*unsafex.Slice)(unsafe.Pointer(&data))
	bufHdr.Data = unsafe.Pointer(mapview)
	bufHdr.Len = int(size)
	bufHdr.Cap = int(size)
	// Register mapping in m and return it.
	p := &data[cap(data)-1]
	m.Lock()
	defer m.Unlock()
	m.active[p] = &active{data: data, fd: uintptr(fd), close: closeHandle(handle)}
	return data, nil
}

// Munmap unmaps the memory and updates the mmapper.
//
// It takes a data []byte as a parameter and returns an error.
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
	// Unmap the memory and update m.
	if err := m.munmap(unsafex.BytesToPtr(data), uintptr(len(b.data))); err != nil {
		_ = b.close()
		return fmt.Errorf("error unmapping handle: %s", err)
	}
	err = b.close()
	if err != nil {
		return fmt.Errorf("error closing handle: %s", err)
	}
	delete(m.active, p)
	return nil
}

// Flush description of the Go function.
//
// data []byte, sz uintptr.
// error.
func (m *mmapper) Flush(data []byte, sz uintptr) (err error) {
	p := &data[cap(data)-1]
	mapper.Lock()
	defer mapper.Unlock()
	b := mapper.active[p]
	return flush(Handle(b.fd), data, sz)
}

// Mmap maps the contents of the file at the given path.
func Mmap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	return mapper.Mmap(fd, offset, length, prot, flags)
}

// Munmap unmaps the memory referenced by data.
func Munmap(data []byte) (err error) {
	return mapper.Munmap(data)
}

// Flush flushes the data to memory referenced.
func Flush(data []byte, size uintptr) (err error) {
	return mapper.Flush(data, size)
}

func mmap(addr uintptr, length uintptr, prot int, flags int, fd int, offset int64) (handle, xaddr uintptr, err error) {
	flProtect := uint32(syscall.PAGE_READONLY)
	dwDesiredAccess := uint32(syscall.FILE_MAP_READ)
	switch {
	case prot&PROT_COPY != 0:
		flProtect = syscall.PAGE_WRITECOPY
		dwDesiredAccess = syscall.FILE_MAP_COPY
	case prot&PROT_WRITE != 0:
		flProtect = syscall.PAGE_READWRITE
		dwDesiredAccess = syscall.FILE_MAP_WRITE
	}
	if prot&PROT_EXEC != 0 {
		flProtect <<= 4
		dwDesiredAccess |= syscall.FILE_MAP_EXECUTE
	}

	// The maximum size is the area of the file, starting from 0,
	// that we wish to allow to be mappable. It is the sum of
	// the length the user requested, plus the offset where that length
	// is starting from. This does not map the data into memory.
	low, high := uint32(length), uint32(length>>32)
	h, errno := syscall.CreateFileMapping(Handle(fd), makeInheritSa(), flProtect, high, low, nil)
	if errno != nil {
		return handle, xaddr, os.NewSyscallError("CreateFileMapping", errno)
	}
	// Actually map a view of the data into memory. The view's size
	// is the length the user requested.
	// fileOffsetHigh := uint32(offset >> 32)
	// fileOffsetLow := uint32(offset & 0xFFFFFFFF)
	ptr, errno := syscall.MapViewOfFile(h, dwDesiredAccess, 0, 0, length)
	if errno != nil {
		_ = syscall.CloseHandle(h)
		return handle, xaddr, os.NewSyscallError("MapViewOfFile", errno)
	}

	return uintptr(h), ptr, nil
}

func flush(fd syscall.Handle, data []byte, len uintptr) (err error) {
	errno := syscall.FlushViewOfFile(unsafex.BytesToPtr(data), len)
	if errno != nil {
		return os.NewSyscallError("FlushViewOfFile", errno)
	}
	errno = syscall.FlushFileBuffers(fd)
	return os.NewSyscallError("FlushFileBuffers", errno)
}

func munmap(addr uintptr, length uintptr) (err error) {
	errno := syscall.UnmapViewOfFile(addr)
	if errno != nil {
		return os.NewSyscallError("UnmapViewOfFile", errno)
	}
	return
}

func makeInheritSa() *syscall.SecurityAttributes {
	var sa syscall.SecurityAttributes
	sa.Length = uint32(unsafe.Sizeof(sa))
	sa.InheritHandle = 1
	return &sa
}

func closeHandle(handle uintptr) func() error {
	return func() error {
		err := syscall.CloseHandle(Handle(handle))
		if err != nil {
			return os.NewSyscallError("CloseHandle", err)
		}
		return nil
	}
}

func syscallOpenFileMapping(access uint32, bInheritHandle bool, lpName *uint16) (handle Handle, err error) {
	var _p0 uint32
	if bInheritHandle {
		_p0 = 1
	}
	r0, _, e1 := sys.SyscallN(procOpenFileMappingW.Addr(), uintptr(access), uintptr(_p0), uintptr(unsafe.Pointer(lpName)))
	handle = Handle(r0)
	if handle == 0 {
		err = e1
	}
	return
}
