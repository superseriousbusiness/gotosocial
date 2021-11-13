package pools

import (
	"sync"

	"codeberg.org/gruf/go-fastpath"
)

// PathBuilderPool is a pooled allocator for fastpath.Builder objects
type PathBuilderPool interface {
	// Get fetches a fastpath.Builder from pool
	Get() *fastpath.Builder

	// Put places supplied fastpath.Builder back in pool
	Put(*fastpath.Builder)
}

// NewPathBuilderPool returns a newly instantiated fastpath.Builder pool
func NewPathBuilderPool(size int) PathBuilderPool {
	return &pathBuilderPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &fastpath.Builder{B: make([]byte, 0, size)}
			},
		},
		size: size,
	}
}

// pathBuilderPool is our implementation of PathBuilderPool
type pathBuilderPool struct {
	pool sync.Pool
	size int
}

func (p *pathBuilderPool) Get() *fastpath.Builder {
	return p.pool.Get().(*fastpath.Builder)
}

func (p *pathBuilderPool) Put(pb *fastpath.Builder) {
	if pb.Cap() < p.size {
		return
	}
	pb.Reset()
	p.pool.Put(pb)
}
