package mutexes

import (
	"sync"
	"time"

	"git.iim.gay/grufwub/go-nowish"
)

// TimeoutMutex defines a Mutex with timeouts on locks
type TimeoutMutex interface {
	Mutex

	// LockFunc is functionally the same as Lock(), but allows setting a custom hook called on timeout
	LockFunc(func()) func()
}

// TimeoutRWMutex defines a RWMutex with timeouts on locks
type TimeoutRWMutex interface {
	RWMutex

	// LockFunc is functionally the same as Lock(), but allows setting a custom hook called on timeout
	LockFunc(func()) func()

	// RLockFunc is functionally the same as RLock(), but allows setting a custom hook called on timeout
	RLockFunc(func()) func()
}

// WithTimeout wraps the supplied Mutex to add a timeout
func WithTimeout(mu Mutex, d time.Duration) TimeoutMutex {
	return &timeoutMutex{mu: mu, d: d}
}

// WithTimeoutRW wraps the supplied RWMutex to add read/write timeouts
func WithTimeoutRW(mu RWMutex, rd, wd time.Duration) TimeoutRWMutex {
	return &timeoutRWMutex{mu: mu, rd: rd, wd: wd}
}

// timeoutMutex wraps a Mutex with timeout
type timeoutMutex struct {
	mu Mutex         // mu is the wrapped mutex
	d  time.Duration // d is the timeout duration
}

func (mu *timeoutMutex) Lock() func() {
	return mu.LockFunc(func() { panic("timed out") })
}

func (mu *timeoutMutex) LockFunc(fn func()) func() {
	return mutexTimeout(mu.d, mu.mu.Lock(), fn)
}

// TimeoutRWMutex wraps a RWMutex with timeouts
type timeoutRWMutex struct {
	mu RWMutex       // mu is the wrapped rwmutex
	rd time.Duration // rd is the rlock timeout duration
	wd time.Duration // wd is the lock timeout duration
}

func (mu *timeoutRWMutex) Lock() func() {
	return mu.LockFunc(func() { panic("timed out") })
}

func (mu *timeoutRWMutex) LockFunc(fn func()) func() {
	return mutexTimeout(mu.wd, mu.mu.Lock(), fn)
}

func (mu *timeoutRWMutex) RLock() func() {
	return mu.RLockFunc(func() { panic("timed out") })
}

func (mu *timeoutRWMutex) RLockFunc(fn func()) func() {
	return mutexTimeout(mu.rd, mu.mu.RLock(), fn)
}

// timeoutPool provides nowish.Timeout objects for timeout mutexes
var timeoutPool = sync.Pool{
	New: func() interface{} {
		return nowish.NewTimeout()
	},
}

// mutexTimeout performs a timed unlock, calling supplied fn if timeout is reached
func mutexTimeout(d time.Duration, unlock func(), fn func()) func() {
	if d < 1 {
		// No timeout, just unlock
		return unlock
	}

	// Acquire timeout obj
	t := timeoutPool.Get().(nowish.Timeout)

	// Start the timeout with hook
	t.Start(d, fn)

	// Return func cancelling timeout,
	// replacing Timeout in pool and
	// finally unlocking mutex
	return func() {
		t.Cancel()
		timeoutPool.Put(t)
		unlock()
	}
}
