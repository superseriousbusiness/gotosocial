package atomics

import (
	"sync/atomic"
	"unsafe"
)

// Error provides user-friendly means of performing atomic operations on error types.
type Error struct{ ptr unsafe.Pointer }

// NewError will return a new Error instance initialized with zero value.
func NewError() *Error {
	var v error
	return &Error{
		ptr: unsafe.Pointer(&v),
	}
}

// Store will atomically store error value in address contained within v.
func (v *Error) Store(val error) {
	atomic.StorePointer(&v.ptr, unsafe.Pointer(&val))
}

// Load will atomically load error value at address contained within v.
func (v *Error) Load() error {
	return *(*error)(atomic.LoadPointer(&v.ptr))
}

// CAS performs a compare-and-swap for a(n) error value at address contained within v.
func (v *Error) CAS(cmp, swp error) bool {
	for {
		// Load current value at address
		ptr := atomic.LoadPointer(&v.ptr)
		cur := *(*error)(ptr)

		// Perform comparison against current
		if !(cur == cmp) {
			return false
		}

		// Attempt to replace pointer
		if atomic.CompareAndSwapPointer(
			&v.ptr,
			ptr,
			unsafe.Pointer(&swp),
		) {
			return true
		}
	}
}

// Swap atomically stores new error value into address contained within v, and returns previous value.
func (v *Error) Swap(swp error) error {
	ptr := unsafe.Pointer(&swp)
	ptr = atomic.SwapPointer(&v.ptr, ptr)
	return *(*error)(ptr)
}
