package pngstructure

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"encoding/binary"
	"hash/crc32"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	riimage "github.com/dsoprea/go-utility/v2/image"
)

var (
	PngSignature  = [8]byte{137, 'P', 'N', 'G', '\r', '\n', 26, '\n'}
	EXifChunkType = "eXIf"
	IHDRChunkType = "IHDR"
)

var (
	ErrNotPng     = errors.New("not png data")
	ErrCrcFailure = errors.New("crc failure")
)

// ChunkSlice encapsulates a slice of chunks.
type ChunkSlice struct {
	chunks []*Chunk
}

func NewChunkSlice(chunks []*Chunk) (*ChunkSlice, error) {
	if len(chunks) == 0 {
		err := errors.New("ChunkSlice must be initialized with at least one chunk (IHDR)")
		return nil, err
	} else if chunks[0].Type != IHDRChunkType {
		err := errors.New("first chunk in any ChunkSlice must be an IHDR")
		return nil, err
	}

	return &ChunkSlice{chunks}, nil
}

func NewPngChunkSlice() (*ChunkSlice, error) {
	ihdrChunk := &Chunk{
		Type: IHDRChunkType,
	}

	ihdrChunk.UpdateCrc32()

	return NewChunkSlice([]*Chunk{ihdrChunk})
}

func (cs *ChunkSlice) String() string {
	return fmt.Sprintf("ChunkSlize<LEN=(%d)>", len(cs.chunks))
}

// Chunks exposes the actual slice.
func (cs *ChunkSlice) Chunks() []*Chunk {
	return cs.chunks
}

// Write encodes and writes all chunks.
func (cs *ChunkSlice) WriteTo(w io.Writer) error {
	if _, err := w.Write(PngSignature[:]); err != nil {
		return err
	}

	// TODO(dustin): !! This should respect
	// the safe-to-copy characteristic.
	for _, c := range cs.chunks {
		if _, err := c.WriteTo(w); err != nil {
			return err
		}
	}

	return nil
}

// Index returns a map of chunk types to chunk slices, grouping all like chunks.
func (cs *ChunkSlice) Index() (index map[string][]*Chunk) {
	index = make(map[string][]*Chunk)
	for _, c := range cs.chunks {
		if grouped, found := index[c.Type]; found {
			index[c.Type] = append(grouped, c)
		} else {
			index[c.Type] = []*Chunk{c}
		}
	}

	return index
}

// FindExif returns the the segment that hosts the EXIF data.
func (cs *ChunkSlice) FindExif() (chunk *Chunk, err error) {
	index := cs.Index()
	if chunks, found := index[EXifChunkType]; found {
		return chunks[0], nil
	}

	return nil, exif.ErrNoExif
}

// Exif returns an `exif.Ifd` instance with the existing tags.
func (cs *ChunkSlice) Exif() (*exif.Ifd, []byte, error) {
	chunk, err := cs.FindExif()
	if err != nil {
		return nil, nil, err
	}

	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return nil, nil, err
	}

	ti := exif.NewTagIndex()

	_, index, err := exif.Collect(im, ti, chunk.Data)
	if err != nil {
		return nil, nil, err
	}

	return index.RootIfd, chunk.Data, nil
}

// ConstructExifBuilder returns an `exif.IfdBuilder` instance
// (needed for modifying) preloaded with all existing tags.
func (cs *ChunkSlice) ConstructExifBuilder() (*exif.IfdBuilder, error) {
	rootIfd, _, err := cs.Exif()
	if err != nil {
		return nil, err
	}

	return exif.NewIfdBuilderFromExistingChain(rootIfd), nil
}

// SetExif encodes and sets EXIF data into this segment.
func (cs *ChunkSlice) SetExif(ib *exif.IfdBuilder) error {
	// Encode.

	ibe := exif.NewIfdByteEncoder()

	exifData, err := ibe.EncodeToExif(ib)
	if err != nil {
		return err
	}

	// Set.

	exifChunk, err := cs.FindExif()

	switch {
	case err == nil:
		// EXIF chunk already exists.
		exifChunk.Data = exifData
		exifChunk.Length = uint32(len(exifData))

	case errors.Is(err, exif.ErrNoExif):
		// Add a EXIF chunk for the first time.
		exifChunk = &Chunk{
			Type:   EXifChunkType,
			Data:   exifData,
			Length: uint32(len(exifData)),
		}

		// Insert exif after the IHDR chunk; it's
		// a reliably appropriate place to put it.
		cs.chunks = append(
			cs.chunks[:1],
			append(
				[]*Chunk{exifChunk},
				cs.chunks[1:]...,
			)...,
		)

	default:
		return err
	}

	exifChunk.UpdateCrc32()
	return nil
}

// PngSplitter hosts the princpal `Split()`
// method uses by `bufio.Scanner`.
type PngSplitter struct {
	chunks        []*Chunk
	currentOffset int

	doCheckCrc bool
	crcErrors  []string
}

func (ps *PngSplitter) Chunks() (*ChunkSlice, error) {
	return NewChunkSlice(ps.chunks)
}

func (ps *PngSplitter) DoCheckCrc(doCheck bool) {
	ps.doCheckCrc = doCheck
}

func (ps *PngSplitter) CrcErrors() []string {
	return ps.crcErrors
}

func NewPngSplitter() *PngSplitter {
	return &PngSplitter{
		chunks:     make([]*Chunk, 0),
		doCheckCrc: true,
		crcErrors:  make([]string, 0),
	}
}

// Chunk describes a single chunk.
type Chunk struct {
	Offset int
	Length uint32
	Type   string
	Data   []byte
	Crc    uint32
}

func (c *Chunk) String() string {
	return fmt.Sprintf("Chunk<OFFSET=(%d) LENGTH=(%d) TYPE=[%s] CRC=(%d)>", c.Offset, c.Length, c.Type, c.Crc)
}

func calculateCrc32(chunk *Chunk) uint32 {
	c := crc32.NewIEEE()

	c.Write([]byte(chunk.Type))
	c.Write(chunk.Data)

	return c.Sum32()
}

func (c *Chunk) UpdateCrc32() {
	c.Crc = calculateCrc32(c)
}

func (c *Chunk) CheckCrc32() bool {
	expected := calculateCrc32(c)
	return c.Crc == expected
}

// Bytes encodes and returns the bytes for this chunk.
func (c *Chunk) Bytes() ([]byte, error) {
	if len(c.Data) != int(c.Length) {
		return nil, errors.New("length of data not correct")
	}

	preallocated := make([]byte, 0, 4+4+c.Length+4)
	b := bytes.NewBuffer(preallocated)

	err := binary.Write(b, binary.BigEndian, c.Length)
	if err != nil {
		return nil, err
	}

	if _, err := b.Write([]byte(c.Type)); err != nil {
		return nil, err
	}

	if c.Data != nil {
		if _, err := b.Write(c.Data); err != nil {
			return nil, err
		}
	}

	if err := binary.Write(b, binary.BigEndian, c.Crc); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// Write encodes and writes the bytes for this chunk.
func (c *Chunk) WriteTo(w io.Writer) (int, error) {
	if len(c.Data) != int(c.Length) {
		return 0, errors.New("length of data not correct")
	}

	if err := binary.Write(w, binary.BigEndian, c.Length); err != nil {
		return 0, err
	}

	if _, err := w.Write([]byte(c.Type)); err != nil {
		return 0, err
	}

	if _, err := w.Write(c.Data); err != nil {
		return 0, err
	}

	if err := binary.Write(w, binary.BigEndian, c.Crc); err != nil {
		return 0, err
	}

	return 4 + len(c.Type) + len(c.Data) + 4, nil
}

// readHeader verifies that the PNG header bytes appear next.
func (ps *PngSplitter) readHeader(r io.Reader) error {
	var (
		sigLen = len(PngSignature)
		header = make([]byte, sigLen)
	)

	if _, err := r.Read(header); err != nil {
		return err
	}

	ps.currentOffset += sigLen
	if !bytes.Equal(header, PngSignature[:]) {
		return ErrNotPng
	}

	return nil
}

// Split fulfills the `bufio.SplitFunc`
// function definition for `bufio.Scanner`.
func (ps *PngSplitter) Split(
	data []byte,
	atEOF bool,
) (
	advance int,
	token []byte,
	err error,
) {
	// We might have more than one chunk's worth, and,
	// if `atEOF` is true, we won't be called again.
	// We'll repeatedly try to read additional chunks,
	// but, when we run out of the data we were given
	// then we'll return the number of bytes for the
	// chunks we've already completely read. Then, we'll
	// be called again from the end ofthose bytes, at
	// which point we'll indicate that we don't yet have
	// enough for another chunk, and we should be then
	// called with more.
	for {
		len_ := len(data)
		if len_ < 8 {
			return advance, nil, nil
		}

		length := binary.BigEndian.Uint32(data[:4])
		type_ := string(data[4:8])
		chunkSize := (8 + int(length) + 4)

		if len_ < chunkSize {
			return advance, nil, nil
		}

		crcIndex := 8 + length
		crc := binary.BigEndian.Uint32(data[crcIndex : crcIndex+4])

		content := make([]byte, length)
		copy(content, data[8:8+length])

		c := &Chunk{
			Length: length,
			Type:   type_,
			Data:   content,
			Crc:    crc,
			Offset: ps.currentOffset,
		}

		ps.chunks = append(ps.chunks, c)

		if !c.CheckCrc32() {
			ps.crcErrors = append(ps.crcErrors, type_)

			if ps.doCheckCrc {
				err = ErrCrcFailure
				return
			}
		}

		advance += chunkSize
		ps.currentOffset += chunkSize

		data = data[chunkSize:]
	}
}

var (
	// Enforce interface conformance.
	_ riimage.MediaContext = new(ChunkSlice)
)
