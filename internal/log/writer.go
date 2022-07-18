package log

import (
	"io"
	"sync"
)

// safewriter wraps a writer to provide mutex safety on write.
type safewriter struct {
	w io.Writer
	m sync.Mutex
}

func (w *safewriter) Write(b []byte) (int, error) {
	w.m.Lock()
	n, err := w.w.Write(b)
	w.m.Unlock()
	return n, err
}
