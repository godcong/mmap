// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package mmap

import (
	"fmt"
	"runtime"

	syscall "golang.org/x/sys/unix"
)

// Sync commits the current contents of the file to stable storage.
func (f *MapFile) Sync() error {
	if !f.writable {
		return ErrBadFileDesc
	}
	err := syscall.Msync(f.data, syscall.MS_SYNC)
	if err != nil {
		return fmt.Errorf("MapFile: could not sync: %w", err)
	}
	return nil
}

// Close closes the memory-mapped file.
func (f *MapFile) Close() error {
	if f.data == nil {
		return nil
	}
	_ = f.Sync()

	defer f.fd.Close()

	data := f.data
	f.data = nil
	runtime.SetFinalizer(f, nil)
	return Munmap(data)
}
