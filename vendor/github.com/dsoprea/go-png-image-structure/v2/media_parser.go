package pngstructure

import (
	"bufio"
	"bytes"
	"image"
	"io"
	"os"

	"image/png"

	"github.com/dsoprea/go-logging"
	"github.com/dsoprea/go-utility/v2/image"
)

// PngMediaParser knows how to parse a PNG stream.
type PngMediaParser struct {
}

// NewPngMediaParser returns a new `PngMediaParser` struct.
func NewPngMediaParser() *PngMediaParser {

	// TODO(dustin): Add test

	return new(PngMediaParser)
}

// Parse parses a PNG stream given a `io.ReadSeeker`.
func (pmp *PngMediaParser) Parse(rs io.ReadSeeker, size int) (mc riimage.MediaContext, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): Add test

	ps := NewPngSplitter()

	err = ps.readHeader(rs)
	log.PanicIf(err)

	s := bufio.NewScanner(rs)

	// Since each segment can be any size, our buffer must be allowed to grow
	// as large as the file.
	buffer := []byte{}
	s.Buffer(buffer, size)
	s.Split(ps.Split)

	for s.Scan() != false {
	}

	log.PanicIf(s.Err())

	return ps.Chunks(), nil
}

// ParseFile parses a PNG stream given a file-path.
func (pmp *PngMediaParser) ParseFile(filepath string) (mc riimage.MediaContext, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	stat, err := f.Stat()
	log.PanicIf(err)

	size := stat.Size()

	chunks, err := pmp.Parse(f, int(size))
	log.PanicIf(err)

	return chunks, nil
}

// ParseBytes parses a PNG stream given a byte-slice.
func (pmp *PngMediaParser) ParseBytes(data []byte) (mc riimage.MediaContext, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): Add test

	br := bytes.NewReader(data)

	chunks, err := pmp.Parse(br, len(data))
	log.PanicIf(err)

	return chunks, nil
}

// LooksLikeFormat returns a boolean indicating whether the stream looks like a
// PNG image.
func (pmp *PngMediaParser) LooksLikeFormat(data []byte) bool {
	return bytes.Compare(data[:len(PngSignature)], PngSignature[:]) == 0
}

// GetImage returns an image.Image-compatible struct.
func (pmp *PngMediaParser) GetImage(r io.Reader) (img image.Image, err error) {
	img, err = png.Decode(r)
	log.PanicIf(err)

	return img, nil
}

var (
	// Enforce interface conformance.
	_ riimage.MediaParser = new(PngMediaParser)
)
