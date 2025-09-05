package parse

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

const PageSize = 4096

type IBinaryReader interface {
	Bytes([]byte, int64, int64) ([]byte, error)
	Len() int64
	Close() error
}

type binaryReaderFile struct {
	f    *os.File
	size int64
}

func newBinaryReaderFile(filename string) (*binaryReaderFile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	return &binaryReaderFile{f, fi.Size()}, nil
}

// Close closes the reader.
func (r *binaryReaderFile) Close() error {
	return r.f.Close()
}

// Len returns the length of the underlying memory-mapped file.
func (r *binaryReaderFile) Len() int64 {
	return r.size
}

func (r *binaryReaderFile) Bytes(b []byte, n, off int64) ([]byte, error) {
	if _, err := r.f.Seek(off, 0); err != nil {
		return nil, err
	} else if b == nil {
		b = make([]byte, n)
	}

	m, err := r.f.Read(b)
	if err != nil {
		return nil, err
	} else if int64(m) != n {
		return nil, errors.New("file: could not read all bytes")
	}
	return b, nil
}

type binaryReaderBytes struct {
	data []byte
}

func newBinaryReaderBytes(data []byte) *binaryReaderBytes {
	return &binaryReaderBytes{data}
}

// Close closes the reader.
func (r *binaryReaderBytes) Close() error {
	return nil
}

// Len returns the length of the underlying memory-mapped file.
func (r *binaryReaderBytes) Len() int64 {
	return int64(len(r.data))
}

func (r *binaryReaderBytes) Bytes(b []byte, n, off int64) ([]byte, error) {
	if off < 0 || n < 0 || int64(len(r.data)) < off || int64(len(r.data))-off < n {
		return nil, fmt.Errorf("bytes: invalid range %d--%d", off, off+n)
	}

	data := r.data[off : off+n : off+n]
	if b == nil {
		return data, nil
	}
	copy(b, data)
	return b, nil
}

type binaryReaderReader struct {
	r        io.Reader
	size     int64
	readerAt bool
	seeker   bool
}

func newBinaryReaderReader(r io.Reader, n int64) *binaryReaderReader {
	_, readerAt := r.(io.ReaderAt)
	_, seeker := r.(io.Seeker)
	return &binaryReaderReader{r, n, readerAt, seeker}
}

// Close closes the reader.
func (r *binaryReaderReader) Close() error {
	if closer, ok := r.r.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// Len returns the length of the underlying memory-mapped file.
func (r *binaryReaderReader) Len() int64 {
	return r.size
}

func (r *binaryReaderReader) Bytes(b []byte, n, off int64) ([]byte, error) {
	if b == nil {
		b = make([]byte, n)
	}

	// seeker seems faster than readerAt by 10%
	if r.seeker {
		if _, err := r.r.(io.Seeker).Seek(off, 0); err != nil {
			return nil, err
		}

		m, err := r.r.Read(b)
		if err != nil {
			return nil, err
		} else if int64(m) != n {
			return nil, errors.New("file: could not read all bytes")
		}
		return b, nil
	} else if r.readerAt {
		m, err := r.r.(io.ReaderAt).ReadAt(b, off)
		if err != nil {
			return nil, err
		} else if int64(m) != n {
			return nil, errors.New("file: could not read all bytes")
		}
		return b, nil
	}
	return nil, errors.New("io.Seeker and io.ReaderAt not implemented")
}

type BinaryReader struct {
	f   IBinaryReader
	pos int64
	err error

	ByteOrder binary.ByteOrder
}

func NewBinaryReader(f IBinaryReader) *BinaryReader {
	return &BinaryReader{
		f:         f,
		ByteOrder: binary.BigEndian,
	}
}

func NewBinaryReaderReader(r io.Reader, n int64) (*BinaryReader, error) {
	_, isReaderAt := r.(io.ReaderAt)
	_, isSeeker := r.(io.Seeker)

	var f IBinaryReader
	if isReaderAt || isSeeker {
		f = newBinaryReaderReader(r, n)
	} else {
		b := make([]byte, n)
		if _, err := io.ReadFull(r, b); err != nil {
			return nil, err
		}
		f = newBinaryReaderBytes(b)
	}
	return NewBinaryReader(f), nil
}

func NewBinaryReaderBytes(data []byte) *BinaryReader {
	f := newBinaryReaderBytes(data)
	return NewBinaryReader(f)
}

func NewBinaryReaderFile(filename string) (*BinaryReader, error) {
	f, err := newBinaryReaderFile(filename)
	if err != nil {
		return nil, err
	}
	return NewBinaryReader(f), nil
}

func (r *BinaryReader) IBinaryReader() IBinaryReader {
	return r.f
}

func (r *BinaryReader) Clone() *BinaryReader {
	f := r.f
	if cloner, ok := f.(interface{ Clone() IBinaryReader }); ok {
		f = cloner.Clone()
	}
	return &BinaryReader{
		f:         f,
		pos:       r.pos,
		err:       r.err,
		ByteOrder: r.ByteOrder,
	}
}

func (r *BinaryReader) Err() error {
	return r.err
}

func (r *BinaryReader) Close() error {
	if err := r.f.Close(); err != nil {
		return err
	}
	return r.err
}

// InPageCache returns true if the range is already in the page cache (for mmap).
func (r *BinaryReader) InPageCache(start, end int64) bool {
	index := r.Pos() / PageSize
	return start/PageSize == index && end/PageSize == index
}

// Pos returns the reader's position.
func (r *BinaryReader) Pos() int64 {
	return r.pos
}

// Len returns the remaining length of the buffer.
func (r *BinaryReader) Len() int64 {
	return r.f.Len() - r.pos
}

// Seek complies with io.Seeker.
func (r *BinaryReader) Seek(off int64, whence int) (int64, error) {
	if whence == 0 {
		if off < 0 || r.f.Len() < off {
			return 0, fmt.Errorf("invalid offset")
		}
		r.pos = off
	} else if whence == 1 {
		if r.pos+off < 0 || r.f.Len() < r.pos+off {
			return 0, fmt.Errorf("invalid offset")
		}
		r.pos += off
	} else if whence == 2 {
		if off < -r.f.Len() || 0 < off {
			return 0, fmt.Errorf("invalid offset")
		}
		r.pos = r.f.Len() - off
	} else {
		return 0, fmt.Errorf("invalid whence")
	}
	return r.pos, nil
}

// Read complies with io.Reader.
func (r *BinaryReader) Read(b []byte) (int, error) {
	data, err := r.f.Bytes(b, int64(len(b)), r.pos)
	if err != nil && err != io.EOF {
		return 0, err
	}
	r.pos += int64(len(data))
	return len(data), err
}

// ReadAt complies with io.ReaderAt.
func (r *BinaryReader) ReadAt(b []byte, off int64) (int, error) {
	data, err := r.f.Bytes(b, int64(len(b)), off)
	if err != nil && err != io.EOF {
		return 0, err
	}
	return len(data), err
}

// ReadBytes reads n bytes.
func (r *BinaryReader) ReadBytes(n int64) []byte {
	data, err := r.f.Bytes(nil, n, r.pos)
	if err != nil {
		r.err = err
		return nil
	}
	r.pos += n
	return data
}

// ReadString reads a string of length n.
func (r *BinaryReader) ReadString(n int64) string {
	return string(r.ReadBytes(n))
}

// ReadByte reads a single byte.
func (r *BinaryReader) ReadByte() (byte, error) {
	data := r.ReadBytes(1)
	if data == nil {
		return 0, r.err
	}
	return data[0], nil
}

// ReadUint8 reads a uint8.
func (r *BinaryReader) ReadUint8() uint8 {
	data := r.ReadBytes(1)
	if data == nil {
		return 0
	}
	return data[0]
}

// ReadUint16 reads a uint16.
func (r *BinaryReader) ReadUint16() uint16 {
	data := r.ReadBytes(2)
	if data == nil {
		return 0
	} else if r.ByteOrder == binary.LittleEndian {
		return uint16(data[1])<<8 | uint16(data[0])
	}
	return uint16(data[0])<<8 | uint16(data[1])
}

// ReadUint24 reads a uint24 into a uint32.
func (r *BinaryReader) ReadUint24() uint32 {
	b := r.ReadBytes(3)
	if b == nil {
		return 0
	} else if r.ByteOrder == binary.LittleEndian {
		return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
	} else {
		return uint32(b[2]) | uint32(b[1])<<8 | uint32(b[0])<<16
	}
}

// ReadUint32 reads a uint32.
func (r *BinaryReader) ReadUint32() uint32 {
	data := r.ReadBytes(4)
	if data == nil {
		return 0
	} else if r.ByteOrder == binary.LittleEndian {
		return uint32(data[3])<<24 | uint32(data[2])<<16 | uint32(data[1])<<8 | uint32(data[0])
	}
	return uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
}

// ReadUint64 reads a uint64.
func (r *BinaryReader) ReadUint64() uint64 {
	data := r.ReadBytes(8)
	if data == nil {
		return 0
	} else if r.ByteOrder == binary.LittleEndian {
		return uint64(data[7])<<56 | uint64(data[6])<<48 | uint64(data[5])<<40 | uint64(data[4])<<32 | uint64(data[3])<<24 | uint64(data[2])<<16 | uint64(data[1])<<8 | uint64(data[0])
	}
	return uint64(data[0])<<56 | uint64(data[1])<<48 | uint64(data[2])<<40 | uint64(data[3])<<32 | uint64(data[4])<<24 | uint64(data[5])<<16 | uint64(data[6])<<8 | uint64(data[7])
}

// ReadInt8 reads a int8.
func (r *BinaryReader) ReadInt8() int8 {
	return int8(r.ReadUint8())
}

// ReadInt16 reads a int16.
func (r *BinaryReader) ReadInt16() int16 {
	return int16(r.ReadUint16())
}

// ReadInt24 reads a int24 into an int32.
func (r *BinaryReader) ReadInt24() int32 {
	return int32(r.ReadUint24())
}

// ReadInt32 reads a int32.
func (r *BinaryReader) ReadInt32() int32 {
	return int32(r.ReadUint32())
}

// ReadInt64 reads a int64.
func (r *BinaryReader) ReadInt64() int64 {
	return int64(r.ReadUint64())
}

// BinaryWriter is a big endian binary file format writer.
type BinaryWriter struct {
	buf       []byte
	ByteOrder binary.AppendByteOrder
}

// NewBinaryWriter returns a big endian binary file format writer.
func NewBinaryWriter(buf []byte) *BinaryWriter {
	return &BinaryWriter{
		buf:       buf,
		ByteOrder: binary.BigEndian,
	}
}

// Len returns the buffer's length in bytes.
func (w *BinaryWriter) Len() int64 {
	return int64(len(w.buf))
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
	w.buf = w.ByteOrder.AppendUint16(w.buf, v)
}

// WriteUint24 writes the given uint32 as a uint24 to the buffer.
func (w *BinaryWriter) WriteUint24(v uint32) {
	if w.ByteOrder == binary.LittleEndian {
		w.buf = append(w.buf, byte(v), byte(v>>8), byte(v>>16))
	} else {
		w.buf = append(w.buf, byte(v>>16), byte(v>>8), byte(v))
	}
}

// WriteUint32 writes the given uint32 to the buffer.
func (w *BinaryWriter) WriteUint32(v uint32) {
	w.buf = w.ByteOrder.AppendUint32(w.buf, v)
}

// WriteUint64 writes the given uint64 to the buffer.
func (w *BinaryWriter) WriteUint64(v uint64) {
	w.buf = w.ByteOrder.AppendUint64(w.buf, v)
}

// WriteInt8 writes the given int8 to the buffer.
func (w *BinaryWriter) WriteInt8(v int8) {
	w.WriteUint8(uint8(v))
}

// WriteInt16 writes the given int16 to the buffer.
func (w *BinaryWriter) WriteInt16(v int16) {
	w.WriteUint16(uint16(v))
}

// WriteInt24 writes the given int32 as an in24 to the buffer.
func (w *BinaryWriter) WriteInt24(v int32) {
	w.WriteUint24(uint32(v))
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
	pos uint64
}

// NewBitmapWriter returns a binary bitmap writer.
func NewBitmapWriter(buf []byte) *BitmapWriter {
	return &BitmapWriter{buf, 0}
}

// Len returns the buffer's length in bytes.
func (w *BitmapWriter) Len() int64 {
	return int64(len(w.buf))
}

// Bytes returns the buffer's bytes.
func (w *BitmapWriter) Bytes() []byte {
	return w.buf
}

// Write writes the next bit.
func (w *BitmapWriter) Write(bit bool) {
	if uint64(len(w.buf)) <= (w.pos+1)/8 {
		w.buf = append(w.buf, 0)
	}
	if bit {
		w.buf[w.pos>>3] = w.buf[w.pos>>3] | (0x80 >> (w.pos & 7))
	}
	w.pos += 1
}
