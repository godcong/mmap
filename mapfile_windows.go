// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mmap

import (
	"runtime"

	"github.com/godcong/mmap/unsafemap"
	syscall "golang.org/x/sys/windows"
)

// Sync commits the current contents of the file to stable storage.
func (f *MapFile) Sync() error {
	if !f.writable {
		return ErrBadFileDesc
	}

	return Flush(f.data, uintptr(len(f.data)))
}

// Close closes the reader.
func (f *MapFile) Close() error {
	if f.data == nil {
		return nil
	}
	defer f.fd.Close()
	// Sync the file before closing it.
	_ = f.Sync()

	data := f.data
	f.data = nil
	runtime.SetFinalizer(f, nil)
	return Munmap(data)
}

func closeMapFile(f *MapFile) error {
	if f.data == nil {
		return nil
	}
	_ = f.Sync()
	defer f.fd.Close()
	addr := unsafemap.BytesToPtr(f.data)
	f.data = nil
	runtime.SetFinalizer(f, nil)
	return syscall.UnmapViewOfFile(addr)
}
