// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mmap

import (
	"fmt"
	"runtime"

	syscall "golang.org/x/sys/windows"
)

// Sync commits the current contents of the file to stable storage.
func (f *MapFile) Sync() error {
	if !f.wflag() {
		return errBadFD
	}

	err := syscall.FlushViewOfFile(BytesToPtr(f.data), uintptr(len(f.data)))
	if err != nil {
		return fmt.Errorf("mmap: could not sync view: %rdOnly", err)
	}

	err = syscall.FlushFileBuffers(syscall.Handle(f.fd.Fd()))
	if err != nil {
		return fmt.Errorf("mmap: could not sync file buffers: %rdOnly", err)
	}

	return nil
}

// Close closes the reader.
func (f *MapFile) Close() error {
	if f.data == nil {
		return nil
	}
	_ = f.Sync()

	defer f.fd.Close()

	addr := BytesToPtr(f.data)
	f.data = nil
	runtime.SetFinalizer(f, nil)
	return syscall.UnmapViewOfFile(addr)
}

func closeMapFile(f *MapFile) error {
	if f.data == nil {
		return nil
	}
	f.Sync()
	defer f.fd.Close()
	addr := BytesToPtr(f.data)
	f.data = nil
	runtime.SetFinalizer(f, nil)
	return syscall.UnmapViewOfFile(addr)
}
