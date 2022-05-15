package pools

import (
	"sync"

	"codeberg.org/gruf/go-byteutil"
)

// BufferPool is a pooled allocator for bytes.Buffer objects
type BufferPool interface {
	// Get fetches a bytes.Buffer from pool
	Get() *byteutil.Buffer

	// Put places supplied bytes.Buffer in pool
	Put(*byteutil.Buffer)
}

// NewBufferPool returns a newly instantiated bytes.Buffer pool
func NewBufferPool(size int) BufferPool {
	return &bufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &byteutil.Buffer{B: make([]byte, 0, size)}
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

func (p *bufferPool) Get() *byteutil.Buffer {
	return p.pool.Get().(*byteutil.Buffer)
}

func (p *bufferPool) Put(buf *byteutil.Buffer) {
	if buf.Cap() < p.size {
		return
	}
	buf.Reset()
	p.pool.Put(buf)
}
