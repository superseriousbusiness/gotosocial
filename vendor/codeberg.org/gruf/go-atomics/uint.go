package atomics

import "sync/atomic"

// Uint32 provides user-friendly means of performing atomic operations on uint32 types.
type Uint32 uint32

// NewUint32 will return a new Uint32 instance initialized with zero value.
func NewUint32() *Uint32 {
	return new(Uint32)
}

// Add will atomically add uint32 delta to value in address contained within i, returning new value.
func (u *Uint32) Add(delta uint32) uint32 {
	return atomic.AddUint32((*uint32)(u), delta)
}

// Store will atomically store uint32 value in address contained within i.
func (u *Uint32) Store(val uint32) {
	atomic.StoreUint32((*uint32)(u), val)
}

// Load will atomically load uint32 value at address contained within i.
func (u *Uint32) Load() uint32 {
	return atomic.LoadUint32((*uint32)(u))
}

// CAS performs a compare-and-swap for a(n) uint32 value at address contained within i.
func (u *Uint32) CAS(cmp, swp uint32) bool {
	return atomic.CompareAndSwapUint32((*uint32)(u), cmp, swp)
}

// Swap atomically stores new uint32 value into address contained within i, and returns previous value.
func (u *Uint32) Swap(swp uint32) uint32 {
	return atomic.SwapUint32((*uint32)(u), swp)
}

// Uint64 provides user-friendly means of performing atomic operations on uint64 types.
type Uint64 uint64

// NewUint64 will return a new Uint64 instance initialized with zero value.
func NewUint64() *Uint64 {
	return new(Uint64)
}

// Add will atomically add uint64 delta to value in address contained within i, returning new value.
func (u *Uint64) Add(delta uint64) uint64 {
	return atomic.AddUint64((*uint64)(u), delta)
}

// Store will atomically store uint64 value in address contained within i.
func (u *Uint64) Store(val uint64) {
	atomic.StoreUint64((*uint64)(u), val)
}

// Load will atomically load uint64 value at address contained within i.
func (u *Uint64) Load() uint64 {
	return atomic.LoadUint64((*uint64)(u))
}

// CAS performs a compare-and-swap for a(n) uint64 value at address contained within i.
func (u *Uint64) CAS(cmp, swp uint64) bool {
	return atomic.CompareAndSwapUint64((*uint64)(u), cmp, swp)
}

// Swap atomically stores new uint64 value into address contained within i, and returns previous value.
func (u *Uint64) Swap(swp uint64) uint64 {
	return atomic.SwapUint64((*uint64)(u), swp)
}
