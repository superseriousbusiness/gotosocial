package atomics

import "sync/atomic"

// Int32 provides user-friendly means of performing atomic operations on int32 types.
type Int32 int32

// NewInt32 will return a new Int32 instance initialized with zero value.
func NewInt32() *Int32 {
	return new(Int32)
}

// Add will atomically add int32 delta to value in address contained within i, returning new value.
func (i *Int32) Add(delta int32) int32 {
	return atomic.AddInt32((*int32)(i), delta)
}

// Store will atomically store int32 value in address contained within i.
func (i *Int32) Store(val int32) {
	atomic.StoreInt32((*int32)(i), val)
}

// Load will atomically load int32 value at address contained within i.
func (i *Int32) Load() int32 {
	return atomic.LoadInt32((*int32)(i))
}

// CAS performs a compare-and-swap for a(n) int32 value at address contained within i.
func (i *Int32) CAS(cmp, swp int32) bool {
	return atomic.CompareAndSwapInt32((*int32)(i), cmp, swp)
}

// Swap atomically stores new int32 value into address contained within i, and returns previous value.
func (i *Int32) Swap(swp int32) int32 {
	return atomic.SwapInt32((*int32)(i), swp)
}

// Int64 provides user-friendly means of performing atomic operations on int64 types.
type Int64 int64

// NewInt64 will return a new Int64 instance initialized with zero value.
func NewInt64() *Int64 {
	return new(Int64)
}

// Add will atomically add int64 delta to value in address contained within i, returning new value.
func (i *Int64) Add(delta int64) int64 {
	return atomic.AddInt64((*int64)(i), delta)
}

// Store will atomically store int64 value in address contained within i.
func (i *Int64) Store(val int64) {
	atomic.StoreInt64((*int64)(i), val)
}

// Load will atomically load int64 value at address contained within i.
func (i *Int64) Load() int64 {
	return atomic.LoadInt64((*int64)(i))
}

// CAS performs a compare-and-swap for a(n) int64 value at address contained within i.
func (i *Int64) CAS(cmp, swp int64) bool {
	return atomic.CompareAndSwapInt64((*int64)(i), cmp, swp)
}

// Swap atomically stores new int64 value into address contained within i, and returns previous value.
func (i *Int64) Swap(swp int64) int64 {
	return atomic.SwapInt64((*int64)(i), swp)
}
