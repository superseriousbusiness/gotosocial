package bytes

import (
	"bytes"
	"sync"
)

type SizedBufferPool struct {
	pool sync.Pool
	len  int
	cap  int
}

func (p *SizedBufferPool) Init(len, cap int) {
	p.pool.New = func() interface{} {
		buf := NewBuffer(make([]byte, len, cap))
		return &buf
	}
	p.len = len
	p.cap = cap
}

func (p *SizedBufferPool) Acquire() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

func (p *SizedBufferPool) Release(buf *bytes.Buffer) {
	// If not enough cap, ignore
	if buf.Cap() < p.cap {
		return
	}

	// Set length to expected
	buf.Reset()
	buf.Grow(p.len)

	// Place in pool
	p.pool.Put(buf)
}
