package errors

import (
	"sync/atomic"
	"unsafe"
)

// OnceError is an error structure that supports safe multi
// threaded usage and setting only once (until reset).
type OnceError struct{ err unsafe.Pointer }

// NewOnce returns a new OnceError instance.
func NewOnce() OnceError {
	return OnceError{
		err: nil,
	}
}

// Store will safely set the OnceError to value, no-op if nil.
func (e *OnceError) Store(err error) {
	// Nothing to do
	if err == nil {
		return
	}

	// Only set if not already
	atomic.CompareAndSwapPointer(
		&e.err,
		nil,
		unsafe.Pointer(&err),
	)
}

// Load will load the currently stored error.
func (e *OnceError) Load() error {
	return *(*error)(atomic.LoadPointer(&e.err))
}

// IsSet returns whether OnceError has been set.
func (e *OnceError) IsSet() bool {
	return (atomic.LoadPointer(&e.err) != nil)
}

// Reset will reset the OnceError value.
func (e *OnceError) Reset() {
	atomic.StorePointer(&e.err, nil)
}
