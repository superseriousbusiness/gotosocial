package iotools

import "io"

// NopCloser is an empty
// implementation of io.Closer,
// that simply does nothing!
type NopCloser struct{}

func (NopCloser) Close() error { return nil }

// CloserFunc is a function signature which allows
// a function to implement the io.Closer type.
type CloserFunc func() error

func (c CloserFunc) Close() error {
	return c()
}

// CloserCallback wraps io.Closer to add a callback deferred to call just after Close().
func CloserCallback(c io.Closer, cb func()) io.Closer {
	return CloserFunc(func() error {
		defer cb()
		return c.Close()
	})
}

// CloserAfterCallback wraps io.Closer to add a callback called just before Close().
func CloserAfterCallback(c io.Closer, cb func()) io.Closer {
	return CloserFunc(func() (err error) {
		defer func() { err = c.Close() }()
		cb()
		return
	})
}

// CloseOnce wraps an io.Closer to ensure it only performs the close logic once.
func CloseOnce(c io.Closer) io.Closer {
	return CloserFunc(func() error {
		if c == nil {
			// already run.
			return nil
		}

		// Acquire.
		cptr := c
		c = nil

		// Call the closer.
		return cptr.Close()
	})
}
