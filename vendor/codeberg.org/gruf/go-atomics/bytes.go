package atomics

import (
	"sync/atomic"
	"unsafe"
)

// Bytes provides user-friendly means of performing atomic operations on []byte types.
type Bytes struct{ ptr unsafe.Pointer }

// NewBytes will return a new Bytes instance initialized with zero value.
func NewBytes() *Bytes {
	var v []byte
	return &Bytes{
		ptr: unsafe.Pointer(&v),
	}
}

// Store will atomically store []byte value in address contained within v.
func (v *Bytes) Store(val []byte) {
	atomic.StorePointer(&v.ptr, unsafe.Pointer(&val))
}

// Load will atomically load []byte value at address contained within v.
func (v *Bytes) Load() []byte {
	return *(*[]byte)(atomic.LoadPointer(&v.ptr))
}

// CAS performs a compare-and-swap for a(n) []byte value at address contained within v.
func (v *Bytes) CAS(cmp, swp []byte) bool {
	for {
		// Load current value at address
		ptr := atomic.LoadPointer(&v.ptr)
		cur := *(*[]byte)(ptr)

		// Perform comparison against current
		if !(string(cur) == string(cmp)) {
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

// Swap atomically stores new []byte value into address contained within v, and returns previous value.
func (v *Bytes) Swap(swp []byte) []byte {
	ptr := unsafe.Pointer(&swp)
	ptr = atomic.SwapPointer(&v.ptr, ptr)
	return *(*[]byte)(ptr)
}
