package storage

import (
	"compress/gzip"
	"compress/zlib"
	"io"

	"git.iim.gay/grufwub/go-store/util"
	"github.com/golang/snappy"
)

// Compressor defines a means of compressing/decompressing values going into a key-value store
type Compressor interface {
	// Reader returns a new decompressing io.ReadCloser based on supplied (compressed) io.Reader
	Reader(io.Reader) (io.ReadCloser, error)

	// Writer returns a new compressing io.WriteCloser based on supplied (uncompressed) io.Writer
	Writer(io.Writer) (io.WriteCloser, error)
}

type gzipCompressor struct {
	level int
}

// GZipCompressor returns a new Compressor that implements GZip at default compression level
func GZipCompressor() Compressor {
	return GZipCompressorLevel(gzip.DefaultCompression)
}

// GZipCompressorLevel returns a new Compressor that implements GZip at supplied compression level
func GZipCompressorLevel(level int) Compressor {
	return &gzipCompressor{
		level: level,
	}
}

func (c *gzipCompressor) Reader(r io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}

func (c *gzipCompressor) Writer(w io.Writer) (io.WriteCloser, error) {
	return gzip.NewWriterLevel(w, c.level)
}

type zlibCompressor struct {
	level int
	dict  []byte
}

// ZLibCompressor returns a new Compressor that implements ZLib at default compression level
func ZLibCompressor() Compressor {
	return ZLibCompressorLevelDict(zlib.DefaultCompression, nil)
}

// ZLibCompressorLevel returns a new Compressor that implements ZLib at supplied compression level
func ZLibCompressorLevel(level int) Compressor {
	return ZLibCompressorLevelDict(level, nil)
}

// ZLibCompressorLevelDict returns a new Compressor that implements ZLib at supplied compression level with supplied dict
func ZLibCompressorLevelDict(level int, dict []byte) Compressor {
	return &zlibCompressor{
		level: level,
		dict:  dict,
	}
}

func (c *zlibCompressor) Reader(r io.Reader) (io.ReadCloser, error) {
	return zlib.NewReaderDict(r, c.dict)
}

func (c *zlibCompressor) Writer(w io.Writer) (io.WriteCloser, error) {
	return zlib.NewWriterLevelDict(w, c.level, c.dict)
}

type snappyCompressor struct{}

// SnappyCompressor returns a new Compressor that implements Snappy
func SnappyCompressor() Compressor {
	return &snappyCompressor{}
}

func (c *snappyCompressor) Reader(r io.Reader) (io.ReadCloser, error) {
	return util.NopReadCloser(snappy.NewReader(r)), nil
}

func (c *snappyCompressor) Writer(w io.Writer) (io.WriteCloser, error) {
	return snappy.NewBufferedWriter(w), nil
}

type nopCompressor struct{}

// NoCompression is a Compressor that simply does nothing
func NoCompression() Compressor {
	return &nopCompressor{}
}

func (c *nopCompressor) Reader(r io.Reader) (io.ReadCloser, error) {
	return util.NopReadCloser(r), nil
}

func (c *nopCompressor) Writer(w io.Writer) (io.WriteCloser, error) {
	return util.NopWriteCloser(w), nil
}
