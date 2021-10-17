package cache

import (
	"sync"
	"sync/atomic"
)

// mutex represents a mutex with the ability to only "attempt" locks
type mutex struct {
	mu sync.Mutex // underlying mutex
	st uint32     // current lock state
	at uint32     // tracks no. failed lock attempts
}

func (m *mutex) Lock() {
	m.mu.Lock()
	atomic.StoreUint32(&m.st, 1)
}

func (m *mutex) Unlock() {
	atomic.StoreUint32(&m.st, 0)
	m.mu.Unlock()
}

// AttemptLock attempts to acquire a mutex lock, returning success state
func (m *mutex) AttemptLock() bool {
	if (atomic.LoadUint32(&m.st) == 0) || (atomic.AddUint32(&m.at, 1) > maxLockAttempts) {
		// Either:
		// - not locked, so we acquire!
		// - we hit max failed attempts, force lock
		m.Lock()
		atomic.StoreUint32(&m.at, 0) // reset
		return true
	}
	return false
}
