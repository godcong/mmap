//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package mmap

import (
	"runtime"

	syscall "golang.org/x/sys/unix"
)

func createMem(id int, prot int, size int) (*MemFile, error) {
	writable := false
	switch {
	case prot&PROT_WRITE != 0:
		writable = true
	}
	var err error
	owner := false
	if id == 0 {
		owner = true
		id, err = syscall.SysvShmGet(GenKey(), size, syscall.IPC_CREAT)
		if err != nil {
			return nil, err
		}
	}
	data, err := syscall.SysvShmAttach(id, 0, 0)
	if err != nil {
		return nil, err
	}

	fd := &MemFile{
		id:     id,
		owner:  owner,
		data:   data,
		rdOnly: !writable,
	}
	runtime.SetFinalizer(fd, (*MemFile).Close)
	return fd, nil
}

func (f *MemFile) Close() (err error) {
	if f.owner {
		_, err = syscall.SysvShmCtl(f.id, syscall.IPC_RMID, nil)
	}
	_ = syscall.SysvShmDetach(f.data)
	return
}
