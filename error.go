package mmap

import (
	"errors"
	"io"
	"os"
)

var (
	ErrBadFileDesc = errors.New("bad file descriptor")
	ErrClosed      = errors.New("file/map already closed")
	ErrShortWrite  = io.ErrShortWrite
	ErrInvalid     = os.ErrInvalid
	EOF            = io.EOF
)
