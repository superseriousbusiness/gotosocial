package atomics

import (
	"sync/atomic"
	"unsafe"
)

// Interface provides user-friendly means of performing atomic operations on interface{} types.
type Interface struct{ ptr unsafe.Pointer }

// NewInterface will return a new Interface instance initialized with zero value.
func NewInterface() *Interface {
	var v interface{}
	return &Interface{
		ptr: unsafe.Pointer(&v),
	}
}

// Store will atomically store interface{} value in address contained within v.
func (v *Interface) Store(val interface{}) {
	atomic.StorePointer(&v.ptr, unsafe.Pointer(&val))
}

// Load will atomically load interface{} value at address contained within v.
func (v *Interface) Load() interface{} {
	return *(*interface{})(atomic.LoadPointer(&v.ptr))
}

// CAS performs a compare-and-swap for a(n) interface{} value at address contained within v.
func (v *Interface) CAS(cmp, swp interface{}) bool {
	for {
		// Load current value at address
		ptr := atomic.LoadPointer(&v.ptr)
		cur := *(*interface{})(ptr)

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

// Swap atomically stores new interface{} value into address contained within v, and returns previous value.
func (v *Interface) Swap(swp interface{}) interface{} {
	ptr := unsafe.Pointer(&swp)
	ptr = atomic.SwapPointer(&v.ptr, ptr)
	return *(*interface{})(ptr)
}
