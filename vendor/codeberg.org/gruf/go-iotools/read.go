package iotools

import (
	"io"
)

// ReaderFunc is a function signature which allows
// a function to implement the io.Reader type.
type ReaderFunc func([]byte) (int, error)

func (r ReaderFunc) Read(b []byte) (int, error) {
	return r(b)
}

// ReaderFromFunc is a function signature which allows
// a function to implement the io.ReaderFrom type.
type ReaderFromFunc func(io.Reader) (int64, error)

func (rf ReaderFromFunc) ReadFrom(r io.Reader) (int64, error) {
	return rf(r)
}

// ReadCloser wraps an io.Reader and io.Closer in order to implement io.ReadCloser.
func ReadCloser(r io.Reader, c io.Closer) io.ReadCloser {
	return &struct {
		io.Reader
		io.Closer
	}{r, c}
}

// NopReadCloser wraps an io.Reader to implement io.ReadCloser with empty io.Closer implementation.
func NopReadCloser(r io.Reader) io.ReadCloser {
	return ReadCloser(r, CloserFunc(func() error {
		return nil
	}))
}
