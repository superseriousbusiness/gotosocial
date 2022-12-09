package storage

import (
	"bytes"
	"io"
	"sync"

	"codeberg.org/gruf/go-store/v2/util"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/snappy"
	"github.com/klauspost/compress/zlib"
)

// Compressor defines a means of compressing/decompressing values going into a key-value store
type Compressor interface {
	// Reader returns a new decompressing io.ReadCloser based on supplied (compressed) io.Reader
	Reader(io.Reader) (io.ReadCloser, error)

	// Writer returns a new compressing io.WriteCloser based on supplied (uncompressed) io.Writer
	Writer(io.Writer) (io.WriteCloser, error)
}

type gzipCompressor struct {
	rpool sync.Pool
	wpool sync.Pool
}

// GZipCompressor returns a new Compressor that implements GZip at default compression level
func GZipCompressor() Compressor {
	return GZipCompressorLevel(gzip.DefaultCompression)
}

// GZipCompressorLevel returns a new Compressor that implements GZip at supplied compression level
func GZipCompressorLevel(level int) Compressor {
	// GZip readers immediately check for valid
	// header data on allocation / reset, so we
	// need a set of valid header data so we can
	// iniitialize reader instances in mempool.
	hdr := bytes.NewBuffer(nil)

	// Init writer to ensure valid level provided
	gw, err := gzip.NewWriterLevel(hdr, level)
	if err != nil {
		panic(err)
	}

	// Write empty data to ensure gzip
	// header data is in byte buffer.
	gw.Write([]byte{})
	gw.Close()

	return &gzipCompressor{
		rpool: sync.Pool{
			New: func() any {
				hdr := bytes.NewReader(hdr.Bytes())
				gr, _ := gzip.NewReader(hdr)
				return gr
			},
		},
		wpool: sync.Pool{
			New: func() any {
				gw, _ := gzip.NewWriterLevel(nil, level)
				return gw
			},
		},
	}
}

func (c *gzipCompressor) Reader(r io.Reader) (io.ReadCloser, error) {
	gr := c.rpool.Get().(*gzip.Reader)
	if err := gr.Reset(r); err != nil {
		c.rpool.Put(gr)
		return nil, err
	}
	return util.ReadCloserWithCallback(gr, func() {
		c.rpool.Put(gr)
	}), nil
}

func (c *gzipCompressor) Writer(w io.Writer) (io.WriteCloser, error) {
	gw := c.wpool.Get().(*gzip.Writer)
	gw.Reset(w)
	return util.WriteCloserWithCallback(gw, func() {
		c.wpool.Put(gw)
	}), nil
}

type zlibCompressor struct {
	rpool sync.Pool
	wpool sync.Pool
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
	// ZLib readers immediately check for valid
	// header data on allocation / reset, so we
	// need a set of valid header data so we can
	// iniitialize reader instances in mempool.
	hdr := bytes.NewBuffer(nil)

	// Init writer to ensure valid level + dict provided
	zw, err := zlib.NewWriterLevelDict(hdr, level, dict)
	if err != nil {
		panic(err)
	}

	// Write empty data to ensure zlib
	// header data is in byte buffer.
	zw.Write([]byte{})
	zw.Close()

	return &zlibCompressor{
		rpool: sync.Pool{
			New: func() any {
				hdr := bytes.NewReader(hdr.Bytes())
				zr, _ := zlib.NewReaderDict(hdr, dict)
				return zr
			},
		},
		wpool: sync.Pool{
			New: func() any {
				zw, _ := zlib.NewWriterLevelDict(nil, level, dict)
				return zw
			},
		},
		dict: dict,
	}
}

func (c *zlibCompressor) Reader(r io.Reader) (io.ReadCloser, error) {
	zr := c.rpool.Get().(interface {
		io.ReadCloser
		zlib.Resetter
	})
	if err := zr.Reset(r, c.dict); err != nil {
		c.rpool.Put(zr)
		return nil, err
	}
	return util.ReadCloserWithCallback(zr, func() {
		c.rpool.Put(zr)
	}), nil
}

func (c *zlibCompressor) Writer(w io.Writer) (io.WriteCloser, error) {
	zw := c.wpool.Get().(*zlib.Writer)
	zw.Reset(w)
	return util.WriteCloserWithCallback(zw, func() {
		c.wpool.Put(zw)
	}), nil
}

type snappyCompressor struct {
	rpool sync.Pool
	wpool sync.Pool
}

// SnappyCompressor returns a new Compressor that implements Snappy.
func SnappyCompressor() Compressor {
	return &snappyCompressor{
		rpool: sync.Pool{
			New: func() any { return snappy.NewReader(nil) },
		},
		wpool: sync.Pool{
			New: func() any { return snappy.NewWriter(nil) },
		},
	}
}

func (c *snappyCompressor) Reader(r io.Reader) (io.ReadCloser, error) {
	sr := c.rpool.Get().(*snappy.Reader)
	sr.Reset(r)
	return util.ReadCloserWithCallback(
		util.NopReadCloser(sr),
		func() { c.rpool.Put(sr) },
	), nil
}

func (c *snappyCompressor) Writer(w io.Writer) (io.WriteCloser, error) {
	sw := c.wpool.Get().(*snappy.Writer)
	sw.Reset(w)
	return util.WriteCloserWithCallback(
		util.NopWriteCloser(sw),
		func() { c.wpool.Put(sw) },
	), nil
}

type nopCompressor struct{}

// NoCompression is a Compressor that simply does nothing.
func NoCompression() Compressor {
	return &nopCompressor{}
}

func (c *nopCompressor) Reader(r io.Reader) (io.ReadCloser, error) {
	return util.NopReadCloser(r), nil
}

func (c *nopCompressor) Writer(w io.Writer) (io.WriteCloser, error) {
	return util.NopWriteCloser(w), nil
}
