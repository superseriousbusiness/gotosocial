package mutexes

import (
	"sync"
)

// Mutex defines a wrappable mutex. By forcing unlocks
// via returned function it makes wrapping much easier
type Mutex interface {
	// Lock performs a mutex lock, returning an unlock function
	Lock() (unlock func())
}

// RWMutex defines a wrappable read-write mutex. By forcing
// unlocks via returned functions it makes wrapping much easier
type RWMutex interface {
	Mutex

	// RLock performs a mutex read lock, returning an unlock function
	RLock() (runlock func())
}

// New returns a new base Mutex implementation
func New() Mutex {
	return &baseMutex{}
}

// NewRW returns a new base RWMutex implementation
func NewRW() RWMutex {
	return &baseRWMutex{}
}

// WithFunc wraps the supplied Mutex to call the provided hooks on lock / unlock
func WithFunc(mu Mutex, onLock, onUnlock func()) Mutex {
	return &fnMutex{mu: mu, lo: onLock, un: onUnlock}
}

// WithFuncRW wrapps the supplied RWMutex to call the provided hooks on lock / rlock / unlock/ runlock
func WithFuncRW(mu RWMutex, onLock, onRLock, onUnlock, onRUnlock func()) RWMutex {
	return &fnRWMutex{mu: mu, lo: onLock, rlo: onRLock, un: onUnlock, run: onRUnlock}
}

// baseMutex simply wraps a sync.Mutex to implement our Mutex interface
type baseMutex struct{ mu sync.Mutex }

func (mu *baseMutex) Lock() func() {
	mu.mu.Lock()
	return mu.mu.Unlock
}

// baseRWMutex simply wraps a sync.RWMutex to implement our RWMutex interface
type baseRWMutex struct{ mu sync.RWMutex }

func (mu *baseRWMutex) Lock() func() {
	mu.mu.Lock()
	return mu.mu.Unlock
}

func (mu *baseRWMutex) RLock() func() {
	mu.mu.RLock()
	return mu.mu.RUnlock
}

// fnMutex wraps a Mutex to add hooks for Lock and Unlock
type fnMutex struct {
	mu Mutex
	lo func()
	un func()
}

func (mu *fnMutex) Lock() func() {
	unlock := mu.mu.Lock()
	mu.lo()
	return func() {
		mu.un()
		unlock()
	}
}

// fnRWMutex wraps a RWMutex to add hooks for Lock, RLock, Unlock and RUnlock
type fnRWMutex struct {
	mu  RWMutex
	lo  func()
	rlo func()
	un  func()
	run func()
}

func (mu *fnRWMutex) Lock() func() {
	unlock := mu.mu.Lock()
	mu.lo()
	return func() {
		mu.un()
		unlock()
	}
}

func (mu *fnRWMutex) RLock() func() {
	unlock := mu.mu.RLock()
	mu.rlo()
	return func() {
		mu.run()
		unlock()
	}
}
