package mmap

import (
	"os"
)

var debug = false

func init() {
	if os.Getenv("GO_MMAP_DEBUG") != "" {
		debug = true
	}
}
