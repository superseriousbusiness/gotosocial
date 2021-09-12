package bytes

import (
	"unicode/utf8"
)

// Buffer is a very simple buffer implementation that allows
// access to and reslicing of the underlying byte slice.
type Buffer struct {
	noCopy noCopy
	B      []byte
}

func NewBuffer(b []byte) Buffer {
	return Buffer{
		noCopy: noCopy{},
		B:      b,
	}
}

func (b *Buffer) Write(p []byte) (int, error) {
	b.Grow(len(p))
	return copy(b.B[b.Len()-len(p):], p), nil
}

func (b *Buffer) WriteString(s string) (int, error) {
	b.Grow(len(s))
	return copy(b.B[b.Len()-len(s):], s), nil
}

func (b *Buffer) WriteByte(c byte) error {
	l := b.Len()
	b.Grow(1)
	b.B[l] = c
	return nil
}

func (b *Buffer) WriteRune(r rune) (int, error) {
	if r < utf8.RuneSelf {
		b.WriteByte(byte(r))
		return 1, nil
	}

	l := b.Len()
	b.Grow(utf8.UTFMax)
	n := utf8.EncodeRune(b.B[l:b.Len()], r)
	b.B = b.B[:l+n]

	return n, nil
}

func (b *Buffer) WriteAt(p []byte, start int64) (int, error) {
	b.Grow(len(p) - int(int64(b.Len())-start))
	return copy(b.B[start:], p), nil
}

func (b *Buffer) WriteStringAt(s string, start int64) (int, error) {
	b.Grow(len(s) - int(int64(b.Len())-start))
	return copy(b.B[start:], s), nil
}

func (b *Buffer) Truncate(size int) {
	b.B = b.B[:b.Len()-size]
}

func (b *Buffer) ShiftByte(index int) {
	copy(b.B[index:], b.B[index+1:])
}

func (b *Buffer) Shift(start int64, size int) {
	copy(b.B[start:], b.B[start+int64(size):])
}

func (b *Buffer) DeleteByte(index int) {
	b.ShiftByte(index)
	b.Truncate(1)
}

func (b *Buffer) Delete(start int64, size int) {
	b.Shift(start, size)
	b.Truncate(size)
}

func (b *Buffer) InsertByte(index int64, c byte) {
	l := b.Len()
	b.Grow(1)
	copy(b.B[index+1:], b.B[index:l])
	b.B[index] = c
}

func (b *Buffer) Insert(index int64, p []byte) {
	l := b.Len()
	b.Grow(len(p))
	copy(b.B[index+int64(len(p)):], b.B[index:l])
	copy(b.B[index:], p)
}

func (b *Buffer) Bytes() []byte {
	return b.B
}

func (b *Buffer) String() string {
	return string(b.B)
}

func (b *Buffer) StringPtr() string {
	return BytesToString(b.B)
}

func (b *Buffer) Cap() int {
	return cap(b.B)
}

func (b *Buffer) Len() int {
	return len(b.B)
}

func (b *Buffer) Reset() {
	b.B = b.B[:0]
}

func (b *Buffer) Grow(size int) {
	b.Guarantee(size)
	b.B = b.B[:b.Len()+size]
}

func (b *Buffer) Guarantee(size int) {
	if size > b.Cap()-b.Len() {
		nb := make([]byte, 2*b.Cap()+size)
		copy(nb, b.B)
		b.B = nb[:b.Len()]
	}
}

type noCopy struct{}

func (n *noCopy) Lock()   {}
func (n *noCopy) Unlock() {}
