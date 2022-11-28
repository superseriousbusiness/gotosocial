package util

import (
	"sync"

	"codeberg.org/gruf/go-fastpath/v2"
)

// pathBuilderPool is the global fastpath.Builder pool.
var pathBuilderPool = sync.Pool{
	New: func() any {
		return &fastpath.Builder{B: make([]byte, 0, 512)}
	},
}

// GetPathBuilder fetches a fastpath.Builder object from the pool.
func GetPathBuilder() *fastpath.Builder {
	pb, _ := pathBuilderPool.Get().(*fastpath.Builder)
	return pb
}

// PutPathBuilder places supplied fastpath.Builder back in the pool.
func PutPathBuilder(pb *fastpath.Builder) {
	pb.Reset()
	pathBuilderPool.Put(pb)
}
