package iotools

import (
	"io"
)

// ReadFnCloser takes an io.Reader and wraps it to use the provided function to implement io.Closer.
func ReadFnCloser(r io.Reader, close func() error) io.ReadCloser {
	return &readFnCloser{
		Reader: r,
		close:  close,
	}
}

type readFnCloser struct {
	io.Reader
	close func() error
}

func (r *readFnCloser) Close() error {
	return r.close()
}

// WriteFnCloser takes an io.Writer and wraps it to use the provided function to implement io.Closer.
func WriteFnCloser(w io.Writer, close func() error) io.WriteCloser {
	return &writeFnCloser{
		Writer: w,
		close:  close,
	}
}

type writeFnCloser struct {
	io.Writer
	close func() error
}

func (r *writeFnCloser) Close() error {
	return r.close()
}

// SilentReader wraps an io.Reader to silence any
// error output during reads. Instead they are stored
// and accessible (not concurrency safe!) via .Error().
type SilentReader struct {
	io.Reader
	err error
}

// SilenceReader wraps an io.Reader within SilentReader{}.
func SilenceReader(r io.Reader) *SilentReader {
	return &SilentReader{Reader: r}
}

func (r *SilentReader) Read(b []byte) (int, error) {
	n, err := r.Reader.Read(b)
	if err != nil {
		// Store error for now
		if r.err == nil {
			r.err = err
		}

		// Pretend we're happy
		// to continue reading.
		n = len(b)
	}
	return n, nil
}

func (r *SilentReader) Error() error {
	return r.err
}

// SilentWriter wraps an io.Writer to silence any
// error output during writes. Instead they are stored
// and accessible (not concurrency safe!) via .Error().
type SilentWriter struct {
	io.Writer
	err error
}

// SilenceWriter wraps an io.Writer within SilentWriter{}.
func SilenceWriter(w io.Writer) *SilentWriter {
	return &SilentWriter{Writer: w}
}

func (w *SilentWriter) Write(b []byte) (int, error) {
	n, err := w.Writer.Write(b)
	if err != nil {
		// Store error for now
		if w.err == nil {
			w.err = err
		}

		// Pretend we're happy
		// to continue writing.
		n = len(b)
	}
	return n, nil
}

func (w *SilentWriter) Error() error {
	return w.err
}
