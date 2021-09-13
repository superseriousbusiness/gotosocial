package mutexes

import (
	"sync"
)

// MutexMap is a structure that allows having a map of self-evicting mutexes
// by key. You do not need to worry about managing the contents of the map,
// only requesting RLock/Lock for keys, and ensuring to call the returned
// unlock functions.
type MutexMap struct {
	// NOTE:
	// Individual keyed mutexes should ONLY ever
	// be locked within the protection of the outer
	// mapMu lock. If you lock these outside the
	// protection of this, there is a chance for
	// deadlocks

	mus   map[string]RWMutex
	mapMu sync.Mutex
	pool  sync.Pool
}

// NewMap returns a new MutexMap instance based on supplied
// RWMutex allocator function, nil implies use default
func NewMap(newFn func() RWMutex) MutexMap {
	if newFn == nil {
		newFn = NewRW
	}
	return MutexMap{
		mus:   make(map[string]RWMutex),
		mapMu: sync.Mutex{},
		pool: sync.Pool{
			New: func() interface{} {
				return newFn()
			},
		},
	}
}

func (mm *MutexMap) evict(key string, mu RWMutex) {
	// Acquire map lock
	mm.mapMu.Lock()

	// Toggle mutex lock to
	// ensure it is unused
	unlock := mu.Lock()
	unlock()

	// Delete mutex key
	delete(mm.mus, key)
	mm.mapMu.Unlock()

	// Release to pool
	mm.pool.Put(mu)
}

// RLock acquires a mutex read lock for supplied key, returning an RUnlock function
func (mm *MutexMap) RLock(key string) func() {
	return mm.getLock(key, func(mu RWMutex) func() {
		return mu.RLock()
	})
}

// Lock acquires a mutex lock for supplied key, returning an Unlock function
func (mm *MutexMap) Lock(key string) func() {
	return mm.getLock(key, func(mu RWMutex) func() {
		return mu.Lock()
	})
}

func (mm *MutexMap) getLock(key string, doLock func(RWMutex) func()) func() {
	// Get map lock
	mm.mapMu.Lock()

	// Look for mutex
	mu, ok := mm.mus[key]
	if ok {
		// Lock and return
		// its unlocker func
		unlock := doLock(mu)
		mm.mapMu.Unlock()
		return unlock
	}

	// Note: even though the mutex data structure is
	// small, benchmarking does actually show that pooled
	// alloc of mutexes here is faster

	// Acquire mu + add
	mu = mm.pool.Get().(RWMutex)
	mm.mus[key] = mu

	// Lock mutex + unlock map
	unlockFn := doLock(mu)
	mm.mapMu.Unlock()

	return func() {
		// Unlock mutex
		unlockFn()

		// Release function
		go mm.evict(key, mu)
	}
}
