package util

import (
	"bytes"
	"io"
)

// ReaderSize ...
type ReaderSize interface {
	io.Reader

	// Size ...
	Size() int64
}

// ByteReaderSize ...
type ByteReaderSize struct {
	bytes.Reader
	sz int64
}

// NewByteReaderSize ...
func NewByteReaderSize(b []byte) *ByteReaderSize {
	rs := ByteReaderSize{}
	rs.Reset(b)
	return &rs
}

// Size implements ReaderSize.Size().
func (rs ByteReaderSize) Size() int64 {
	return rs.sz
}

// Reset resets the ReaderSize to be reading from b.
func (rs *ByteReaderSize) Reset(b []byte) {
	rs.Reader.Reset(b)
	rs.sz = int64(len(b))
}

// NopReadCloser turns a supplied io.Reader into io.ReadCloser with a nop Close() implementation.
func NopReadCloser(r io.Reader) io.ReadCloser {
	return &nopReadCloser{r}
}

// NopWriteCloser turns a supplied io.Writer into io.WriteCloser with a nop Close() implementation.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}

// ReadCloserWithCallback adds a customizable callback to be called upon Close() of a supplied io.ReadCloser.
// Note that the callback will never be called more than once, after execution this will remove the func reference.
func ReadCloserWithCallback(rc io.ReadCloser, cb func()) io.ReadCloser {
	return &callbackReadCloser{
		ReadCloser: rc,
		callback:   cb,
	}
}

// WriteCloserWithCallback adds a customizable callback to be called upon Close() of a supplied io.WriteCloser.
// Note that the callback will never be called more than once, after execution this will remove the func reference.
func WriteCloserWithCallback(wc io.WriteCloser, cb func()) io.WriteCloser {
	return &callbackWriteCloser{
		WriteCloser: wc,
		callback:    cb,
	}
}

// nopReadCloser turns an io.Reader -> io.ReadCloser with a nop Close().
type nopReadCloser struct{ io.Reader }

func (r *nopReadCloser) Close() error { return nil }

// nopWriteCloser turns an io.Writer -> io.WriteCloser with a nop Close().
type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

// callbackReadCloser allows adding our own custom callback to an io.ReadCloser.
type callbackReadCloser struct {
	io.ReadCloser
	callback func()
}

func (c *callbackReadCloser) Close() error {
	if c.callback != nil {
		cb := c.callback
		c.callback = nil
		defer cb()
	}
	return c.ReadCloser.Close()
}

// callbackWriteCloser allows adding our own custom callback to an io.WriteCloser.
type callbackWriteCloser struct {
	io.WriteCloser
	callback func()
}

func (c *callbackWriteCloser) Close() error {
	if c.callback != nil {
		cb := c.callback
		c.callback = nil
		defer cb()
	}
	return c.WriteCloser.Close()
}
