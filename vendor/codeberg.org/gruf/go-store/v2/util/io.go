package util

import (
	"bytes"
	"io"
)

// ReaderSize defines a reader of known size in bytes.
type ReaderSize interface {
	io.Reader
	Size() int64
}

// ByteReaderSize implements ReaderSize for an in-memory byte-slice.
type ByteReaderSize struct {
	br bytes.Reader
	sz int64
}

// NewByteReaderSize returns a new ByteReaderSize instance reset to slice b.
func NewByteReaderSize(b []byte) *ByteReaderSize {
	rs := new(ByteReaderSize)
	rs.Reset(b)
	return rs
}

// Read implements io.Reader.
func (rs *ByteReaderSize) Read(b []byte) (int, error) {
	return rs.br.Read(b)
}

// Size implements ReaderSize.
func (rs *ByteReaderSize) Size() int64 {
	return rs.sz
}

// Reset resets the ReaderSize to be reading from b.
func (rs *ByteReaderSize) Reset(b []byte) {
	rs.br.Reset(b)
	rs.sz = int64(len(b))
}
