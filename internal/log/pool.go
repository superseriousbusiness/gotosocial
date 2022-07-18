package log

import (
	"sync"

	"codeberg.org/gruf/go-byteutil"
)

// bufPool provides a memory pool of log buffers.
var bufPool = sync.Pool{
	New: func() any {
		return &byteutil.Buffer{
			B: make([]byte, 0, 512),
		}
	},
}

// getBuf acquires a buffer from memory pool.
func getBuf() *byteutil.Buffer {
	buf, _ := bufPool.Get().(*byteutil.Buffer)
	return buf
}

// putBuf places (after resetting) buffer back in memory pool, dropping if capacity too large.
func putBuf(buf *byteutil.Buffer) {
	if buf.Cap() > int(^uint16(0)) {
		return // drop large buffer
	}
	buf.Reset()
	bufPool.Put(buf)
}
