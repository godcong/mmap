package mmap

import (
	"errors"
	"io"
	"os"
)

var (
	ErrBadFileDesc = errors.New("bad file descriptor")
	ErrShortWrite  = io.ErrShortWrite
	ErrInvalid     = os.ErrInvalid
	EOF            = io.EOF
)
