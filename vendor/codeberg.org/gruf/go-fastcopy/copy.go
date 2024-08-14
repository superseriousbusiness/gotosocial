package fastcopy

import (
	"errors"
	"io"
	"sync"
	_ "unsafe" // link to io.errInvalidWrite.
)

var (
	// global pool instance.
	pool = CopyPool{size: 4096}

	// errInvalidWrite means that a write returned an impossible count.
	errInvalidWrite = errors.New("invalid write result")
)

// CopyPool provides a memory pool of byte
// buffers for io copies from readers to writers.
type CopyPool struct {
	size int
	pool sync.Pool
}

// See CopyPool.Buffer().
func Buffer(sz int) int {
	return pool.Buffer(sz)
}

// See CopyPool.CopyN().
func CopyN(dst io.Writer, src io.Reader, n int64) (int64, error) {
	return pool.CopyN(dst, src, n)
}

// See CopyPool.Copy().
func Copy(dst io.Writer, src io.Reader) (int64, error) {
	return pool.Copy(dst, src)
}

// Buffer sets the pool buffer size to allocate. Returns current size.
// Note this is NOT atomically safe, please call BEFORE other calls to CopyPool.
func (cp *CopyPool) Buffer(sz int) int {
	if sz > 0 {
		// update size
		cp.size = sz
	} else if cp.size < 1 {
		// default size
		return 4096
	}
	return cp.size
}

// CopyN performs the same logic as io.CopyN(), with the difference
// being that the byte buffer is acquired from a memory pool.
func (cp *CopyPool) CopyN(dst io.Writer, src io.Reader, n int64) (int64, error) {
	written, err := cp.Copy(dst, io.LimitReader(src, n))
	if written == n {
		return n, nil
	}
	if written < n && err == nil {
		// src stopped early; must have been EOF.
		err = io.EOF
	}
	return written, err
}

// Copy performs the same logic as io.Copy(), with the difference
// being that the byte buffer is acquired from a memory pool.
func (cp *CopyPool) Copy(dst io.Writer, src io.Reader) (int64, error) {
	// Prefer using io.WriterTo to do the copy (avoids alloc + copy)
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}

	// Prefer using io.ReaderFrom to do the copy.
	if rt, ok := dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(src)
	}

	var buf []byte

	if b, ok := cp.pool.Get().(*[]byte); ok {
		// Acquired buf from pool
		buf = *b
	} else {
		// Allocate new buffer of size
		buf = make([]byte, cp.Buffer(0))
	}

	// Defer release to pool
	defer cp.pool.Put(&buf)

	var n int64
	for {
		// Perform next read into buf
		nr, err := src.Read(buf)
		if nr > 0 {
			// We error check AFTER checking
			// no. read bytes so incomplete
			// read still gets written up to nr.

			// Perform next write from buf
			nw, ew := dst.Write(buf[0:nr])

			// Check for valid write
			if nw < 0 || nr < nw {
				if ew == nil {
					ew = errInvalidWrite
				}
				return n, ew
			}

			// Incr total count
			n += int64(nw)

			// Check write error
			if ew != nil {
				return n, ew
			}

			// Check unequal read/writes
			if nr != nw {
				return n, io.ErrShortWrite
			}
		}

		// Return on err
		if err != nil {
			if err == io.EOF {
				err = nil // expected
			}
			return n, err
		}
	}
}
