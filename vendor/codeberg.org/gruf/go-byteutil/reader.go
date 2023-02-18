package byteutil

import "bytes"

// Reader wraps a bytes.Reader{} to provide Rewind() capabilities.
type Reader struct {
	B []byte
	bytes.Reader
}

// NewReader returns a new Reader{} instance reset to b.
func NewReader(b []byte) *Reader {
	r := &Reader{}
	r.Reset(b)
	return r
}

// Reset resets the Reader{} to be reading from b and sets Reader{}.B.
func (r *Reader) Reset(b []byte) {
	r.B = b
	r.Rewind()
}

// Rewind resets the Reader{} to be reading from the start of Reader{}.B.
func (r *Reader) Rewind() {
	r.Reader.Reset(r.B)
}

// ReadNopCloser wraps a Reader{} to provide nop close method.
type ReadNopCloser struct {
	Reader
}

func (*ReadNopCloser) Close() error {
	return nil
}
