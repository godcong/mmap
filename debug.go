package mmap

import "os"

var debug = true

func init() {
	if os.Getenv("GO_MMAP_DEBUG") != "" {
		debug = true
	}
}
