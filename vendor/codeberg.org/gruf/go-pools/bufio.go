package pools

import (
	"bufio"
	"io"
	"sync"
)

// BufioReaderPool is a pooled allocator for bufio.Reader objects
type BufioReaderPool interface {
	// Get fetches a bufio.Reader from pool and resets to supplied reader
	Get(io.Reader) *bufio.Reader

	// Put places supplied bufio.Reader back in pool
	Put(*bufio.Reader)
}

// NewBufioReaderPool returns a newly instantiated bufio.Reader pool
func NewBufioReaderPool(size int) BufioReaderPool {
	return &bufioReaderPool{
		Pool: sync.Pool{
			New: func() interface{} {
				return bufio.NewReaderSize(nil, size)
			},
		},
	}
}

// bufioReaderPool is our implementation of BufioReaderPool
type bufioReaderPool struct{ sync.Pool }

func (p *bufioReaderPool) Get(r io.Reader) *bufio.Reader {
	br := p.Pool.Get().(*bufio.Reader)
	br.Reset(r)
	return br
}

func (p *bufioReaderPool) Put(br *bufio.Reader) {
	br.Reset(nil)
	p.Pool.Put(br)
}

// BufioWriterPool is a pooled allocator for bufio.Writer objects
type BufioWriterPool interface {
	// Get fetches a bufio.Writer from pool and resets to supplied writer
	Get(io.Writer) *bufio.Writer

	// Put places supplied bufio.Writer back in pool
	Put(*bufio.Writer)
}

// NewBufioWriterPool returns a newly instantiated bufio.Writer pool
func NewBufioWriterPool(size int) BufioWriterPool {
	return &bufioWriterPool{
		Pool: sync.Pool{
			New: func() interface{} {
				return bufio.NewWriterSize(nil, size)
			},
		},
	}
}

// bufioWriterPool is our implementation of BufioWriterPool
type bufioWriterPool struct{ sync.Pool }

func (p *bufioWriterPool) Get(w io.Writer) *bufio.Writer {
	bw := p.Pool.Get().(*bufio.Writer)
	bw.Reset(w)
	return bw
}

func (p *bufioWriterPool) Put(bw *bufio.Writer) {
	bw.Reset(nil)
	p.Pool.Put(bw)
}
