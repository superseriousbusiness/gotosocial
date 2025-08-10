//go:build go1.24 && !go1.26

package format

import (
	"reflect"
	"unsafe"

	"codeberg.org/gruf/go-xunsafe"
)

// add returns the ptr addition of starting ptr and a delta.
func add(ptr unsafe.Pointer, delta uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) + delta)
}

// typeof is short-hand for reflect.TypeFor[T]().
func typeof[T any]() reflect.Type {
	return reflect.TypeFor[T]()
}

const (
	// custom xunsafe.Reflect_flag to indicate key types.
	flagKeyType xunsafe.Reflect_flag = 1 << 10
)
