package iotools

import (
	"errors"
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

// AtEOF returns true when reader at EOF,
// this is checked with a 0 length read.
func AtEOF(r io.Reader) bool {
	_, err := r.Read(nil)
	return (err == io.EOF)
}

// ErrLimitReached is returned when an io.Reader is limited with more data remaining.
var ErrLimitReached = errors.New("read limit reached")

// LimitReader wraps io.Reader to limit reads to at-most 'limit'.
func LimitReader(r io.Reader, limit int64) io.Reader {
	return &LimitedReader{r: r, n: limit}
}

// LimitedReader wraps an io.Reader to
// limit reads to a predefined limit.
type LimitedReader struct {
	r io.Reader

	// > 0 = reading
	// < 0 = limited
	//  0  = eof
	n int64
}

func (l *LimitedReader) Read(p []byte) (int, error) {
	switch {
	case l.n < 0:
		return 0, ErrLimitReached
	case l.n == 0:
		return 0, io.EOF
	}
	if int64(len(p)) > l.n {
		p = p[0:l.n]
	}
	n, err := l.r.Read(p)
	l.n -= int64(n)
	if err == nil {
		if l.n <= 0 {
			if AtEOF(l.r) {
				err = io.EOF
				l.n = 0
			} else {
				err = ErrLimitReached
				l.n = -1
			}
		}
	} else if err == io.EOF {
		l.n = 0
	}
	return n, err
}
