//go:build linux || darwin || freebsd

package mmap

import (
	"os"
	"runtime"

	syscall "golang.org/x/sys/unix"
)

func openMapMem(id int, size int) (*MapMem, error) {
	var err error
	owner := false
	closer := func() error { return nil }
	size = getPageSize(size)
	if id == MapMemKeyInvalid {
		owner = true
	}
	if owner {
		k := GenKey()
		id, err = syscall.SysvShmGet(k, size, syscall.IPC_CREAT|syscall.IPC_EXCL|0o600)
		if err != nil {
			return nil, os.NewSyscallError("SysvShmGet", err)
		}

		Log().Info("OpenMapMem with owner", "new_id", id, "key", k, "size", size)

		closer = closeShm(id)
	} else {
		Log().Info("OpenMapMem with friendly", "id", id, "size", size)
	}

	data, err := syscall.SysvShmAttach(id, 0, 0)
	if err != nil {
		return nil, os.NewSyscallError("SysvShmAttach", err)
	}

	fd := &MapMem{
		id:    id,
		owner: owner,
		data:  data[:size],
		close: closer,
	}
	runtime.SetFinalizer(fd, (*MapMem).Close)
	return fd, nil
}

func (f *MapMem) Close() (err error) {
	err = syscall.SysvShmDetach(f.data)
	if err != nil {
		return os.NewSyscallError("SysvShmDetach", err)
	}

	f.data = nil
	return f.close()
}

func closeShm(id int) func() error {
	return func() error {
		_, err := syscall.SysvShmCtl(id, syscall.IPC_RMID, nil)
		if err != nil {
			return os.NewSyscallError("SysvShmCtl", err)
		}
		return nil
	}
}
