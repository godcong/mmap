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
	data     []byte
	off      int
	readOnly bool

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
		return nil, ErrInvalid
	}

	return f.fd.Stat()
}

// Read implements the io.Reader interface.
func (f *MapFile) Read(p []byte) (int, error) {
	if f == nil {
		return 0, ErrInvalid
	}

	if f.off >= len(f.data) {
		return 0, EOF
	}
	n := copy(p, f.data[f.off:])
	f.off += n
	return n, nil
}

// ReadByte implements the io.ByteReader interface.
func (f *MapFile) ReadByte() (byte, error) {
	if f == nil {
		return 0, ErrInvalid
	}

	if f.off >= len(f.data) {
		return 0, EOF
	}
	v := f.data[f.off]
	f.off++
	return v, nil
}

// ReadAt implements the io.ReaderAt interface.
func (f *MapFile) ReadAt(p []byte, off int64) (int, error) {
	if f == nil {
		return 0, ErrInvalid
	}

	if f.data == nil {
		return 0, errors.New("MapFile: closed")
	}
	if off < 0 || int64(len(f.data)) < off {
		return 0, fmt.Errorf("MapFile: invalid ReadAt offset %d", off)
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
		return 0, ErrInvalid
	}

	if f.readOnly {
		return 0, ErrBadFileDesc
	}
	if f.off >= len(f.data) {
		return 0, ErrShortWrite
	}
	n := copy(f.data[f.off:], p)
	f.off += n
	if len(p) > n {
		return n, ErrShortWrite
	}
	return n, nil
}

// WriteByte implements the io.ByteWriter interface.
func (f *MapFile) WriteByte(c byte) error {
	if f == nil {
		return ErrInvalid
	}

	if f.readOnly {
		return ErrBadFileDesc
	}
	if f.off >= len(f.data) {
		return ErrShortWrite
	}
	f.data[f.off] = c
	f.off++
	return nil
}

// WriteAt implements the io.WriterAt interface.
func (f *MapFile) WriteAt(p []byte, off int64) (int, error) {
	if f == nil {
		return 0, ErrInvalid
	}

	if f.readOnly {
		return 0, ErrBadFileDesc
	}
	if f.data == nil {
		return 0, errors.New("MapFile: closed")
	}
	if off < 0 || int64(len(f.data)) < off {
		return 0, fmt.Errorf("MapFile: invalid WriteAt offset %d", off)
	}
	n := copy(f.data[off:], p)
	if n < len(p) {
		return n, ErrShortWrite
	}
	return n, nil
}

func (f *MapFile) Seek(offset int64, whence int) (int64, error) {
	if f == nil {
		return 0, ErrInvalid
	}

	switch whence {
	case io.SeekStart:
		f.off = int(offset)
	case io.SeekCurrent:
		f.off += int(offset)
	case io.SeekEnd:
		f.off = len(f.data) - int(offset)
	default:
		return 0, fmt.Errorf("MapFile: invalid whence")
	}
	if f.off < 0 {
		return 0, fmt.Errorf("MapFile: negative position")
	}
	return int64(f.off), nil
}

func (f *MapFile) Fd() uintptr {
	return f.fd.Fd()
}

// Open memory-maps the named file for reading.
func Open(filename string) (*MapFile, error) {
	return openFile(filename, os.O_RDONLY, 0, 0)
}

// OpenFile memory-maps the named file for reading/writing, depending on
// the flag value.
func OpenFile(filename string, flag int, mode os.FileMode) (*MapFile, error) {
	return openFile(filename, flag, mode, int(pageSize))
}

// OpenFileS memory-maps the named file for reading/writing, depending on
// the flag value.
func OpenFileS(filename string, flag int, mode os.FileMode, size int) (*MapFile, error) {
	return openFile(filename, flag, mode, size)
}

func openFile(filename string, mode int, perm os.FileMode, size int) (*MapFile, error) {
	if len(filename) == 0 {
		return nil, syscall.ENOENT
	}

	f, err := os.OpenFile(filename, mode|os.O_CREATE, perm)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	fsize := fi.Size()
	prot := PROT_READ
	rdOnly := true
	switch mode & (os.O_RDONLY | os.O_WRONLY | os.O_RDWR) {
	case os.O_RDONLY:
		// fmt.Println("MapFile: read only")
		prot = PROT_READ
	case os.O_WRONLY:
		// fmt.Println("MapFile: write only")
		prot = PROT_WRITE
		rdOnly = false
	case os.O_RDWR:
		// fmt.Println("MapFile: read write")
		prot = PROT_READ | PROT_WRITE
		rdOnly = false
	}

	// fmt.Println("MapFile: file", filename, "mode", fmt.Sprintf("0x%x", mode), "modeis", mode&os.O_RDONLY, "read only", rdOnly)
	if fsize == 0 && rdOnly {
		// fmt.Println("MapFile: empty file", filename)
		return &MapFile{readOnly: rdOnly}, nil
	}
	if fsize < 0 {
		return nil, fmt.Errorf("MapFile: file %q has negative size", filename)
	}
	if fsize != int64(int(fsize)) {
		return nil, fmt.Errorf("MapFile: file %q is too large", filename)
	}

	if int(fsize) > size {
		size = int(fsize)
	}
	data, err := Mmap(int(f.Fd()), 0, size, prot, MAP_SHARED)
	if err != nil {
		return nil, err
	}
	fd := &MapFile{
		data:     data,
		fd:       f,
		readOnly: rdOnly,
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
