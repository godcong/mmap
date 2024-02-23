package mmap

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

func openMem(id int, size int) (*MapMem, error) {
	flProtect := uint32(syscall.PAGE_READONLY)
	dwDesiredAccess := uint32(syscall.FILE_MAP_READ)
	owner := false
	if id == 0 {
		owner = true
		id = GenKey()
		flProtect = syscall.PAGE_READWRITE
		dwDesiredAccess = syscall.FILE_MAP_WRITE
	}

	maxSizeHigh := uint32((0 + int64(size)) >> 32)
	maxSizeLow := uint32((0 + int64(size)) & 0xFFFFFFFF)

	wname, _ := syscall.UTF16PtrFromString(fmt.Sprintf("mmap_%d_index", id))
	h, errno := syscall.CreateFileMapping(0, nil, flProtect, maxSizeHigh, maxSizeLow, wname)
	if errno != nil {
		return nil, os.NewSyscallError("CreateFileMapping", errno)
	}
	// c := &close{handle: h}
	// runtime.SetFinalizer(c, (*close).Close)
	// Actually map a view of the data into memory. The view's size
	// is the length the user requested.
	fileOffsetHigh := uint32(0 >> 32)
	fileOffsetLow := uint32(0 & 0xFFFFFFFF)
	ptr, errno := syscall.MapViewOfFile(h, dwDesiredAccess, fileOffsetHigh, fileOffsetLow, uintptr(size))
	if errno != nil {
		return nil, os.NewSyscallError("MapViewOfFile", errno)
	}
	fd := &MapMem{
		owner: owner,
		id:    id,
		data:  PtrToBytes(ptr, int(size)),
		close: closeHandle(uintptr(h)),
	}
	runtime.SetFinalizer(fd, (*MapMem).Close)
	return fd, nil
}

func (f *MapMem) Sync() error {
	if !f.owner {
		return ErrBadFileDesc
	}

	errno := syscall.FlushViewOfFile(BytesToPtr(f.data), uintptr(len(f.data)))
	if errno != nil {
		return os.NewSyscallError("FlushViewOfFile", errno)
	}

	// memory has no file
	// err = syscall.FlushFileBuffers(syscall.Handle(f.fd.Fd()))
	// if err != nil {
	// 	return fmt.Errorf("mmap: could not sync file buffers: %w readOnly", err)
	// }

	return nil
}

func (f *MapMem) Close() (err error) {
	if f.data == nil {
		return nil
	}
	_ = f.Sync()

	addr := BytesToPtr(f.data)
	f.data = nil
	runtime.SetFinalizer(f, nil)
	err = syscall.UnmapViewOfFile(addr)
	if err != nil {
		return err
	}
	return f.close()
}
