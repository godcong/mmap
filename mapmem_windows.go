package mmap

import (
	"fmt"
	"os"
	"runtime"

	"github.com/godcong/mmap/unsafex"
	syscall "golang.org/x/sys/windows"
)

func openMapMem(id int, size int) (*MapMem, error) {
	owner := false
	if id == MapMemKeyInvalid {
		owner = true
		id = GenKey()
	}
	wname, _ := syscall.UTF16PtrFromString(fmt.Sprintf("mmap_%d_index", id))

	err := error(nil)
	handle := Handle(0)
	flProtect := syscall.PAGE_READONLY
	dwDesiredAccess := syscall.FILE_MAP_READ

	size = getPageSize(size)
	if owner {
		flProtect = syscall.PAGE_READWRITE
		dwDesiredAccess = syscall.FILE_MAP_WRITE
		low, high := uint32(size), uint32(size>>32)
		handle, err = syscall.CreateFileMapping(syscall.InvalidHandle, makeInheritSa(), uint32(flProtect), high, low, wname)
		if err != nil {
			return nil, os.NewSyscallError("CreateFileMapping", err)
		}
		// }
	} else {
		handle, err = syscallOpenFileMapping(uint32(dwDesiredAccess), true, wname)
		if err != nil {
			return nil, err
		}
	}

	// fileOffsetHigh := uint32(0 >> 32)
	// fileOffsetLow := uint32(0 & 0xFFFFFFFF)
	mapview, errno := syscall.MapViewOfFile(handle, uint32(dwDesiredAccess), 0, 0, uintptr(size))
	if errno != nil {
		return nil, os.NewSyscallError("MapViewOfFile", errno)
	}

	fd := &MapMem{
		owner: owner,
		id:    id,
		data:  unsafex.PtrToBytes(mapview, size),
		close: dummyCloser,
	}
	runtime.SetFinalizer(fd, (*MapMem).Close)
	return fd, nil
}

func (f *MapMem) Sync() error {
	if !f.owner {
		return ErrBadFileDesc
	}

	errno := syscall.FlushViewOfFile(unsafex.BytesToPtr(f.data), uintptr(len(f.data)))
	if errno != nil {
		return os.NewSyscallError("FlushViewOfFile", errno)
	}

	return nil
}

func (f *MapMem) Close() (err error) {
	if f.data == nil {
		return nil
	}
	_ = f.Sync()

	addr := unsafex.BytesToPtr(f.data)
	f.data = nil
	runtime.SetFinalizer(f, nil)
	err = syscall.UnmapViewOfFile(addr)
	if err != nil {
		return err
	}
	return f.close()
}
