package byteutil

import (
	"unsafe"
)

// Copy returns a copy of []byte.
func Copy(b []byte) []byte {
	if b == nil {
		return nil
	}
	p := make([]byte, len(b))
	copy(p, b)
	return p
}

// B2S returns a string representation of []byte without allocation.
// Since Go strings are immutable, the bytes passed to String must
// not be modified as long as the returned string value exists.
func B2S(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// S2B returns a []byte representation of string without allocation.
func S2B(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
