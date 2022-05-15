package pools

import (
	"hash"
	"sync"

	"codeberg.org/gruf/go-hashenc"
)

// HashEncoderPool is a pooled allocator for hashenc.HashEncoder objects.
type HashEncoderPool interface {
	// Get fetches a hashenc.HashEncoder from pool
	Get() hashenc.HashEncoder

	// Put places supplied hashenc.HashEncoder back in pool
	Put(hashenc.HashEncoder)
}

// NewHashEncoderPool returns a newly instantiated hashenc.HashEncoder pool.
func NewHashEncoderPool(hash func() hash.Hash, enc func() hashenc.Encoder) HashEncoderPool {
	return &hencPool{
		pool: sync.Pool{
			New: func() interface{} {
				return hashenc.New(hash(), enc())
			},
		},
		size: hashenc.New(hash(), enc()).Size(),
	}
}

// hencPool is our implementation of HashEncoderPool.
type hencPool struct {
	pool sync.Pool
	size int
}

func (p *hencPool) Get() hashenc.HashEncoder {
	return p.pool.Get().(hashenc.HashEncoder)
}

func (p *hencPool) Put(henc hashenc.HashEncoder) {
	if henc.Size() < p.size {
		return
	}
	p.pool.Put(henc)
}
