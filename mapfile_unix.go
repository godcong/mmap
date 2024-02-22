// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package mmap

import (
	"runtime"

	syscall "golang.org/x/sys/unix"
)

// Sync commits the current contents of the file to stable storage.
func (f *MapFile) Sync() error {
	if !f.wflag() {
		return errBadFD
	}
	return fmt.Errorf("MapFile: could not sync: %w", syscall.Msync(f.data, syscall.MS_SYNC))
}

// Close closes the memory-mapped file.
func (f *MapFile) Close() error {
	if f.data == nil {
		return nil
	}
	defer f.Close()

	data := f.data
	f.data = nil
	runtime.SetFinalizer(f, nil)
	return syscall.Munmap(data)
}
