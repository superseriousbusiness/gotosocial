package pools

import (
	"bufio"
	"io"
	"sync"
)

// BufioReaderPool is a pooled allocator for bufio.Reader objects.
type BufioReaderPool interface {
	// Get fetches a bufio.Reader from pool and resets to supplied reader
	Get(io.Reader) *bufio.Reader

	// Put places supplied bufio.Reader back in pool
	Put(*bufio.Reader)
}

// NewBufioReaderPool returns a newly instantiated bufio.Reader pool.
func NewBufioReaderPool(size int) BufioReaderPool {
	return &bufioReaderPool{
		pool: sync.Pool{
			New: func() interface{} {
				return bufio.NewReaderSize(nil, size)
			},
		},
		size: size,
	}
}

// bufioReaderPool is our implementation of BufioReaderPool.
type bufioReaderPool struct {
	pool sync.Pool
	size int
}

func (p *bufioReaderPool) Get(r io.Reader) *bufio.Reader {
	br := p.pool.Get().(*bufio.Reader)
	br.Reset(r)
	return br
}

func (p *bufioReaderPool) Put(br *bufio.Reader) {
	if br.Size() < p.size {
		return
	}
	br.Reset(nil)
	p.pool.Put(br)
}

// BufioWriterPool is a pooled allocator for bufio.Writer objects.
type BufioWriterPool interface {
	// Get fetches a bufio.Writer from pool and resets to supplied writer
	Get(io.Writer) *bufio.Writer

	// Put places supplied bufio.Writer back in pool
	Put(*bufio.Writer)
}

// NewBufioWriterPool returns a newly instantiated bufio.Writer pool.
func NewBufioWriterPool(size int) BufioWriterPool {
	return &bufioWriterPool{
		pool: sync.Pool{
			New: func() interface{} {
				return bufio.NewWriterSize(nil, size)
			},
		},
		size: size,
	}
}

// bufioWriterPool is our implementation of BufioWriterPool.
type bufioWriterPool struct {
	pool sync.Pool
	size int
}

func (p *bufioWriterPool) Get(w io.Writer) *bufio.Writer {
	bw := p.pool.Get().(*bufio.Writer)
	bw.Reset(w)
	return bw
}

func (p *bufioWriterPool) Put(bw *bufio.Writer) {
	if bw.Size() < p.size {
		return
	}
	bw.Reset(nil)
	p.pool.Put(bw)
}
