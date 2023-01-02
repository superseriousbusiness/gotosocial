package storage

import (
	"bytes"
	"io"
	"sync"

	"codeberg.org/gruf/go-iotools"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/snappy"
	"github.com/klauspost/compress/zlib"
)

// Compressor defines a means of compressing/decompressing values going into a key-value store
type Compressor interface {
	// Reader returns a new decompressing io.ReadCloser based on supplied (compressed) io.Reader
	Reader(io.ReadCloser) (io.ReadCloser, error)

	// Writer returns a new compressing io.WriteCloser based on supplied (uncompressed) io.Writer
	Writer(io.WriteCloser) (io.WriteCloser, error)
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
	_, _ = gw.Write([]byte{})
	_ = gw.Close()

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

func (c *gzipCompressor) Reader(rc io.ReadCloser) (io.ReadCloser, error) {
	var released bool

	// Acquire from pool.
	gr := c.rpool.Get().(*gzip.Reader)
	if err := gr.Reset(rc); err != nil {
		c.rpool.Put(gr)
		return nil, err
	}

	return iotools.ReadCloser(gr, iotools.CloserFunc(func() error {
		if !released {
			released = true
			defer c.rpool.Put(gr)
		}

		// Close compressor
		err1 := gr.Close()

		// Close original stream.
		err2 := rc.Close()

		// Return err1 or 2
		if err1 != nil {
			return err1
		}
		return err2
	})), nil
}

func (c *gzipCompressor) Writer(wc io.WriteCloser) (io.WriteCloser, error) {
	var released bool

	// Acquire from pool.
	gw := c.wpool.Get().(*gzip.Writer)
	gw.Reset(wc)

	return iotools.WriteCloser(gw, iotools.CloserFunc(func() error {
		if !released {
			released = true
			c.wpool.Put(gw)
		}

		// Close compressor
		err1 := gw.Close()

		// Close original stream.
		err2 := wc.Close()

		// Return err1 or 2
		if err1 != nil {
			return err1
		}
		return err2
	})), nil
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

func (c *zlibCompressor) Reader(rc io.ReadCloser) (io.ReadCloser, error) {
	var released bool
	zr := c.rpool.Get().(interface {
		io.ReadCloser
		zlib.Resetter
	})
	if err := zr.Reset(rc, c.dict); err != nil {
		c.rpool.Put(zr)
		return nil, err
	}
	return iotools.ReadCloser(zr, iotools.CloserFunc(func() error {
		if !released {
			released = true
			defer c.rpool.Put(zr)
		}

		// Close compressor
		err1 := zr.Close()

		// Close original stream.
		err2 := rc.Close()

		// Return err1 or 2
		if err1 != nil {
			return err1
		}
		return err2
	})), nil
}

func (c *zlibCompressor) Writer(wc io.WriteCloser) (io.WriteCloser, error) {
	var released bool

	// Acquire from pool.
	zw := c.wpool.Get().(*zlib.Writer)
	zw.Reset(wc)

	return iotools.WriteCloser(zw, iotools.CloserFunc(func() error {
		if !released {
			released = true
			c.wpool.Put(zw)
		}

		// Close compressor
		err1 := zw.Close()

		// Close original stream.
		err2 := wc.Close()

		// Return err1 or 2
		if err1 != nil {
			return err1
		}
		return err2
	})), nil
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

func (c *snappyCompressor) Reader(rc io.ReadCloser) (io.ReadCloser, error) {
	var released bool

	// Acquire from pool.
	sr := c.rpool.Get().(*snappy.Reader)
	sr.Reset(rc)

	return iotools.ReadCloser(sr, iotools.CloserFunc(func() error {
		if !released {
			released = true
			defer c.rpool.Put(sr)
		}

		// Close original stream.
		return rc.Close()
	})), nil
}

func (c *snappyCompressor) Writer(wc io.WriteCloser) (io.WriteCloser, error) {
	var released bool

	// Acquire from pool.
	sw := c.wpool.Get().(*snappy.Writer)
	sw.Reset(wc)

	return iotools.WriteCloser(sw, iotools.CloserFunc(func() error {
		if !released {
			released = true
			c.wpool.Put(sw)
		}

		// Close original stream.
		return wc.Close()
	})), nil
}

type nopCompressor struct{}

// NoCompression is a Compressor that simply does nothing.
func NoCompression() Compressor {
	return &nopCompressor{}
}

func (c *nopCompressor) Reader(rc io.ReadCloser) (io.ReadCloser, error) {
	return rc, nil
}

func (c *nopCompressor) Writer(wc io.WriteCloser) (io.WriteCloser, error) {
	return wc, nil
}
