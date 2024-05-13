package mempool

import (
	"unsafe"
)

const DefaultDirtyFactor = 128

// Pool provides a type-safe form
// of UnsafePool using generics.
//
// Note it is NOT safe for concurrent
// use, you must protect it yourself!
type Pool[T any] struct {

	// New is an optionally provided
	// allocator used when no value
	// is available for use in pool.
	New func() T

	// Reset is an optionally provided
	// value resetting function called
	// on passed value to Put().
	Reset func(T)

	UnsafePool
}

func (p *Pool[T]) Get() T {
	if ptr := p.UnsafePool.Get(); ptr != nil {
		return *(*T)(ptr)
	} else if p.New != nil {
		return p.New()
	}
	var z T
	return z
}

func (p *Pool[T]) Put(t T) {
	if p.Reset != nil {
		p.Reset(t)
	}
	ptr := unsafe.Pointer(&t)
	p.UnsafePool.Put(ptr)
}

// UnsafePool provides an incredibly
// simple memory pool implementation
// that stores ptrs to memory values,
// and regularly flushes internal pool
// structures according to DirtyFactor.
//
// Note it is NOT safe for concurrent
// use, you must protect it yourself!
type UnsafePool struct {

	// DirtyFactor determines the max
	// number of $dirty count before
	// pool is garbage collected. Where:
	// $dirty = len(current) - len(victim)
	DirtyFactor int

	current []unsafe.Pointer
	victim  []unsafe.Pointer
}

func (p *UnsafePool) Get() unsafe.Pointer {
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

func (p *UnsafePool) Put(ptr unsafe.Pointer) {
	p.current = append(p.current, ptr)

	// Get dirty factor.
	df := p.DirtyFactor
	if df == 0 {
		df = DefaultDirtyFactor
	}

	if len(p.current)-len(p.victim) > df {
		// Garbage collection!
		p.victim = p.current
		p.current = nil
	}
}
