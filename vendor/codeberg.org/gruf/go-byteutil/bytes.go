package byteutil

import (
	"reflect"
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
//
// According to the Go spec strings are immutable and byte slices are not. The way this gets implemented is strings under the hood are:
//
//	type StringHeader struct {
//		Data uintptr
//		Len  int
//	}
//
// while slices are:
//
//	type SliceHeader struct {
//		Data uintptr
//		Len  int
//		Cap  int
//	}
//
// because being mutable, you can change the data, length etc, but the string has to promise to be read-only to all who get copies of it.
//
// So in practice when you do a conversion of `string(byteSlice)` it actually performs an allocation because it has to copy the contents of the byte slice into a safe read-only state.
//
// Being that the shared fields are in the same struct indices (no different offsets), means that if you have a byte slice you can "forcibly" cast it to a string. Which in a lot of situations can be risky, because then it means you have a string that is NOT immutable, as if someone changes the data in the originating byte slice then the string will reflect that change! Now while this does seem hacky, and it _kind_ of is, it is something that you see performed in the standard library. If you look at the definition for `strings.Builder{}.String()` you'll see this :)
func B2S(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// S2B returns a []byte representation of string without allocation (minus slice header).
// See B2S() code comment, and this function's implementation for a better understanding.
func S2B(s string) []byte {
	var b []byte

	// Get byte + string headers
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))

	// Manually set bytes to string
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len

	return b
}
