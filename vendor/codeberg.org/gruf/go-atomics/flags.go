package atomics

import (
	"sync/atomic"

	"codeberg.org/gruf/go-bitutil"
)

// Flags32 provides user-friendly means of performing atomic operations on bitutil.Flags32 types.
type Flags32 bitutil.Flags32

// NewFlags32 will return a new Flags32 instance initialized with zero value.
func NewFlags32() *Flags32 {
	return new(Flags32)
}

// Get will atomically load a(n) bitutil.Flags32 value contained within f, and check if bit value is set.
func (f *Flags32) Get(bit uint8) bool {
	return f.Load().Get(bit)
}

// Set performs a compare-and-swap for a(n) bitutil.Flags32 with bit value set, at address contained within f.
func (f *Flags32) Set(bit uint8) bool {
	cur := f.Load()
	return f.CAS(cur, cur.Set(bit))
}

// Unset performs a compare-and-swap for a(n) bitutil.Flags32 with bit value unset, at address contained within f.
func (f *Flags32) Unset(bit uint8) bool {
	cur := f.Load()
	return f.CAS(cur, cur.Unset(bit))
}

// Store will atomically store bitutil.Flags32 value in address contained within f.
func (f *Flags32) Store(val bitutil.Flags32) {
	atomic.StoreUint32((*uint32)(f), uint32(val))
}

// Load will atomically load bitutil.Flags32 value at address contained within f.
func (f *Flags32) Load() bitutil.Flags32 {
	return bitutil.Flags32(atomic.LoadUint32((*uint32)(f)))
}

// CAS performs a compare-and-swap for a(n) bitutil.Flags32 value at address contained within f.
func (f *Flags32) CAS(cmp, swp bitutil.Flags32) bool {
	return atomic.CompareAndSwapUint32((*uint32)(f), uint32(cmp), uint32(swp))
}

// Swap atomically stores new bitutil.Flags32 value into address contained within f, and returns previous value.
func (f *Flags32) Swap(swp bitutil.Flags32) bitutil.Flags32 {
	return bitutil.Flags32(atomic.SwapUint32((*uint32)(f), uint32(swp)))
}

// Flags64 provides user-friendly means of performing atomic operations on bitutil.Flags64 types.
type Flags64 bitutil.Flags64

// NewFlags64 will return a new Flags64 instance initialized with zero value.
func NewFlags64() *Flags64 {
	return new(Flags64)
}

// Get will atomically load a(n) bitutil.Flags64 value contained within f, and check if bit value is set.
func (f *Flags64) Get(bit uint8) bool {
	return f.Load().Get(bit)
}

// Set performs a compare-and-swap for a(n) bitutil.Flags64 with bit value set, at address contained within f.
func (f *Flags64) Set(bit uint8) bool {
	cur := f.Load()
	return f.CAS(cur, cur.Set(bit))
}

// Unset performs a compare-and-swap for a(n) bitutil.Flags64 with bit value unset, at address contained within f.
func (f *Flags64) Unset(bit uint8) bool {
	cur := f.Load()
	return f.CAS(cur, cur.Unset(bit))
}

// Store will atomically store bitutil.Flags64 value in address contained within f.
func (f *Flags64) Store(val bitutil.Flags64) {
	atomic.StoreUint64((*uint64)(f), uint64(val))
}

// Load will atomically load bitutil.Flags64 value at address contained within f.
func (f *Flags64) Load() bitutil.Flags64 {
	return bitutil.Flags64(atomic.LoadUint64((*uint64)(f)))
}

// CAS performs a compare-and-swap for a(n) bitutil.Flags64 value at address contained within f.
func (f *Flags64) CAS(cmp, swp bitutil.Flags64) bool {
	return atomic.CompareAndSwapUint64((*uint64)(f), uint64(cmp), uint64(swp))
}

// Swap atomically stores new bitutil.Flags64 value into address contained within f, and returns previous value.
func (f *Flags64) Swap(swp bitutil.Flags64) bitutil.Flags64 {
	return bitutil.Flags64(atomic.SwapUint64((*uint64)(f), uint64(swp)))
}
