package parse

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

// BinaryReader is a binary big endian file format reader.
type BinaryReader struct {
	Endianness binary.ByteOrder
	buf        []byte
	pos        uint32
	eof        bool
}

// NewBinaryReader returns a big endian binary file format reader.
func NewBinaryReader(buf []byte) *BinaryReader {
	if math.MaxUint32 < uint(len(buf)) {
		return &BinaryReader{binary.BigEndian, nil, 0, true}
	}
	return &BinaryReader{binary.BigEndian, buf, 0, false}
}

// NewBinaryReaderLE returns a little endian binary file format reader.
func NewBinaryReaderLE(buf []byte) *BinaryReader {
	r := NewBinaryReader(buf)
	r.Endianness = binary.LittleEndian
	return r
}

// Seek set the reader position in the buffer.
func (r *BinaryReader) Seek(pos uint32) error {
	if uint32(len(r.buf)) < pos {
		r.eof = true
		return io.EOF
	}
	r.pos = pos
	r.eof = false
	return nil
}

// Pos returns the reader's position.
func (r *BinaryReader) Pos() uint32 {
	return r.pos
}

// Len returns the remaining length of the buffer.
func (r *BinaryReader) Len() uint32 {
	return uint32(len(r.buf)) - r.pos
}

// EOF returns true if we reached the end-of-file.
func (r *BinaryReader) EOF() bool {
	return r.eof
}

// Read complies with io.Reader.
func (r *BinaryReader) Read(b []byte) (int, error) {
	n := copy(b, r.buf[r.pos:])
	r.pos += uint32(n)
	if r.pos == uint32(len(r.buf)) {
		r.eof = true
		return n, io.EOF
	}
	return n, nil
}

// ReadBytes reads n bytes.
func (r *BinaryReader) ReadBytes(n uint32) []byte {
	if r.eof || uint32(len(r.buf))-r.pos < n {
		r.eof = true
		return nil
	}
	buf := r.buf[r.pos : r.pos+n : r.pos+n]
	r.pos += n
	return buf
}

// ReadString reads a string of length n.
func (r *BinaryReader) ReadString(n uint32) string {
	return string(r.ReadBytes(n))
}

// ReadByte reads a single byte.
func (r *BinaryReader) ReadByte() byte {
	b := r.ReadBytes(1)
	if b == nil {
		return 0
	}
	return b[0]
}

// ReadUint8 reads a uint8.
func (r *BinaryReader) ReadUint8() uint8 {
	return r.ReadByte()
}

// ReadUint16 reads a uint16.
func (r *BinaryReader) ReadUint16() uint16 {
	b := r.ReadBytes(2)
	if b == nil {
		return 0
	}
	return r.Endianness.Uint16(b)
}

// ReadUint32 reads a uint32.
func (r *BinaryReader) ReadUint32() uint32 {
	b := r.ReadBytes(4)
	if b == nil {
		return 0
	}
	return r.Endianness.Uint32(b)
}

// ReadUint64 reads a uint64.
func (r *BinaryReader) ReadUint64() uint64 {
	b := r.ReadBytes(8)
	if b == nil {
		return 0
	}
	return r.Endianness.Uint64(b)
}

// ReadInt8 reads a int8.
func (r *BinaryReader) ReadInt8() int8 {
	return int8(r.ReadByte())
}

// ReadInt16 reads a int16.
func (r *BinaryReader) ReadInt16() int16 {
	return int16(r.ReadUint16())
}

// ReadInt32 reads a int32.
func (r *BinaryReader) ReadInt32() int32 {
	return int32(r.ReadUint32())
}

// ReadInt64 reads a int64.
func (r *BinaryReader) ReadInt64() int64 {
	return int64(r.ReadUint64())
}

type BinaryFileReader struct {
	f      *os.File
	size   uint64
	offset uint64

	Endianness binary.ByteOrder
	buf        []byte
	pos        int
}

func NewBinaryFileReader(f *os.File, chunk int) (*BinaryFileReader, error) {
	var buf []byte
	var size uint64
	if chunk == 0 {
		var err error
		if buf, err = io.ReadAll(f); err != nil {
			return nil, err
		}
	} else {
		buf = make([]byte, 0, chunk)
	}
	if info, err := f.Stat(); err != nil {
		return nil, err
	} else {
		size = uint64(info.Size())
	}
	return &BinaryFileReader{
		f:          f,
		size:       size,
		Endianness: binary.BigEndian,
		buf:        buf,
	}, nil
}

func (r *BinaryFileReader) buffer(pos, length uint64) error {
	if pos < r.offset || r.offset+uint64(len(r.buf)) < pos+length {
		if math.MaxInt64 < pos {
			return fmt.Errorf("seek position too large")
		} else if _, err := r.f.Seek(int64(pos), 0); err != nil {
			return err
		} else if n, err := r.f.Read(r.buf[:cap(r.buf)]); err != nil {
			return err
		} else {
			r.offset = pos
			r.buf = r.buf[:n]
			r.pos = 0
		}
	}
	return nil
}

// Seek set the reader position in the buffer.
func (r *BinaryFileReader) Seek(pos uint64) error {
	if r.size <= pos {
		return io.EOF
	} else if err := r.buffer(pos, 0); err != nil {
		return err
	}
	r.pos = int(pos - r.offset)
	return nil
}

// Pos returns the reader's position.
func (r *BinaryFileReader) Pos() uint64 {
	return r.offset + uint64(r.pos)
}

// Len returns the remaining length of the buffer.
func (r *BinaryFileReader) Len() uint64 {
	return r.size - r.Pos()
}

// Offset returns the offset of the buffer.
func (r *BinaryFileReader) Offset() uint64 {
	return r.offset
}

// BufferLen returns the length of the buffer.
func (r *BinaryFileReader) BufferLen() int {
	return len(r.buf)
}

// Read complies with io.Reader.
func (r *BinaryFileReader) Read(b []byte) (int, error) {
	if len(b) <= cap(r.buf) {
		if err := r.buffer(r.offset+uint64(r.pos), uint64(len(b))); err != nil {
			return 0, err
		}
		n := copy(b, r.buf[r.pos:])
		r.pos += n
		return n, nil
	}

	// read directly from file
	if _, err := r.f.Seek(int64(r.offset)+int64(r.pos), 0); err != nil {
		return 0, err
	}
	n, err := r.f.Read(b)
	r.offset += uint64(r.pos + n)
	r.pos = 0
	r.buf = r.buf[:0]
	return n, err
}

// ReadBytes reads n bytes.
func (r *BinaryFileReader) ReadBytes(n int) []byte {
	if n < len(r.buf)-r.pos {
		b := r.buf[r.pos : r.pos+n]
		r.pos += n
		return b
	}

	b := make([]byte, n)
	if _, err := r.Read(b); err != nil {
		return nil
	}
	return b
}

// ReadString reads a string of length n.
func (r *BinaryFileReader) ReadString(n int) string {
	return string(r.ReadBytes(n))
}

// ReadByte reads a single byte.
func (r *BinaryFileReader) ReadByte() byte {
	b := r.ReadBytes(1)
	if b == nil {
		return 0
	}
	return b[0]
}

// ReadUint8 reads a uint8.
func (r *BinaryFileReader) ReadUint8() uint8 {
	return r.ReadByte()
}

// ReadUint16 reads a uint16.
func (r *BinaryFileReader) ReadUint16() uint16 {
	b := r.ReadBytes(2)
	if b == nil {
		return 0
	}
	return r.Endianness.Uint16(b)
}

// ReadUint32 reads a uint32.
func (r *BinaryFileReader) ReadUint32() uint32 {
	b := r.ReadBytes(4)
	if b == nil {
		return 0
	}
	return r.Endianness.Uint32(b)
}

// ReadUint64 reads a uint64.
func (r *BinaryFileReader) ReadUint64() uint64 {
	b := r.ReadBytes(8)
	if b == nil {
		return 0
	}
	return r.Endianness.Uint64(b)
}

// ReadInt8 reads a int8.
func (r *BinaryFileReader) ReadInt8() int8 {
	return int8(r.ReadByte())
}

// ReadInt16 reads a int16.
func (r *BinaryFileReader) ReadInt16() int16 {
	return int16(r.ReadUint16())
}

// ReadInt32 reads a int32.
func (r *BinaryFileReader) ReadInt32() int32 {
	return int32(r.ReadUint32())
}

// ReadInt64 reads a int64.
func (r *BinaryFileReader) ReadInt64() int64 {
	return int64(r.ReadUint64())
}

// BinaryWriter is a big endian binary file format writer.
type BinaryWriter struct {
	buf []byte
}

// NewBinaryWriter returns a big endian binary file format writer.
func NewBinaryWriter(buf []byte) *BinaryWriter {
	return &BinaryWriter{buf}
}

// Len returns the buffer's length in bytes.
func (w *BinaryWriter) Len() uint32 {
	return uint32(len(w.buf))
}

// Bytes returns the buffer's bytes.
func (w *BinaryWriter) Bytes() []byte {
	return w.buf
}

// Write complies with io.Writer.
func (w *BinaryWriter) Write(b []byte) (int, error) {
	w.buf = append(w.buf, b...)
	return len(b), nil
}

// WriteBytes writes the given bytes to the buffer.
func (w *BinaryWriter) WriteBytes(v []byte) {
	w.buf = append(w.buf, v...)
}

// WriteString writes the given string to the buffer.
func (w *BinaryWriter) WriteString(v string) {
	w.WriteBytes([]byte(v))
}

// WriteByte writes the given byte to the buffer.
func (w *BinaryWriter) WriteByte(v byte) {
	w.WriteBytes([]byte{v})
}

// WriteUint8 writes the given uint8 to the buffer.
func (w *BinaryWriter) WriteUint8(v uint8) {
	w.WriteByte(v)
}

// WriteUint16 writes the given uint16 to the buffer.
func (w *BinaryWriter) WriteUint16(v uint16) {
	pos := len(w.buf)
	w.buf = append(w.buf, make([]byte, 2)...)
	binary.BigEndian.PutUint16(w.buf[pos:], v)
}

// WriteUint32 writes the given uint32 to the buffer.
func (w *BinaryWriter) WriteUint32(v uint32) {
	pos := len(w.buf)
	w.buf = append(w.buf, make([]byte, 4)...)
	binary.BigEndian.PutUint32(w.buf[pos:], v)
}

// WriteUint64 writes the given uint64 to the buffer.
func (w *BinaryWriter) WriteUint64(v uint64) {
	pos := len(w.buf)
	w.buf = append(w.buf, make([]byte, 8)...)
	binary.BigEndian.PutUint64(w.buf[pos:], v)
}

// WriteInt8 writes the given int8 to the buffer.
func (w *BinaryWriter) WriteInt8(v int8) {
	w.WriteUint8(uint8(v))
}

// WriteInt16 writes the given int16 to the buffer.
func (w *BinaryWriter) WriteInt16(v int16) {
	w.WriteUint16(uint16(v))
}

// WriteInt32 writes the given int32 to the buffer.
func (w *BinaryWriter) WriteInt32(v int32) {
	w.WriteUint32(uint32(v))
}

// WriteInt64 writes the given int64 to the buffer.
func (w *BinaryWriter) WriteInt64(v int64) {
	w.WriteUint64(uint64(v))
}

// BitmapReader is a binary bitmap reader.
type BitmapReader struct {
	buf []byte
	pos uint32 // TODO: to uint64
	eof bool
}

// NewBitmapReader returns a binary bitmap reader.
func NewBitmapReader(buf []byte) *BitmapReader {
	return &BitmapReader{buf, 0, false}
}

// Pos returns the current bit position.
func (r *BitmapReader) Pos() uint32 {
	return r.pos
}

// EOF returns if we reached the buffer's end-of-file.
func (r *BitmapReader) EOF() bool {
	return r.eof
}

// Read reads the next bit.
func (r *BitmapReader) Read() bool {
	if r.eof || uint32(len(r.buf)) <= (r.pos+1)/8 {
		r.eof = true
		return false
	}
	bit := r.buf[r.pos>>3]&(0x80>>(r.pos&7)) != 0
	r.pos += 1
	return bit
}

// BitmapWriter is a binary bitmap writer.
type BitmapWriter struct {
	buf []byte
	pos uint32
}

// NewBitmapWriter returns a binary bitmap writer.
func NewBitmapWriter(buf []byte) *BitmapWriter {
	return &BitmapWriter{buf, 0}
}

// Len returns the buffer's length in bytes.
func (w *BitmapWriter) Len() uint32 {
	return uint32(len(w.buf))
}

// Bytes returns the buffer's bytes.
func (w *BitmapWriter) Bytes() []byte {
	return w.buf
}

// Write writes the next bit.
func (w *BitmapWriter) Write(bit bool) {
	if uint32(len(w.buf)) <= (w.pos+1)/8 {
		w.buf = append(w.buf, 0)
	}
	if bit {
		w.buf[w.pos>>3] = w.buf[w.pos>>3] | (0x80 >> (w.pos & 7))
	}
	w.pos += 1
}
