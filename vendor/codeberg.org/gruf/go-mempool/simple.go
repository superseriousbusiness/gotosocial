package mempool

import (
	"unsafe"
)

// SimplePool provides a type-safe form
// of UnsafePool using generics.
//
// Note it is NOT safe for concurrent
// use, you must protect it yourself!
type SimplePool[T any] struct {
	UnsafeSimplePool

	// New is an optionally provided
	// allocator used when no value
	// is available for use in pool.
	New func() T

	// Reset is an optionally provided
	// value resetting function called
	// on passed value to Put().
	Reset func(T) bool
}

func (p *SimplePool[T]) Get() T {
	if ptr := p.UnsafeSimplePool.Get(); ptr != nil {
		return *(*T)(ptr)
	}
	var t T
	if p.New != nil {
		t = p.New()
	}
	return t
}

func (p *SimplePool[T]) Put(t T) {
	if p.Reset != nil && !p.Reset(t) {
		return
	}
	ptr := unsafe.Pointer(&t)
	p.UnsafeSimplePool.Put(ptr)
}

// UnsafeSimplePool provides an incredibly
// simple memory pool implementation
// that stores ptrs to memory values,
// and regularly flushes internal pool
// structures according to CheckGC().
//
// Note it is NOT safe for concurrent
// use, you must protect it yourself!
type UnsafeSimplePool struct {

	// Check determines how often to flush
	// internal pools based on underlying
	// current and victim pool sizes. It gets
	// called on every pool Put() operation.
	//
	// A flush will start a new current
	// pool, make victim the old current,
	// and drop the existing victim pool.
	Check func(current, victim int) bool

	current []unsafe.Pointer
	victim  []unsafe.Pointer
}

func (p *UnsafeSimplePool) Get() unsafe.Pointer {
	// First try current list.
	if len(p.current) > 0 {
		ptr := p.current[len(p.current)-1]
		p.current = p.current[:len(p.current)-1]
		return ptr
	}

	// Fallback to victim.
	if len(p.victim) > 0 {
		ptr := p.victim[len(p.victim)-1]
		p.victim = p.victim[:len(p.victim)-1]
		return ptr
	}

	return nil
}

func (p *UnsafeSimplePool) Put(ptr unsafe.Pointer) {
	p.current = append(p.current, ptr)

	// Get GC check func.
	if p.Check == nil {
		p.Check = defaultCheck
	}

	if p.Check(len(p.current), len(p.victim)) {
		p.GC() // garbage collection time!
	}
}

func (p *UnsafeSimplePool) GC() {
	p.victim = p.current
	p.current = nil
}

func (p *UnsafeSimplePool) Size() int {
	return len(p.current) + len(p.victim)
}

func defaultCheck(current, victim int) bool {
	return current-victim > 128 || victim > 256
}
