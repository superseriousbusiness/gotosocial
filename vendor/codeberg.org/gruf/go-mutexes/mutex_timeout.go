package mutexes

import (
	"sync"
	"time"
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
	return mu.LockFunc(func() { panic("lock timed out") })
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
	return mu.LockFunc(func() { panic("lock timed out") })
}

func (mu *timeoutRWMutex) LockFunc(fn func()) func() {
	return mutexTimeout(mu.wd, mu.mu.Lock(), fn)
}

func (mu *timeoutRWMutex) RLock() func() {
	return mu.RLockFunc(func() { panic("rlock timed out") })
}

func (mu *timeoutRWMutex) RLockFunc(fn func()) func() {
	return mutexTimeout(mu.rd, mu.mu.RLock(), fn)
}

// mutexTimeout performs a timed unlock, calling supplied fn if timeout is reached
func mutexTimeout(d time.Duration, unlock func(), fn func()) func() {
	if d < 1 {
		// No timeout, just unlock
		return unlock
	}

	// Acquire timer from pool
	t := timerPool.Get().(*timer)

	// Start the timer
	go t.Start(d, fn)

	// Return func cancelling timeout,
	// replacing Timeout in pool and
	// finally unlocking mutex
	return func() {
		defer timerPool.Put(t)
		t.Cancel()
		unlock()
	}
}

// timerPool is the global &timer{} pool.
var timerPool = sync.Pool{
	New: func() interface{} {
		t := time.NewTimer(time.Minute)
		t.Stop()
		return &timer{t: t, c: make(chan struct{})}
	},
}

// timer represents a reusable cancellable timer.
type timer struct {
	t *time.Timer
	c chan struct{}
}

// Start will start the timer with duration 'd', performing 'fn' on timeout.
func (t *timer) Start(d time.Duration, fn func()) {
	t.t.Reset(d)
	select {
	// Timed out
	case <-t.t.C:
		fn()

	// Cancelled
	case <-t.c:
	}
}

// Cancel will attempt to cancel the running timer.
func (t *timer) Cancel() {
	select {
	// cancel successful
	case t.c <- struct{}{}:
		if !t.t.Stop() {
			<-t.t.C
		} // stop timer

	// already stopped
	default:
	}
}
