package mangler

import (
	"unsafe"
)

func append_uint16(b []byte, u uint16) []byte {
	return append(b, // LE
		byte(u),
		byte(u>>8),
	)
}

func append_uint32(b []byte, u uint32) []byte {
	return append(b, // LE
		byte(u),
		byte(u>>8),
		byte(u>>16),
		byte(u>>24),
	)
}

func append_uint64(b []byte, u uint64) []byte {
	return append(b, // LE
		byte(u),
		byte(u>>8),
		byte(u>>16),
		byte(u>>24),
		byte(u>>32),
		byte(u>>40),
		byte(u>>48),
		byte(u>>56),
	)
}

func empty_mangler(buf []byte, _ unsafe.Pointer) []byte {
	return buf
}

// add returns the ptr addition of starting ptr and a delta.
func add(ptr unsafe.Pointer, delta uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) + delta)
}
