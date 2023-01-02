package iotools

import "io"

// WriterFunc is a function signature which allows
// a function to implement the io.Writer type.
type WriterFunc func([]byte) (int, error)

func (w WriterFunc) Write(b []byte) (int, error) {
	return w(b)
}

// WriteCloser wraps an io.Writer and io.Closer in order to implement io.WriteCloser.
func WriteCloser(w io.Writer, c io.Closer) io.WriteCloser {
	return &struct {
		io.Writer
		io.Closer
	}{w, c}
}

// NopWriteCloser wraps an io.Writer to implement io.WriteCloser with empty io.Closer implementation.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return WriteCloser(w, CloserFunc(func() error {
		return nil
	}))
}
