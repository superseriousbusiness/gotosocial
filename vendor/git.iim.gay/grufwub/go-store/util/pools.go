package util

import (
	"sync"

	"git.iim.gay/grufwub/fastpath"
	"git.iim.gay/grufwub/go-bufpool"
	"git.iim.gay/grufwub/go-bytes"
)

// pathBuilderPool is the global fastpath.Builder pool, we implement
// our own here instead of using fastpath's default one because we
// don't want to deal with fastpath's sync.Once locks on every Acquire/Release
var pathBuilderPool = sync.Pool{
	New: func() interface{} {
		pb := fastpath.NewBuilder(make([]byte, 0, 512))
		return &pb
	},
}

// AcquirePathBuilder returns a reset fastpath.Builder instance
func AcquirePathBuilder() *fastpath.Builder {
	return pathBuilderPool.Get().(*fastpath.Builder)
}

// ReleasePathBuilder resets and releases provided fastpath.Builder instance to global pool
func ReleasePathBuilder(pb *fastpath.Builder) {
	pb.Reset()
	pathBuilderPool.Put(pb)
}

// bufferPool is the global BufferPool, we implement this here
// so we can share allocations across whatever libaries need them.
var bufferPool = bufpool.BufferPool{}

// AcquireBuffer returns a reset bytes.Buffer with at least requested capacity
func AcquireBuffer(cap int) *bytes.Buffer {
	return bufferPool.Get(cap)
}

// ReleaseBuffer resets and releases provided bytes.Buffer to global BufferPool
func ReleaseBuffer(buf *bytes.Buffer) {
	bufferPool.Put(buf)
}
