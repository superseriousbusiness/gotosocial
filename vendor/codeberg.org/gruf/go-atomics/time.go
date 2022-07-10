package atomics

import (
	"sync/atomic"
	"time"
	"unsafe"
)

// Time provides user-friendly means of performing atomic operations on time.Time types.
type Time struct{ ptr unsafe.Pointer }

// NewTime will return a new Time instance initialized with zero value.
func NewTime() *Time {
	var v time.Time
	return &Time{
		ptr: unsafe.Pointer(&v),
	}
}

// Store will atomically store time.Time value in address contained within v.
func (v *Time) Store(val time.Time) {
	atomic.StorePointer(&v.ptr, unsafe.Pointer(&val))
}

// Load will atomically load time.Time value at address contained within v.
func (v *Time) Load() time.Time {
	return *(*time.Time)(atomic.LoadPointer(&v.ptr))
}

// CAS performs a compare-and-swap for a(n) time.Time value at address contained within v.
func (v *Time) CAS(cmp, swp time.Time) bool {
	for {
		// Load current value at address
		ptr := atomic.LoadPointer(&v.ptr)
		cur := *(*time.Time)(ptr)

		// Perform comparison against current
		if !(cur.Equal(cmp)) {
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

// Swap atomically stores new time.Time value into address contained within v, and returns previous value.
func (v *Time) Swap(swp time.Time) time.Time {
	ptr := unsafe.Pointer(&swp)
	ptr = atomic.SwapPointer(&v.ptr, ptr)
	return *(*time.Time)(ptr)
}
