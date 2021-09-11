package bufpool

import (
	"sync"

	"git.iim.gay/grufwub/go-bytes"
)

// MAX returns the maximum possible sized slice that can be stored in a BufferPool
func MAX() int {
	return log2Max
}

// BufferPool is a variable sized buffer pool, separated into memory pages increasing
// by powers of 2. This can offer large improvements over a sync.Pool designed to allocate
// buffers of single sizes, or multiple buffer pools of differing allocation sizes
type BufferPool struct {
	noCopy noCopy //nolint

	// pools is a predefined-length array of sync.Pools, handling
	// ranges in capacity of 2**(n) --> 2**(n+1)
	pools [log2MaxPower + 1]sync.Pool
	once  sync.Once
}

// init simply sets the allocator funcs for each of the pools
func (p *BufferPool) init() {
	for i := range p.pools {
		p.pools[i].New = func() interface{} {
			return &bytes.Buffer{}
		}
	}
}

// Get retrieves a Buffer of at least supplied capacity from the pool,
// allocating only if strictly necessary. If a capacity above the maximum
// supported (see .MAX()) is requested, a slice is allocated with
// expectance that it will just be dropped on call to .Put()
func (p *BufferPool) Get(cap int) *bytes.Buffer {
	// If cap out of bounds, just alloc
	if cap < 2 || cap > log2Max {
		buf := bytes.NewBuffer(make([]byte, 0, cap))
		return &buf
	}

	// Ensure initialized
	p.once.Do(p.init)

	// Calculate page idx from log2 table
	pow := uint8(log2Table[cap])
	pool := &p.pools[pow-1]

	// Attempt to fetch buf from pool
	buf := pool.Get().(*bytes.Buffer)

	// Check of required capacity
	if buf.Cap() < cap {
		// We allocate via this method instead
		// of by buf.Guarantee() as this way we
		// can allocate only what the user requested.
		//
		// buf.Guarantee() can allocate alot more...
		buf.B = make([]byte, 0, cap)
	}

	return buf
}

// Put resets and place the supplied Buffer back in its appropriate pool. Buffers
// Buffers below or above maximum supported capacity (see .MAX()) will be dropped
func (p *BufferPool) Put(buf *bytes.Buffer) {
	// Drop out of size range buffers
	if buf.Cap() < 2 || buf.Cap() > log2Max {
		return
	}

	// Ensure initialized
	p.once.Do(p.init)

	// Calculate page idx from log2 table
	pow := uint8(log2Table[buf.Cap()])
	pool := &p.pools[pow-1]

	// Reset, place in pool
	buf.Reset()
	pool.Put(buf)
}

//nolint
type noCopy struct{}

//nolint
func (n *noCopy) Lock() {}

//nolint
func (n *noCopy) Unlock() {}
