package atomics

import "sync/atomic"

// Bool provides user-friendly means of performing atomic operations on bool types.
type Bool uint32

// NewBool will return a new Bool instance initialized with zero value.
func NewBool() *Bool {
	return new(Bool)
}

// Store will atomically store bool value in address contained within i.
func (b *Bool) Store(val bool) {
	atomic.StoreUint32((*uint32)(b), fromBool(val))
}

// Load will atomically load bool value at address contained within i.
func (b *Bool) Load() bool {
	return toBool(atomic.LoadUint32((*uint32)(b)))
}

// CAS performs a compare-and-swap for a(n) bool value at address contained within i.
func (b *Bool) CAS(cmp, swp bool) bool {
	return atomic.CompareAndSwapUint32((*uint32)(b), fromBool(cmp), fromBool(swp))
}

// Swap atomically stores new bool value into address contained within i, and returns previous value.
func (b *Bool) Swap(swp bool) bool {
	return toBool(atomic.SwapUint32((*uint32)(b), fromBool(swp)))
}

// toBool converts uint32 value to bool.
func toBool(u uint32) bool {
	if u == 0 {
		return false
	}
	return true
}

// fromBool converts from bool to uint32 value.
func fromBool(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}
