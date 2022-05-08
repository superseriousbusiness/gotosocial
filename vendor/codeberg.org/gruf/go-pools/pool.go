package pools

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

type Pool struct {
	// New is used to instantiate new items
	New func() interface{}

	// Evict is called on evicted items during pool .Clean()
	Evict func(interface{})

	local    unsafe.Pointer // ptr to []_ppool
	localSz  int64          // count of all elems in local
	victim   unsafe.Pointer // ptr to []_ppool
	victimSz int64          // count of all elems in victim
	mutex    sync.Mutex     // mutex protects new cleanups, and new allocations of local
}

// Get attempts to fetch an item from the pool, failing that allocates with supplied .New() function
func (p *Pool) Get() interface{} {
	// Get local pool for proc
	// (also pins proc)
	pool, pid := p.pin()

	if v := pool.getPrivate(); v != nil {
		// local _ppool private elem acquired
		runtime_procUnpin()
		atomic.AddInt64(&p.localSz, -1)
		return v
	}

	if v := pool.get(); v != nil {
		// local _ppool queue elem acquired
		runtime_procUnpin()
		atomic.AddInt64(&p.localSz, -1)
		return v
	}

	// Unpin before attempting slow
	runtime_procUnpin()
	if v := p.getSlow(pid); v != nil {
		// note size decrementing
		// is handled within p.getSlow()
		// as we don't know if it came
		// from the local or victim pools
		return v
	}

	// Alloc new
	return p.New()
}

// Put places supplied item in the proc local pool
func (p *Pool) Put(v interface{}) {
	// Don't store nil
	if v == nil {
		return
	}

	// Get proc local pool
	// (also pins proc)
	pool, _ := p.pin()

	// first try private, then queue
	if !pool.setPrivate(v) {
		pool.put(v)
	}
	runtime_procUnpin()

	// Increment local pool size
	atomic.AddInt64(&p.localSz, 1)
}

// Clean will drop the current victim pools, move the current local pools to its
// place and reset the local pools ptr in order to be regenerated
func (p *Pool) Clean() {
	p.mutex.Lock()

	// victim becomes local, local becomes nil
	localPtr := atomic.SwapPointer(&p.local, nil)
	victimPtr := atomic.SwapPointer(&p.victim, localPtr)
	localSz := atomic.SwapInt64(&p.localSz, 0)
	atomic.StoreInt64(&p.victimSz, localSz)

	var victim []ppool
	if victimPtr != nil {
		victim = *(*[]ppool)(victimPtr)
	}

	// drain each of the vict _ppool items
	for i := 0; i < len(victim); i++ {
		ppool := &victim[i]
		ppool.evict(p.Evict)
	}

	p.mutex.Unlock()
}

// LocalSize returns the total number of elements in all the proc-local pools
func (p *Pool) LocalSize() int64 {
	return atomic.LoadInt64(&p.localSz)
}

// VictimSize returns the total number of elements in all the victim (old proc-local) pools
func (p *Pool) VictimSize() int64 {
	return atomic.LoadInt64(&p.victimSz)
}

// getSlow is the slow path for fetching an element, attempting to steal from other proc's
// local pools, and failing that, from the aging-out victim pools. pid is still passed so
// not all procs start iterating from the same index
func (p *Pool) getSlow(pid int) interface{} {
	// get local pools
	local := p.localPools()

	// Try to steal from other proc locals
	for i := 0; i < len(local); i++ {
		pool := &local[(pid+i+1)%len(local)]
		if v := pool.get(); v != nil {
			atomic.AddInt64(&p.localSz, -1)
			return v
		}
	}

	// get victim pools
	victim := p.victimPools()

	// Attempt to steal from victim pools
	for i := 0; i < len(victim); i++ {
		pool := &victim[(pid+i+1)%len(victim)]
		if v := pool.get(); v != nil {
			atomic.AddInt64(&p.victimSz, -1)
			return v
		}
	}

	// Set victim pools to nil (none found)
	atomic.StorePointer(&p.victim, nil)

	return nil
}

// localPools safely loads slice of local _ppools
func (p *Pool) localPools() []ppool {
	local := atomic.LoadPointer(&p.local)
	if local == nil {
		return nil
	}
	return *(*[]ppool)(local)
}

// victimPools safely loads slice of victim _ppools
func (p *Pool) victimPools() []ppool {
	victim := atomic.LoadPointer(&p.victim)
	if victim == nil {
		return nil
	}
	return *(*[]ppool)(victim)
}

// pin will get fetch pin proc to PID, fetch proc-local _ppool and current PID we're pinned to
func (p *Pool) pin() (*ppool, int) {
	for {
		// get local pools
		local := p.localPools()

		if len(local) > 0 {
			// local already initialized

			// pin to current proc
			pid := runtime_procPin()

			// check for pid local pool
			if pid < len(local) {
				return &local[pid], pid
			}

			// unpin from proc
			runtime_procUnpin()
		} else {
			// local not yet initialized

			// Check functions are set
			if p.New == nil {
				panic("new func must not be nil")
			}
			if p.Evict == nil {
				panic("evict func must not be nil")
			}
		}

		// allocate local
		p.allocLocal()
	}
}

// allocLocal allocates a new local pool slice, with the old length passed to check
// if pool was previously nil, or whether a change in GOMAXPROCS occurred
func (p *Pool) allocLocal() {
	// get pool lock
	p.mutex.Lock()

	// Calculate new size to use
	size := runtime.GOMAXPROCS(0)

	local := p.localPools()
	if len(local) != size {
		// GOMAXPROCS changed, reallocate
		pools := make([]ppool, size)
		atomic.StorePointer(&p.local, unsafe.Pointer(&pools))

		// Evict old local elements
		for i := 0; i < len(local); i++ {
			pool := &local[i]
			pool.evict(p.Evict)
		}
	}

	// Unlock pool
	p.mutex.Unlock()
}

// _ppool is a proc local pool
type _ppool struct {
	// root is the root element of the _ppool queue,
	// and protects concurrent access to the queue
	root unsafe.Pointer

	// private is a proc private member accessible
	// only to the pid this _ppool is assigned to,
	// except during evict (hence the unsafe pointer)
	private unsafe.Pointer
}

// ppool wraps _ppool with pad.
type ppool struct {
	_ppool

	// Prevents false sharing on widespread platforms with
	// 128 mod (cache line size) = 0 .
	pad [128 - unsafe.Sizeof(_ppool{})%128]byte
}

// getPrivate gets the proc private member
func (pp *_ppool) getPrivate() interface{} {
	ptr := atomic.SwapPointer(&pp.private, nil)
	if ptr == nil {
		return nil
	}
	return *(*interface{})(ptr)
}

// setPrivate sets the proc private member (only if unset)
func (pp *_ppool) setPrivate(v interface{}) bool {
	return atomic.CompareAndSwapPointer(&pp.private, nil, unsafe.Pointer(&v))
}

// get fetches an element from the queue
func (pp *_ppool) get() interface{} {
	for {
		// Attempt to load root elem
		root := atomic.LoadPointer(&pp.root)
		if root == nil {
			return nil
		}

		// Attempt to consume root elem
		if root == inUsePtr ||
			!atomic.CompareAndSwapPointer(&pp.root, root, inUsePtr) {
			continue
		}

		// Root becomes next in chain
		e := (*elem)(root)
		v := e.value

		// Place new root back in the chain
		atomic.StorePointer(&pp.root, unsafe.Pointer(e.next))
		putElem(e)

		return v
	}
}

// put places an element in the queue
func (pp *_ppool) put(v interface{}) {
	// Prepare next elem
	e := getElem()
	e.value = v

	for {
		// Attempt to load root elem
		root := atomic.LoadPointer(&pp.root)
		if root == inUsePtr {
			continue
		}

		// Set the next elem value (might be nil)
		e.next = (*elem)(root)

		// Attempt to store this new value at root
		if atomic.CompareAndSwapPointer(&pp.root, root, unsafe.Pointer(e)) {
			break
		}
	}
}

// hook evicts all entries from pool, calling hook on each
func (pp *_ppool) evict(hook func(interface{})) {
	if v := pp.getPrivate(); v != nil {
		hook(v)
	}
	for {
		v := pp.get()
		if v == nil {
			break
		}
		hook(v)
	}
}

// inUsePtr is a ptr used to indicate _ppool is in use
var inUsePtr = unsafe.Pointer(&elem{
	next:  nil,
	value: "in_use",
})

// elem defines an element in the _ppool queue
type elem struct {
	next  *elem
	value interface{}
}

// elemPool is a simple pool of unused elements
var elemPool = struct {
	root unsafe.Pointer
}{}

// getElem fetches a new elem from pool, or creates new
func getElem() *elem {
	// Attempt to load root elem
	root := atomic.LoadPointer(&elemPool.root)
	if root == nil {
		return &elem{}
	}

	// Attempt to consume root elem
	if root == inUsePtr ||
		!atomic.CompareAndSwapPointer(&elemPool.root, root, inUsePtr) {
		return &elem{}
	}

	// Root becomes next in chain
	e := (*elem)(root)
	atomic.StorePointer(&elemPool.root, unsafe.Pointer(e.next))
	e.next = nil

	return e
}

// putElem will place element in the pool
func putElem(e *elem) {
	e.value = nil

	// Attempt to load root elem
	root := atomic.LoadPointer(&elemPool.root)
	if root == inUsePtr {
		return // drop
	}

	// Set the next elem value (might be nil)
	e.next = (*elem)(root)

	// Attempt to store this new value at root
	atomic.CompareAndSwapPointer(&elemPool.root, root, unsafe.Pointer(e))
}

//go:linkname runtime_procPin sync.runtime_procPin
func runtime_procPin() int

//go:linkname runtime_procUnpin sync.runtime_procUnpin
func runtime_procUnpin()
