package pngstructure

import (
	"bufio"
	"bytes"
	"image"
	"io"
	"os"

	"image/png"

	riimage "github.com/dsoprea/go-utility/v2/image"
)

// PngMediaParser knows how to parse a PNG stream.
type PngMediaParser struct {
}

// NewPngMediaParser returns a new `PngMediaParser`.
func NewPngMediaParser() riimage.MediaParser {
	return new(PngMediaParser)
}

// Parse parses a PNG stream given a `io.ReadSeeker`.
func (pmp *PngMediaParser) Parse(
	rs io.ReadSeeker,
	size int,
) (riimage.MediaContext, error) {
	ps := NewPngSplitter()
	if err := ps.readHeader(rs); err != nil {
		return nil, err
	}

	s := bufio.NewScanner(rs)

	// Since each segment can be any
	// size, our buffer must be allowed
	// to grow as large as the file.
	buffer := []byte{}
	s.Buffer(buffer, size)
	s.Split(ps.Split)

	for s.Scan() {
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	return ps.Chunks()
}

// ParseFile parses a PNG stream given a file-path.
func (pmp *PngMediaParser) ParseFile(filepath string) (riimage.MediaContext, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := stat.Size()
	return pmp.Parse(f, int(size))
}

// ParseBytes parses a PNG stream given a byte-slice.
func (pmp *PngMediaParser) ParseBytes(data []byte) (riimage.MediaContext, error) {
	br := bytes.NewReader(data)
	return pmp.Parse(br, len(data))
}

// LooksLikeFormat returns a boolean indicating
// whether the stream looks like a PNG image.
func (pmp *PngMediaParser) LooksLikeFormat(data []byte) bool {
	return bytes.Equal(data[:len(PngSignature)], PngSignature[:])
}

// GetImage returns an image.Image-compatible struct.
func (pmp *PngMediaParser) GetImage(r io.Reader) (img image.Image, err error) {
	return png.Decode(r)
}
