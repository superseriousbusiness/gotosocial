package util

import (
	"codeberg.org/gruf/go-fastpath"
	"codeberg.org/gruf/go-pools"
)

// pathBuilderPool is the global fastpath.Builder pool.
var pathBuilderPool = pools.NewPathBuilderPool(512)

// GetPathBuilder fetches a fastpath.Builder object from the pool.
func GetPathBuilder() *fastpath.Builder {
	return pathBuilderPool.Get()
}

// PutPathBuilder places supplied fastpath.Builder back in the pool.
func PutPathBuilder(pb *fastpath.Builder) {
	pb.Reset()
	pathBuilderPool.Put(pb)
}
