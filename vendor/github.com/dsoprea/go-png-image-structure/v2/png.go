package pngstructure

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"encoding/binary"
	"hash/crc32"

	"github.com/dsoprea/go-exif/v3"
	"github.com/dsoprea/go-exif/v3/common"
	"github.com/dsoprea/go-logging"
	"github.com/dsoprea/go-utility/v2/image"
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

func NewChunkSlice(chunks []*Chunk) *ChunkSlice {
	if len(chunks) == 0 {
		log.Panicf("ChunkSlice must be initialized with at least one chunk (IHDR)")
	} else if chunks[0].Type != IHDRChunkType {
		log.Panicf("first chunk in any ChunkSlice must be an IHDR")
	}

	return &ChunkSlice{
		chunks: chunks,
	}
}

func NewPngChunkSlice() *ChunkSlice {

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
func (cs *ChunkSlice) WriteTo(w io.Writer) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	_, err = w.Write(PngSignature[:])
	log.PanicIf(err)

	// TODO(dustin): !! This should respect the safe-to-copy characteristic.
	for _, c := range cs.chunks {
		_, err := c.WriteTo(w)
		log.PanicIf(err)
	}

	return nil
}

// Index returns a map of chunk types to chunk slices, grouping all like chunks.
func (cs *ChunkSlice) Index() (index map[string][]*Chunk) {
	index = make(map[string][]*Chunk)
	for _, c := range cs.chunks {
		if grouped, found := index[c.Type]; found == true {
			index[c.Type] = append(grouped, c)
		} else {
			index[c.Type] = []*Chunk{c}
		}
	}

	return index
}

// FindExif returns the the segment that hosts the EXIF data.
func (cs *ChunkSlice) FindExif() (chunk *Chunk, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	index := cs.Index()

	if chunks, found := index[EXifChunkType]; found == true {
		return chunks[0], nil
	}

	log.Panic(exif.ErrNoExif)

	// Never called.
	return nil, nil
}

// Exif returns an `exif.Ifd` instance with the existing tags.
func (cs *ChunkSlice) Exif() (rootIfd *exif.Ifd, data []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	chunk, err := cs.FindExif()
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := exif.NewTagIndex()

	// TODO(dustin): Refactor and support `exif.GetExifData()`.

	_, index, err := exif.Collect(im, ti, chunk.Data)
	log.PanicIf(err)

	return index.RootIfd, chunk.Data, nil
}

// ConstructExifBuilder returns an `exif.IfdBuilder` instance (needed for
// modifying) preloaded with all existing tags.
func (cs *ChunkSlice) ConstructExifBuilder() (rootIb *exif.IfdBuilder, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	rootIfd, _, err := cs.Exif()
	log.PanicIf(err)

	ib := exif.NewIfdBuilderFromExistingChain(rootIfd)

	return ib, nil
}

// SetExif encodes and sets EXIF data into this segment.
func (cs *ChunkSlice) SetExif(ib *exif.IfdBuilder) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// Encode.

	ibe := exif.NewIfdByteEncoder()

	exifData, err := ibe.EncodeToExif(ib)
	log.PanicIf(err)

	// Set.

	exifChunk, err := cs.FindExif()
	if err == nil {
		// EXIF chunk already exists.

		exifChunk.Data = exifData
		exifChunk.Length = uint32(len(exifData))
	} else {
		if log.Is(err, exif.ErrNoExif) != true {
			log.Panic(err)
		}

		// Add a EXIF chunk for the first time.

		exifChunk = &Chunk{
			Type:   EXifChunkType,
			Data:   exifData,
			Length: uint32(len(exifData)),
		}

		// Insert it after the IHDR chunk (it's a reliably appropriate place to
		// put it).
		cs.chunks = append(cs.chunks[:1], append([]*Chunk{exifChunk}, cs.chunks[1:]...)...)
	}

	exifChunk.UpdateCrc32()

	return nil
}

// PngSplitter hosts the princpal `Split()` method uses by `bufio.Scanner`.
type PngSplitter struct {
	chunks        []*Chunk
	currentOffset int

	doCheckCrc bool
	crcErrors  []string
}

func (ps *PngSplitter) Chunks() *ChunkSlice {
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
func (c *Chunk) Bytes() []byte {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	if len(c.Data) != int(c.Length) {
		log.Panicf("length of data not correct")
	}

	preallocated := make([]byte, 0, 4+4+c.Length+4)
	b := bytes.NewBuffer(preallocated)

	err := binary.Write(b, binary.BigEndian, c.Length)
	log.PanicIf(err)

	_, err = b.Write([]byte(c.Type))
	log.PanicIf(err)

	if c.Data != nil {
		_, err = b.Write(c.Data)
		log.PanicIf(err)
	}

	err = binary.Write(b, binary.BigEndian, c.Crc)
	log.PanicIf(err)

	return b.Bytes()
}

// Write encodes and writes the bytes for this chunk.
func (c *Chunk) WriteTo(w io.Writer) (count int, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if len(c.Data) != int(c.Length) {
		log.Panicf("length of data not correct")
	}

	err = binary.Write(w, binary.BigEndian, c.Length)
	log.PanicIf(err)

	_, err = w.Write([]byte(c.Type))
	log.PanicIf(err)

	_, err = w.Write(c.Data)
	log.PanicIf(err)

	err = binary.Write(w, binary.BigEndian, c.Crc)
	log.PanicIf(err)

	return 4 + len(c.Type) + len(c.Data) + 4, nil
}

// readHeader verifies that the PNG header bytes appear next.
func (ps *PngSplitter) readHeader(r io.Reader) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	len_ := len(PngSignature)
	header := make([]byte, len_)

	_, err = r.Read(header)
	log.PanicIf(err)

	ps.currentOffset += len_

	if bytes.Compare(header, PngSignature[:]) != 0 {
		log.Panic(ErrNotPng)
	}

	return nil
}

// Split fulfills the `bufio.SplitFunc` function definition for
// `bufio.Scanner`.
func (ps *PngSplitter) Split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// We might have more than one chunk's worth, and, if `atEOF` is true, we
	// won't be called again. We'll repeatedly try to read additional chunks,
	// but, when we run out of the data we were given then we'll return the
	// number of bytes fo rthe chunks we've already completely read. Then,
	// we'll be called again from theend ofthose bytes, at which point we'll
	// indicate that we don't yet have enough for another chunk, and we should
	// be then called with more.
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

		if c.CheckCrc32() == false {
			ps.crcErrors = append(ps.crcErrors, type_)

			if ps.doCheckCrc == true {
				log.Panic(ErrCrcFailure)
			}
		}

		advance += chunkSize
		ps.currentOffset += chunkSize

		data = data[chunkSize:]
	}

	return advance, nil, nil
}

var (
	// Enforce interface conformance.
	_ riimage.MediaContext = new(ChunkSlice)
)
