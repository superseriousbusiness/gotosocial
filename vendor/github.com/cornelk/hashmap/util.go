package hashmap

import (
	"strconv"
)

const (
	// intSizeBytes is the size in byte of an int or uint value.
	intSizeBytes = strconv.IntSize >> 3
)

// roundUpPower2 rounds a number to the next power of 2.
func roundUpPower2(i uintptr) uintptr {
	i--
	i |= i >> 1
	i |= i >> 2
	i |= i >> 4
	i |= i >> 8
	i |= i >> 16
	i |= i >> 32
	i++
	return i
}

// log2 computes the binary logarithm of x, rounded up to the next integer.
func log2(i uintptr) uintptr {
	var n, p uintptr
	for p = 1; p < i; p += p {
		n++
	}
	return n
}
