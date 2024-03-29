package mmap

import (
	"fmt"
	"io"
	"os"
)

const (
	// MapMemKeyInvalid share memory id will return 0,so we set the invalid id to -1
	MapMemKeyInvalid = -1
)

type MapMem struct {
	owner bool
	id    int
	data  []byte
	off   int
	close func() error
}

var pageSize int

func init() {
	pageSize = os.Getpagesize()
}

func (f *MapMem) Seek(offset int64, whence int) (int64, error) {
	// Log().Debug("MapMem.Seek sta", "len", len(f.data), "off", f.off)
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
		return 0, fmt.Errorf("MapMem: invalid whence")
	}
	if f.off < 0 {
		return 0, fmt.Errorf("MapMem: negative position")
	}
	// Log().Debug("MapMem.Seek end", "len", len(f.data), "off", f.off)
	return int64(f.off), nil
}

func (f *MapMem) WriteByte(c byte) error {
	// Log().Debug("MapMem.WriteByte sta", "len", len(f.data), "off", f.off)
	if f == nil {
		return ErrInvalid
	}

	if !f.owner {
		return ErrBadFileDesc
	}
	if f.off >= len(f.data) {
		// Log().Error("MapMem.WriteByte error", "err", ErrShortWrite, "len", len(f.data), "off", f.off)
		return ErrShortWrite
	}

	f.data[f.off] = c
	f.off++
	// Log().Debug("MapMem.WriteByte end", "len", len(f.data), "off", f.off)
	return nil
}

func (f *MapMem) WriteAt(p []byte, off int64) (n int, err error) {
	// Log().Debug("MapMem.WriteAt sta", "len", len(f.data), "off", f.off)
	if f == nil {
		return 0, ErrInvalid
	}

	if !f.owner {
		return 0, ErrBadFileDesc
	}
	if f.data == nil {
		return 0, fmt.Errorf("MapMem: %w", ErrClosed)
	}
	if off < 0 || int64(len(f.data)) < off {
		err := fmt.Errorf("MapMem: invalid WriteAt offset %d", off)
		// Log().Error("MapMem.WriteAt error", "err", err, "len", len(f.data), "off", f.off)
		return 0, err
	}
	n = copy(f.data[off:], p)
	if n < len(p) {
		// Log().Error("MapMem.WriteAt error", "err", ErrShortWrite, "len", len(f.data), "off", f.off)
		return n, ErrShortWrite
	}
	// Log().Debug("MapMem.WriteAt end", "len", len(f.data), "off", f.off)
	return n, nil
}

func (f *MapMem) ReadByte() (byte, error) {
	// Log().Debug("MapMem.ReadByte sta", "len", len(f.data), "off", f.off)
	if f == nil {
		return 0, ErrInvalid
	}

	if f.off >= len(f.data) {
		return 0, EOF
	}
	v := f.data[f.off]
	f.off++
	// Log().Debug("MapMem.ReadByte end", "len", len(f.data), "off", f.off)
	return v, nil
}

func (f *MapMem) ReadAt(p []byte, off int64) (n int, err error) {
	// Log().Debug("MapMem.ReadByte sta", "len", len(f.data), "off", f.off)
	if f == nil {
		return 0, ErrInvalid
	}

	if f.data == nil {
		// Log().Error("MapMem.ReadAt error", "err", ErrClosed, "len", len(f.data), "off", f.off)
		return 0, fmt.Errorf("MapMem: %v", ErrClosed)
	}
	if off < 0 || int64(len(f.data)) < off {
		err := fmt.Errorf("MapMem: invalid ReadAt offset %d", off)
		// Log().Error("MapMem.ReadAt error", "err", err, "len", len(f.data), "off", f.off)
		return 0, err
	}
	n = copy(p, f.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	// Log().Debug("MapMem.ReadByte end", "len", len(f.data), "off", f.off)
	return n, nil
}

// Read implements the io.Reader interface.
func (f *MapMem) Read(p []byte) (int, error) {
	// Log().Debug("MapMem.Read sta", "len", len(f.data), "off", f.off)
	if f == nil {
		return 0, ErrInvalid
	}

	if f.off >= len(f.data) {
		return 0, EOF
	}
	n := copy(p, f.data[f.off:])
	f.off += n
	// Log().Debug("MapMem.Read end", "len", len(f.data), "off", f.off)
	return n, nil
}

// Write implements the io.Writer interface.
func (f *MapMem) Write(p []byte) (int, error) {
	// Log().Debug("MapMem.Write sta", "len", len(f.data), "off", f.off)
	if f == nil {
		return 0, ErrInvalid
	}

	if !f.owner {
		return 0, ErrBadFileDesc
	}
	if f.off >= len(f.data) {
		// Log().Error("MapMem.Write error", "err", ErrShortWrite, "len", len(f.data), "off", f.off)
		return 0, ErrShortWrite
	}
	n := copy(f.data[f.off:], p)
	f.off += n
	if len(p) > n {
		// Log().Error("MapMem.Write written error", "err", ErrShortWrite, "len", len(f.data), "off", f.off)
		return n, ErrShortWrite
	}
	// Log().Debug("MapMem.Write end", "len", len(f.data), "off", f.off)
	return n, nil
}

func (f *MapMem) ID() int {
	return f.id
}

func (f *MapMem) IsOwner() bool {
	return f.owner
}

func (f *MapMem) Len() int {
	return len(f.data)
}

func (f *MapMem) Cap() int {
	return cap(f.data)
}

func OpenMem(id int, size int) (*MapMem, error) {
	return openMapMem(id, size)
}

func OpenMemS(id int) (*MapMem, error) {
	return openMapMem(id, 0)
}

func getPageSize(size int) int {
	if size == 0 {
		return pageSize
	}
	return size
}

var (
	_ io.Reader     = (*MapMem)(nil)
	_ io.ReaderAt   = (*MapMem)(nil)
	_ io.ByteReader = (*MapMem)(nil)
	_ io.Writer     = (*MapMem)(nil)
	_ io.WriterAt   = (*MapMem)(nil)
	_ io.ByteWriter = (*MapMem)(nil)
	_ io.Closer     = (*MapMem)(nil)
	_ io.Seeker     = (*MapMem)(nil)
)
