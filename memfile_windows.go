package mmap

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

func createMem(id int, prot int, size int) (*MemFile, error) {
	flProtect := uint32(syscall.PAGE_READONLY)
	dwDesiredAccess := uint32(syscall.FILE_MAP_READ)
	writable := false
	if prot&PROT_WRITE != 0 {
		flProtect = syscall.PAGE_READWRITE
		dwDesiredAccess = syscall.FILE_MAP_WRITE
		writable = true
	}

	maxSizeHigh := uint32((0 + int64(size)) >> 32)
	maxSizeLow := uint32((0 + int64(size)) & 0xFFFFFFFF)
	owner := false
	if id == 0 {
		owner = true
		id = GenKey()
	}

	wname, _ := syscall.UTF16PtrFromString(fmt.Sprintf("mmap_%d_index", id))
	h, errno := syscall.CreateFileMapping(0, nil, flProtect, maxSizeHigh, maxSizeLow, wname)
	if errno != nil {
		return nil, os.NewSyscallError("CreateFileMapping", errno)
	}
	runtime.SetFinalizer(&close{handle: h}, (*close).Close)
	// Actually map a view of the data into memory. The view's size
	// is the length the user requested.
	fileOffsetHigh := uint32(0 >> 32)
	fileOffsetLow := uint32(0 & 0xFFFFFFFF)
	ptr, errno := syscall.MapViewOfFile(h, dwDesiredAccess, fileOffsetHigh, fileOffsetLow, uintptr(size))
	if errno != nil {
		return nil, os.NewSyscallError("MapViewOfFile", errno)
	}
	data := PtrToBytes(ptr, int(size))
	fd := &MemFile{
		owner:  owner,
		id:     id,
		data:   data,
		rdOnly: !writable,
	}
	runtime.SetFinalizer(fd, (*MemFile).Close)
	return fd, nil
}

func (f *MemFile) Sync() error {
	if f.rdOnly {
		return errBadFD
	}

	err := syscall.FlushViewOfFile(BytesToPtr(f.data), uintptr(len(f.data)))
	if err != nil {
		return fmt.Errorf("mmap: could not sync view: %w readOnly", err)
	}

	// memory has no file
	// err = syscall.FlushFileBuffers(syscall.Handle(f.fd.Fd()))
	// if err != nil {
	// 	return fmt.Errorf("mmap: could not sync file buffers: %w readOnly", err)
	// }

	return nil
}

func (f *MemFile) Close() (err error) {
	if f.data == nil {
		return nil
	}
	_ = f.Sync()

	addr := BytesToPtr(f.data)
	f.data = nil
	runtime.SetFinalizer(f, nil)
	return syscall.UnmapViewOfFile(addr)
}
