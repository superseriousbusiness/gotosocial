package errors

import (
	"sync/atomic"
	"unsafe"
)

// OnceError is an error structure that supports safe multi-threaded
// usage and setting only once (until reset)
type OnceError struct {
	err unsafe.Pointer
}

// NewOnce returns a new OnceError instance
func NewOnce() OnceError {
	return OnceError{
		err: nil,
	}
}

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

func (e *OnceError) Load() error {
	return *(*error)(atomic.LoadPointer(&e.err))
}

func (e *OnceError) IsSet() bool {
	return (atomic.LoadPointer(&e.err) != nil)
}

func (e *OnceError) Reset() {
	atomic.StorePointer(&e.err, nil)
}
