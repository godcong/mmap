//go:build go1.22

package mmap

import (
	"math"
	"math/rand/v2"
)

// GenKey generates a random int id, not including 0.
func GenKey() int {
	return rand.IntN(math.MaxInt-1) + 1
}
