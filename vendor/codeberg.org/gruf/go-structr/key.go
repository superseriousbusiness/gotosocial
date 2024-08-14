package structr

import (
	"sync"

	"codeberg.org/gruf/go-byteutil"
)

// Key represents one key to
// lookup (potentially) stored
// entries in an Index.
type Key struct {
	key string
	raw []any
}

// Key returns the underlying cache key string.
// NOTE: this will not be log output friendly.
func (k Key) Key() string {
	return k.key
}

// Equal returns whether keys are equal.
func (k Key) Equal(o Key) bool {
	return (k.key == o.key)
}

// Value returns the raw slice of
// values that comprise this Key.
func (k Key) Values() []any {
	return k.raw
}

// Zero indicates a zero value key.
func (k Key) Zero() bool {
	return (k.raw == nil)
}

var buf_pool sync.Pool

// new_buffer returns a new initialized byte buffer.
func new_buffer() *byteutil.Buffer {
	v := buf_pool.Get()
	if v == nil {
		buf := new(byteutil.Buffer)
		buf.B = make([]byte, 0, 512)
		v = buf
	}
	return v.(*byteutil.Buffer)
}

// free_buffer releases the byte buffer.
func free_buffer(buf *byteutil.Buffer) {
	if cap(buf.B) > int(^uint16(0)) {
		return // drop large bufs
	}
	buf_pool.Put(buf)
}
