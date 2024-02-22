// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mmap provides a way to memory-map a file.
package mmap

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"syscall"
)

// MapFile reads/writes a memory-mapped file.
type MapFile struct {
	data   []byte
	off    int
	rdOnly bool

	fd *os.File
}

// Len returns the length of the underlying memory-mapped file.
func (f *MapFile) Len() int {
	return len(f.data)
}

// At returns the byte at index i.
func (f *MapFile) At(i int) byte {
	return f.data[i]
}

// Stat returns the MapFileInfo structure describing file.
// If there is an error, it will be of type *os.PathError.
func (f *MapFile) Stat() (os.FileInfo, error) {
	if f == nil {
		return nil, os.ErrInvalid
	}

	return f.fd.Stat()
}

func (f *MapFile) rflag() bool {
	return true
}

func (f *MapFile) wflag() bool {
	return !f.rdOnly
}

// Read implements the io.Reader interface.
func (f *MapFile) Read(p []byte) (int, error) {
	if f == nil {
		return 0, os.ErrInvalid
	}

	if !f.rflag() {
		return 0, errBadFD
	}
	if f.off >= len(f.data) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.off:])
	f.off += n
	return n, nil
}

// ReadByte implements the io.ByteReader interface.
func (f *MapFile) ReadByte() (byte, error) {
	if f == nil {
		return 0, os.ErrInvalid
	}

	if !f.rflag() {
		return 0, errBadFD
	}
	if f.off >= len(f.data) {
		return 0, io.EOF
	}
	v := f.data[f.off]
	f.off++
	return v, nil
}

// ReadAt implements the io.ReaderAt interface.
func (f *MapFile) ReadAt(p []byte, off int64) (int, error) {
	if f == nil {
		return 0, os.ErrInvalid
	}

	if !f.rflag() {
		return 0, errBadFD
	}
	if f.data == nil {
		return 0, errors.New("mmap: closed")
	}
	if off < 0 || int64(len(f.data)) < off {
		return 0, fmt.Errorf("mmap: invalid ReadAt offset %d", off)
	}
	n := copy(p, f.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// Write implements the io.Writer interface.
func (f *MapFile) Write(p []byte) (int, error) {
	if f == nil {
		return 0, os.ErrInvalid
	}

	if !f.wflag() {
		return 0, errBadFD
	}
	if f.off >= len(f.data) {
		return 0, io.ErrShortWrite
	}
	n := copy(f.data[f.off:], p)
	f.off += n
	if len(p) > n {
		return n, io.ErrShortWrite
	}
	return n, nil
}

// WriteByte implements the io.ByteWriter interface.
func (f *MapFile) WriteByte(c byte) error {
	if f == nil {
		return os.ErrInvalid
	}

	if !f.wflag() {
		return errBadFD
	}
	if f.off >= len(f.data) {
		return io.ErrShortWrite
	}
	f.data[f.off] = c
	f.off++
	return nil
}

// WriteAt implements the io.WriterAt interface.
func (f *MapFile) WriteAt(p []byte, off int64) (int, error) {
	if f == nil {
		return 0, os.ErrInvalid
	}

	if !f.wflag() {
		return 0, errBadFD
	}
	if f.data == nil {
		return 0, errors.New("mmap: closed")
	}
	if off < 0 || int64(len(f.data)) < off {
		return 0, fmt.Errorf("mmap: invalid WriteAt offset %d", off)
	}
	n := copy(f.data[off:], p)
	if n < len(p) {
		return n, io.ErrShortWrite
	}
	return n, nil
}

func (f *MapFile) Seek(offset int64, whence int) (int64, error) {
	if f == nil {
		return 0, os.ErrInvalid
	}

	switch whence {
	case io.SeekStart:
		f.off = int(offset)
	case io.SeekCurrent:
		f.off += int(offset)
	case io.SeekEnd:
		f.off = len(f.data) - int(offset)
	default:
		return 0, fmt.Errorf("mmap: invalid whence")
	}
	if f.off < 0 {
		return 0, fmt.Errorf("mmap: negative position")
	}
	return int64(f.off), nil
}

func (f *MapFile) Fd() uintptr {
	return f.fd.Fd()
}

var errBadFD = errors.New("bad file descriptor")

// Open memory-maps the named file for reading.
func Open(filename string) (*MapFile, error) {
	return openFile(filename, os.O_RDONLY, 0)
}

// OpenFile memory-maps the named file for reading/writing, depending on
// the flag value.
func OpenFile(filename string, flag int, mode os.FileMode) (*MapFile, error) {
	return openFile(filename, flag, mode)
}

func openFile(filename string, mode int, perm os.FileMode) (*MapFile, error) {
	if len(filename) == 0 {
		return nil, syscall.ENOENT
	}

	f, err := os.OpenFile(filename, mode, perm)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := fi.Size()
	//
	// if size == 0 {
	// 	return &MapFile{rdOnly: mode&os.O_RDONLY == os.O_RDONLY}, nil
	// }
	// if size < 0 {
	// 	return nil, fmt.Errorf("mmap: file %q has negative size", filename)
	// }
	// if size != int64(int(size)) {
	// 	return nil, fmt.Errorf("mmap: file %q is too large", filename)
	// }

	prot := PROT_READ
	rdOnly := true
	switch mode & (os.O_RDONLY | os.O_WRONLY | os.O_RDWR) {
	case os.O_RDONLY:
		prot = PROT_READ
	case os.O_WRONLY:
		prot = PROT_WRITE
		rdOnly = false
	case os.O_RDWR:
		prot = PROT_READ | PROT_WRITE
		rdOnly = false
	}
	data, err := Mmap(int(f.Fd()), 0, int(size), prot, MAP_SHARED)
	fd := &MapFile{
		data:   data,
		fd:     f,
		rdOnly: rdOnly,
	}
	runtime.SetFinalizer(fd, (*MapFile).Close)
	return fd, nil

}

var (
	_ io.Reader     = (*MapFile)(nil)
	_ io.ReaderAt   = (*MapFile)(nil)
	_ io.ByteReader = (*MapFile)(nil)
	_ io.Writer     = (*MapFile)(nil)
	_ io.WriterAt   = (*MapFile)(nil)
	_ io.ByteWriter = (*MapFile)(nil)
	_ io.Closer     = (*MapFile)(nil)
	_ io.Seeker     = (*MapFile)(nil)
)
