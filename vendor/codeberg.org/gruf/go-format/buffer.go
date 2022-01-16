package format

import (
	"io"
	"unicode/utf8"
	"unsafe"
)

// ensure we conform to io.Writer.
var _ io.Writer = (*Buffer)(nil)

// Buffer is a simple wrapper around a byte slice.
type Buffer struct {
	B []byte
}

// Write will append given byte slice to buffer, fulfilling io.Writer.
func (buf *Buffer) Write(b []byte) (int, error) {
	buf.B = append(buf.B, b...)
	return len(b), nil
}

// AppendByte appends given byte to the buffer.
func (buf *Buffer) AppendByte(b byte) {
	buf.B = append(buf.B, b)
}

// AppendRune appends given rune to the buffer.
func (buf *Buffer) AppendRune(r rune) {
	if r < utf8.RuneSelf {
		buf.B = append(buf.B, byte(r))
		return
	}

	l := buf.Len()
	for i := 0; i < utf8.UTFMax; i++ {
		buf.B = append(buf.B, 0)
	}
	n := utf8.EncodeRune(buf.B[l:buf.Len()], r)
	buf.B = buf.B[:l+n]
}

// Append will append given byte slice to the buffer.
func (buf *Buffer) Append(b []byte) {
	buf.B = append(buf.B, b...)
}

// AppendString appends given string to the buffer.
func (buf *Buffer) AppendString(s string) {
	buf.B = append(buf.B, s...)
}

// Len returns the length of the buffer's underlying byte slice.
func (buf *Buffer) Len() int {
	return len(buf.B)
}

// Cap returns the capacity of the buffer's underlying byte slice.
func (buf *Buffer) Cap() int {
	return cap(buf.B)
}

// Truncate will reduce the length of the buffer by 'n'.
func (buf *Buffer) Truncate(n int) {
	if n > len(buf.B) {
		n = len(buf.B)
	}
	buf.B = buf.B[:buf.Len()-n]
}

// Reset will reset the buffer length to 0 (retains capacity).
func (buf *Buffer) Reset() {
	buf.B = buf.B[:0]
}

// String returns the underlying byte slice as a string. Please note
// this value is tied directly to the underlying byte slice, if you
// write to the buffer then returned string values will also change.
func (buf *Buffer) String() string {
	return *(*string)(unsafe.Pointer(&buf.B))
}
