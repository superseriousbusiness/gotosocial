package bytebuffer

import (
	"sync"
	"unsafe"
)

// ByteBuffer is a wrapper around a slice to reduce memory allocation while handling blob of data.
type ByteBuffer struct {
	blob []byte
}

// New creates a new ByteBuffer instance.
func New() *ByteBuffer {
	entry := bufferPool.Get().(*ByteBuffer)
	entry.blob = entry.blob[:0]
	return entry
}

// Bytes returns the content buffer.
func (buffer *ByteBuffer) Bytes() []byte {
	return buffer.blob
}

// String returns the content buffer.
func (buffer *ByteBuffer) String() string {
	// Copied from strings.(*Builder).String
	return *(*string)(unsafe.Pointer(&buffer.blob)) // nolint: gosec
}

// Concat appends given arguments to blob content
func (buffer *ByteBuffer) Concat(args ...string) {
	for i := range args {
		buffer.blob = append(buffer.blob, args[i]...)
	}
}

// Close recycles underlying resources of encoder.
func (buffer *ByteBuffer) Close() {
	// Proper usage of a sync.Pool requires each entry to have approximately
	// the same memory cost. To obtain this property when the stored type
	// contains a variably-sized buffer, we add a hard limit on the maximum buffer
	// to place back in the pool.
	//
	// See https://golang.org/issue/23199
	if buffer != nil && cap(buffer.blob) < (1<<16) {
		bufferPool.Put(buffer)
	}
}

// A byte buffer pool to reduce memory allocation pressure.
var bufferPool = &sync.Pool{
	New: func() interface{} {
		return &ByteBuffer{
			blob: make([]byte, 0, 1024),
		}
	},
}
