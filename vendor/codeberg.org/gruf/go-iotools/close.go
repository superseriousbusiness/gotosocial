package iotools

import "io"

// CloserFunc is a function signature which allows
// a function to implement the io.Closer type.
type CloserFunc func() error

func (c CloserFunc) Close() error {
	return c()
}

func CloserCallback(c io.Closer, cb func()) io.Closer {
	return CloserFunc(func() error {
		defer cb()
		return c.Close()
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
