package logger

import (
	"io"
	"io/ioutil"
	"sync"
)

// AddSafety wraps an io.Writer to provide mutex locking protection
func AddSafety(w io.Writer) io.Writer {
	if w == nil {
		w = ioutil.Discard
	} else if sw, ok := w.(*safeWriter); ok {
		return sw
	}
	return &safeWriter{wr: w}
}

// safeWriter wraps an io.Writer to provide mutex locking on write
type safeWriter struct {
	wr io.Writer
	mu sync.Mutex
}

func (w *safeWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.wr.Write(b)
}
