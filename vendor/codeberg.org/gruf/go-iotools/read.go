package iotools

import (
	"io"
)

// ReadCloserType implements io.ReadCloser
// by combining the two underlying interfaces,
// while providing an exported type to still
// access the underlying original io.Reader or
// io.Closer separately (e.g. without wrapping).
type ReadCloserType struct {
	io.Reader
	io.Closer
}

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
	return &ReadCloserType{r, c}
}

// NopReadCloser wraps io.Reader with NopCloser{} in ReadCloserType.
func NopReadCloser(r io.Reader) io.ReadCloser {
	return &ReadCloserType{r, NopCloser{}}
}
