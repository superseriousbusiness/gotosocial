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
// type StringHeader struct {
// 	Data uintptr
// 	Len  int
// }
//
// while slices are:
// type SliceHeader struct {
// 	Data uintptr
// 	Len  int
// 	Cap  int
// }
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

// ToUpper offers a faster ToUpper implementation using a lookup table.
func ToUpper(b []byte) {
	const toUpperTable = "\x00\x01\x02\x03\x04\x05\x06\a\b\t\n\v\f\r\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f" +
		" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`ABCDEFGHIJKLMNOPQRSTUVWXYZ{|}~" +
		"\u007f\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96" +
		"\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf" +
		"\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7\xc8" +
		"\xc9\xca\xcb\xcc\xcd\xce\xcf\xd0\xd1\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xdb\xdc\xdd\xde\xdf\xe0\xe1" +
		"\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xeb\xec\xed\xee\xef\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"
	for i := 0; i < len(b); i++ {
		b[i] = toUpperTable[b[i]]
	}
}

// ToLower offers a faster ToLower implementation using a lookup table.
func ToLower(b []byte) {
	const toLowerTable = "\x00\x01\x02\x03\x04\x05\x06\a\b\t\n\v\f\r\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f" +
		" !\"#$%&'()*+,-./0123456789:;<=>?@abcdefghijklmnopqrstuvwxyz[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~" +
		"\u007f\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96" +
		"\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf" +
		"\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7\xc8" +
		"\xc9\xca\xcb\xcc\xcd\xce\xcf\xd0\xd1\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xdb\xdc\xdd\xde\xdf\xe0\xe1" +
		"\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xeb\xec\xed\xee\xef\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"
	for i := 0; i < len(b); i++ {
		b[i] = toLowerTable[b[i]]
	}
}
