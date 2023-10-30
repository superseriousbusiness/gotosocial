package mutexes

import (
	"sync"
	"unsafe"
)

const (
	// possible lock types.
	lockTypeRead  = uint8(1) << 0
	lockTypeWrite = uint8(1) << 1
	lockTypeMap   = uint8(1) << 2

	// frequency of GC cycles
	// per no. unlocks. i.e.
	// every 'gcfreq' unlocks.
	gcfreq = 1024
)

// MutexMap is a structure that allows read / write locking
// per key, performing as you'd expect a map[string]*RWMutex
// to perform, without you needing to worry about deadlocks
// between competing read / write locks and the map's own mutex.
// It uses memory pooling for the internal "mutex" (ish) types
// and performs self-eviction of keys.
//
// Under the hood this is achieved using a single mutex for the
// map, state tracking for individual keys, and some simple waitgroup
// type structures to park / block goroutines waiting for keys.
type MutexMap struct {
	mapmu  sync.Mutex
	mumap  map[string]*rwmutexish
	mupool rwmutexPool
	count  uint32
}

// checkInit ensures MutexMap is initialized (UNSAFE).
func (mm *MutexMap) checkInit() {
	if mm.mumap == nil {
		mm.mumap = make(map[string]*rwmutexish)
	}
}

// Lock acquires a write lock on key in map, returning unlock function.
func (mm *MutexMap) Lock(key string) func() {
	return mm.lock(key, lockTypeWrite)
}

// RLock acquires a read lock on key in map, returning runlock function.
func (mm *MutexMap) RLock(key string) func() {
	return mm.lock(key, lockTypeRead)
}

func (mm *MutexMap) lock(key string, lt uint8) func() {
	// Perform first map lock
	// and check initialization
	// OUTSIDE the main loop.
	mm.mapmu.Lock()
	mm.checkInit()

	for {
		// Check map for mu.
		mu := mm.mumap[key]

		if mu == nil {
			// Allocate new mutex.
			mu = mm.mupool.Acquire()
			mm.mumap[key] = mu
		}

		if !mu.Lock(lt) {
			// Wait on mutex unlock, after
			// immediately relocking map mu.
			mu.WaitRelock(&mm.mapmu)
			continue
		}

		// Done with map.
		mm.mapmu.Unlock()

		// Return mutex unlock function.
		return func() { mm.unlock(key, mu) }
	}
}

func (mm *MutexMap) unlock(key string, mu *rwmutexish) {
	// Get map lock.
	mm.mapmu.Lock()

	// Unlock mutex.
	if mu.Unlock() {

		// Mutex fully unlocked
		// with zero waiters. Self
		// evict and release it.
		delete(mm.mumap, key)
		mm.mupool.Release(mu)
	}

	if mm.count++; mm.count%gcfreq == 0 {
		// Every 'gcfreq' unlocks perform
		// a garbage collection to keep
		// us squeaky clean :]
		mm.mupool.GC()
	}

	// Done with map.
	mm.mapmu.Unlock()
}

// rwmutexPool is a very simply memory rwmutexPool.
type rwmutexPool struct {
	current []*rwmutexish
	victim  []*rwmutexish
}

// Acquire will returns a rwmutexState from rwmutexPool (or alloc new).
func (p *rwmutexPool) Acquire() *rwmutexish {
	// First try the current queue
	if l := len(p.current) - 1; l >= 0 {
		mu := p.current[l]
		p.current = p.current[:l]
		return mu
	}

	// Next try the victim queue.
	if l := len(p.victim) - 1; l >= 0 {
		mu := p.victim[l]
		p.victim = p.victim[:l]
		return mu
	}

	// Lastly, alloc new.
	mu := new(rwmutexish)
	return mu
}

// Release places a sync.rwmutexState back in the rwmutexPool.
func (p *rwmutexPool) Release(mu *rwmutexish) {
	p.current = append(p.current, mu)
}

// GC will clear out unused entries from the rwmutexPool.
func (p *rwmutexPool) GC() {
	current := p.current
	p.current = nil
	p.victim = current
}

// rwmutexish is a RW mutex (ish), i.e. the representation
// of one only to be accessed within
type rwmutexish struct {
	tr trigger
	ln int32 // no. locks
	wn int32 // no. waiters
	lt uint8 // lock type
}

// Lock will lock the mutex for given lock type, in the
// sense that it will update the internal state tracker
// accordingly. Return value is true on successful lock.
func (mu *rwmutexish) Lock(lt uint8) bool {
	switch mu.lt {
	case lockTypeRead:
		// already read locked,
		// only permit more reads.
		if lt != lockTypeRead {
			return false
		}

	case lockTypeWrite:
		// already write locked,
		// no other locks allowed.
		return false

	default:
		// Fully unlocked.
		mu.lt = lt
	}

	// Update
	// count.
	mu.ln++

	return true
}

// Unlock will unlock the mutex, in the sense that
// it will update the internal state tracker accordingly.
// On any unlock it will awaken sleeping waiting threads.
// Returned boolean is if unlocked=true AND waiters=0.
func (mu *rwmutexish) Unlock() bool {
	var ok bool

	switch mu.ln--; {
	case mu.ln > 0 && mu.lt == lockTypeWrite:
		panic("BUG: multiple writer locks")
	case mu.ln < 0:
		panic("BUG: negative lock count")
	case mu.ln == 0:
		// Fully unlocked.
		mu.lt = 0

		// Only return true
		// with no waiters.
		ok = (mu.wn == 0)
	}

	// Awake all waiting
	// goroutines for mu.
	mu.tr.Trigger()
	return ok
}

// WaitRelock expects a mutex to be passed in already in
// the lock state. It incr the rwmutexish waiter count before
// unlocking the outer mutex and blocking on internal trigger.
// On awake it will relock outer mutex and decr wait count.
func (mu *rwmutexish) WaitRelock(outer *sync.Mutex) {
	mu.wn++
	outer.Unlock()
	mu.tr.Wait()
	outer.Lock()
	mu.wn--
}

// trigger uses the internals of sync.Cond to provide
// a waitgroup type structure (including goroutine parks)
// without such a heavy reliance on a delta value.
type trigger struct{ notifyList }

func (t *trigger) Trigger() {
	runtime_notifyListNotifyAll(&t.notifyList)
}

func (t *trigger) Wait() {
	v := runtime_notifyListAdd(&t.notifyList)
	runtime_notifyListWait(&t.notifyList, v)
}

// Approximation of notifyList in runtime/sema.go.
type notifyList struct {
	wait   uint32
	notify uint32
	lock   uintptr // key field of the mutex
	head   unsafe.Pointer
	tail   unsafe.Pointer
}

// See runtime/sema.go for documentation.
//
//go:linkname runtime_notifyListAdd sync.runtime_notifyListAdd
func runtime_notifyListAdd(l *notifyList) uint32

// See runtime/sema.go for documentation.
//
//go:linkname runtime_notifyListWait sync.runtime_notifyListWait
func runtime_notifyListWait(l *notifyList, t uint32)

// See runtime/sema.go for documentation.
//
//go:linkname runtime_notifyListNotifyAll sync.runtime_notifyListNotifyAll
func runtime_notifyListNotifyAll(l *notifyList)
