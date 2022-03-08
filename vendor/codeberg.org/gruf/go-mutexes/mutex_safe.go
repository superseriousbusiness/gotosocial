package mutexes

import (
	"sync/atomic"
)

// WithSafety wrapps the supplied Mutex to protect unlock fns
// from being called multiple times
func WithSafety(mu Mutex) Mutex {
	return &safeMutex{mu: mu}
}

// WithSafetyRW wrapps the supplied RWMutex to protect unlock
// fns from being called multiple times
func WithSafetyRW(mu RWMutex) RWMutex {
	return &safeRWMutex{mu: mu}
}

// safeMutex simply wraps a Mutex to add multi-unlock safety
type safeMutex struct{ mu Mutex }

func (mu *safeMutex) Lock() func() {
	unlock := mu.mu.Lock()
	return once(unlock)
}

// safeRWMutex simply wraps a RWMutex to add multi-unlock safety
type safeRWMutex struct{ mu RWMutex }

func (mu *safeRWMutex) Lock() func() {
	unlock := mu.mu.Lock()
	return once(unlock)
}

func (mu *safeRWMutex) RLock() func() {
	unlock := mu.mu.RLock()
	return once(unlock)
}

// once will perform 'do' only once, this is safe for unlocks
// as 2 functions calling 'unlock()' don't need absolute guarantees
// that by the time it is completed the unlock was finished.
func once(do func()) func() {
	var done uint32
	return func() {
		if atomic.CompareAndSwapUint32(&done, 0, 1) {
			do()
		}
	}
}
