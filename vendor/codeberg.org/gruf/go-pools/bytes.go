package pools

import (
	"sync"

	"codeberg.org/gruf/go-bytes"
)

// BufferPool is a pooled allocator for bytes.Buffer objects
type BufferPool interface {
	// Get fetches a bytes.Buffer from pool
	Get() *bytes.Buffer

	// Put places supplied bytes.Buffer in pool
	Put(*bytes.Buffer)
}

// NewBufferPool returns a newly instantiated bytes.Buffer pool
func NewBufferPool(size int) BufferPool {
	return &bufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{B: make([]byte, 0, size)}
			},
		},
		size: size,
	}
}

// bufferPool is our implementation of BufferPool
type bufferPool struct {
	pool sync.Pool
	size int
}

func (p *bufferPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

func (p *bufferPool) Put(buf *bytes.Buffer) {
	if buf.Cap() < p.size {
		return
	}
	buf.Reset()
	p.pool.Put(buf)
}
