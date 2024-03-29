//go:build !go1.22

package mmap

import (
	"math"
	"math/rand"
	"time"
)

var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// GenKey generates a random int id, not including 0.
func GenKey() int {
	return r.Intn(math.MaxInt-1) + 1
}
