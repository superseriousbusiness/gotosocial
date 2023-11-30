package errors

import (
	"sync/atomic"
)

// OnceError is an error structure that supports safe multi
// threaded usage and setting only once (until reset).
type OnceError struct{ ptr atomic.Pointer[error] }

// Store will safely set the OnceError to value, no-op if nil.
func (e *OnceError) Store(err error) bool {
	if err == nil {
		return false
	}
	return e.ptr.CompareAndSwap(nil, &err)
}

// Load will load the currently stored error.
func (e *OnceError) Load() error {
	if ptr := e.ptr.Load(); ptr != nil {
		return *ptr
	}
	return nil
}

// IsSet returns whether OnceError has been set.
func (e *OnceError) IsSet() bool { return (e.ptr.Load() != nil) }

// Reset will reset the OnceError value.
func (e *OnceError) Reset() { e.ptr.Store(nil) }
