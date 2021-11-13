package mutexes

import "sync"

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
	once := sync.Once{}
	return func() { once.Do(unlock) }
}

// safeRWMutex simply wraps a RWMutex to add multi-unlock safety
type safeRWMutex struct{ mu RWMutex }

func (mu *safeRWMutex) Lock() func() {
	unlock := mu.mu.Lock()
	once := sync.Once{}
	return func() { once.Do(unlock) }
}

func (mu *safeRWMutex) RLock() func() {
	unlock := mu.mu.RLock()
	once := sync.Once{}
	return func() { once.Do(unlock) }
}
