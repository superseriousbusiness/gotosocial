package byteutil

import (
	"errors"
	"io"
	"unicode/utf8"
)

var (
	// ensure we conform
	// to interfaces.
	_ interface {
		io.Writer
		io.ByteWriter
		WriteRune(rune) (int, error)
		io.StringWriter
		io.WriterAt
		WriteStringAt(string, int64) (int, error)
		io.ReaderFrom
		io.WriterTo
	} = (*Buffer)(nil)

	// ErrBeyondBufferLen is returned if .WriteAt() is attempted beyond buffer length.
	ErrBeyondBufferLen = errors.New("start beyond buffer length")
)

// Buffer is a simple wrapper around a byte slice.
type Buffer struct{ B []byte }

// WriteByte will append given byte to buffer, fulfilling io.ByteWriter.
func (buf *Buffer) WriteByte(c byte) error {
	buf.B = append(buf.B, c)
	return nil
}

// WriteRune will append given rune to buffer.
func (buf *Buffer) WriteRune(r rune) (int, error) {
	// Check for single-byte rune
	if r < utf8.RuneSelf {
		buf.B = append(buf.B, byte(r))
		return 1, nil
	}

	// Before-len
	l := len(buf.B)

	// Grow to max size rune
	buf.Grow(utf8.UTFMax)

	// Write encoded rune to buffer
	n := utf8.EncodeRune(buf.B[l:len(buf.B)], r)
	buf.B = buf.B[:l+n]

	return n, nil
}

// Write will append given byte slice to buffer, fulfilling io.Writer.
func (buf *Buffer) Write(b []byte) (int, error) {
	buf.B = append(buf.B, b...)
	return len(b), nil
}

// WriteString will append given string to buffer, fulfilling io.StringWriter.
func (buf *Buffer) WriteString(s string) (int, error) {
	buf.B = append(buf.B, s...)
	return len(s), nil
}

// WriteAt will append given byte slice to buffer at index 'start', fulfilling io.WriterAt.
func (buf *Buffer) WriteAt(b []byte, start int64) (int, error) {
	if start > int64(len(buf.B)) {
		return 0, ErrBeyondBufferLen
	}
	buf.Grow(len(b) - int(int64(len(buf.B))-start))
	return copy(buf.B[start:], b), nil
}

// WriteStringAt will append given string to buffer at index 'start'.
func (buf *Buffer) WriteStringAt(s string, start int64) (int, error) {
	if start > int64(len(buf.B)) {
		return 0, ErrBeyondBufferLen
	}
	buf.Grow(len(s) - int(int64(len(buf.B))-start))
	return copy(buf.B[start:], s), nil
}

// ReadFrom will read bytes from reader into buffer, fulfilling io.ReaderFrom.
func (buf *Buffer) ReadFrom(r io.Reader) (int64, error) {
	var nn int64

	// Ensure there's cap
	// for a first read.
	buf.Guarantee(512)

	for {
		// Read into next chunk of buffer.
		n, err := r.Read(buf.B[len(buf.B):cap(buf.B)])

		// Reslice buf + update count.
		buf.B = buf.B[:len(buf.B)+n]
		nn += int64(n)

		if err != nil {
			if err == io.EOF {
				// mask EOF.
				err = nil
			}
			return nn, err
		}

		if len(buf.B) == cap(buf.B) {
			// Add capacity (let append pick).
			buf.B = append(buf.B, 0)[:len(buf.B)]
		}
	}
}

// WriteTo will write bytes from buffer into writer, fulfilling io.WriterTo.
func (buf *Buffer) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(buf.B)
	return int64(n), err
}

// Len returns the length of the buffer's underlying byte slice.
func (buf *Buffer) Len() int {
	return len(buf.B)
}

// Cap returns the capacity of the buffer's underlying byte slice.
func (buf *Buffer) Cap() int {
	return cap(buf.B)
}

// Grow will increase the buffers length by 'sz', and the capacity by at least this.
func (buf *Buffer) Grow(sz int) {
	buf.Guarantee(sz)
	buf.B = buf.B[:len(buf.B)+sz]
}

// Guarantee will guarantee buffer containers at least 'sz' remaining capacity.
func (buf *Buffer) Guarantee(sz int) {
	if sz > cap(buf.B)-len(buf.B) {
		nb := make([]byte, 2*cap(buf.B)+sz)
		copy(nb, buf.B)
		buf.B = nb[:len(buf.B)]
	}
}

// Truncate will reduce the length of the buffer by 'n'.
func (buf *Buffer) Truncate(n int) {
	if n > len(buf.B) {
		n = len(buf.B)
	}
	buf.B = buf.B[:len(buf.B)-n]
}

// Reset will reset the buffer length to 0 (retains capacity).
func (buf *Buffer) Reset() {
	buf.B = buf.B[:0]
}

// String returns the underlying byte slice as a string. Please note
// this value is tied directly to the underlying byte slice, if you
// write to the buffer then returned string values will also change.
//
// To get an immutable string from buffered data, use string(buf.B).
func (buf *Buffer) String() string {
	return B2S(buf.B)
}

// Full returns the full capacity byteslice allocated for this buffer.
func (buf *Buffer) Full() []byte {
	return buf.B[0:cap(buf.B)]
}
