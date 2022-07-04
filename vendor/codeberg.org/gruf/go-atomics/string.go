package atomics

import (
	"sync/atomic"
	"unsafe"
)

// String provides user-friendly means of performing atomic operations on string types.
type String struct{ ptr unsafe.Pointer }

// NewString will return a new String instance initialized with zero value.
func NewString() *String {
	var v string
	return &String{
		ptr: unsafe.Pointer(&v),
	}
}

// Store will atomically store string value in address contained within v.
func (v *String) Store(val string) {
	atomic.StorePointer(&v.ptr, unsafe.Pointer(&val))
}

// Load will atomically load string value at address contained within v.
func (v *String) Load() string {
	return *(*string)(atomic.LoadPointer(&v.ptr))
}

// CAS performs a compare-and-swap for a(n) string value at address contained within v.
func (v *String) CAS(cmp, swp string) bool {
	for {
		// Load current value at address
		ptr := atomic.LoadPointer(&v.ptr)
		cur := *(*string)(ptr)

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

// Swap atomically stores new string value into address contained within v, and returns previous value.
func (v *String) Swap(swp string) string {
	ptr := unsafe.Pointer(&swp)
	ptr = atomic.SwapPointer(&v.ptr, ptr)
	return *(*string)(ptr)
}
