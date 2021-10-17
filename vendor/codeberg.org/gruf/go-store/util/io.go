package util

import "io"

// NopReadCloser turns a supplied io.Reader into io.ReadCloser with a nop Close() implementation
func NopReadCloser(r io.Reader) io.ReadCloser {
	return &nopReadCloser{r}
}

// NopWriteCloser turns a supplied io.Writer into io.WriteCloser with a nop Close() implementation
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}

// ReadCloserWithCallback adds a customizable callback to be called upon Close() of a supplied io.ReadCloser
func ReadCloserWithCallback(rc io.ReadCloser, cb func()) io.ReadCloser {
	return &callbackReadCloser{
		ReadCloser: rc,
		callback:   cb,
	}
}

// nopReadCloser turns an io.Reader -> io.ReadCloser with a nop Close()
type nopReadCloser struct{ io.Reader }

func (r *nopReadCloser) Close() error { return nil }

// nopWriteCloser turns an io.Writer -> io.WriteCloser with a nop Close()
type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

// callbackReadCloser allows adding our own custom callback to an io.ReadCloser
type callbackReadCloser struct {
	io.ReadCloser
	callback func()
}

func (c *callbackReadCloser) Close() error {
	defer c.callback()
	return c.ReadCloser.Close()
}
