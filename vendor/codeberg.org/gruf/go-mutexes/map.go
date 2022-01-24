package mutexes

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// locktype defines maskable mutexmap lock types.
type locktype uint8

const (
	// possible lock types.
	lockTypeRead  = locktype(1) << 0
	lockTypeWrite = locktype(1) << 1
	lockTypeMap   = locktype(1) << 2

	// possible mutexmap states.
	stateUnlockd = uint8(0)
	stateRLocked = uint8(1)
	stateLocked  = uint8(2)
	stateInUse   = uint8(3)
)

// permitLockType returns if provided locktype is permitted to go ahead in current state.
func permitLockType(state uint8, lt locktype) bool {
	switch state {
	// Unlocked state
	// (all allowed)
	case stateUnlockd:
		return true

	// Keys locked, no state lock.
	// (don't allow map locks)
	case stateInUse:
		return lt&lockTypeMap == 0

	// Read locked
	// (only allow read locks)
	case stateRLocked:
		return lt&lockTypeRead != 0

	// Write locked
	// (none allowed)
	case stateLocked:
		return false

	// shouldn't reach here
	default:
		panic("unexpected state")
	}
}

// MutexMap is a structure that allows having a map of self-evicting mutexes
// by key. You do not need to worry about managing the contents of the map,
// only requesting RLock/Lock for keys, and ensuring to call the returned
// unlock functions.
type MutexMap struct {
	mus   map[string]RWMutex
	mapMu sync.Mutex
	pool  sync.Pool
	queue []func()
	evict []func()
	count int32
	maxmu int32
	state uint8
}

// NewMap returns a new MutexMap instance with provided max no. open mutexes.
func NewMap(max int32) MutexMap {
	if max < 1 {
		// Default = 128 * GOMAXPROCS
		procs := runtime.GOMAXPROCS(0)
		max = int32(procs * 128)
	}
	return MutexMap{
		mus: make(map[string]RWMutex),
		pool: sync.Pool{
			New: func() interface{} {
				return NewRW()
			},
		},
		maxmu: max,
	}
}

// acquire will either acquire a mutex from pool or alloc.
func (mm *MutexMap) acquire() RWMutex {
	return mm.pool.Get().(RWMutex)
}

// release will release provided mutex to pool.
func (mm *MutexMap) release(mu RWMutex) {
	mm.pool.Put(mu)
}

// spinLock will wait (using a mutex to sleep thread) until 'cond()' returns true,
// returning with map lock. Note that 'cond' is performed within a map lock.
func (mm *MutexMap) spinLock(cond func() bool) {
	mu := mm.acquire()
	defer mm.release(mu)

	for {
		// Get map lock
		mm.mapMu.Lock()

		// Check if return
		if cond() {
			return
		}

		// Queue ourselves
		unlock := mu.Lock()
		mm.queue = append(mm.queue, unlock)
		mm.mapMu.Unlock()

		// Wait on notify
		mu.Lock()()
	}
}

// lockMutex will acquire a lock on the mutex at provided key, handling earlier allocated mutex if provided. Unlocks map on return.
func (mm *MutexMap) lockMutex(key string, lt locktype) func() {
	var unlock func()

	// Incr counter
	mm.count++

	// Check for existing mutex at key
	mu, ok := mm.mus[key]
	if !ok {
		// Alloc from pool
		mu = mm.acquire()
		mm.mus[key] = mu

		// Queue mutex for eviction
		mm.evict = append(mm.evict, func() {
			delete(mm.mus, key)
			mm.pool.Put(mu)
		})
	}

	// If no state, set in use.
	// State will already have been
	// set if this is from LockState{}
	if mm.state == stateUnlockd {
		mm.state = stateInUse
	}

	switch {
	// Read lock
	case lt&lockTypeRead != 0:
		unlock = mu.RLock()

	// Write lock
	case lt&lockTypeWrite != 0:
		unlock = mu.Lock()

	// shouldn't reach here
	default:
		panic("unexpected lock type")
	}

	// Unlock map + return
	mm.mapMu.Unlock()
	return func() {
		mm.mapMu.Lock()
		unlock()
		go mm.onUnlock()
	}
}

// onUnlock is performed as the final (async) stage of releasing an acquired key / map mutex.
func (mm *MutexMap) onUnlock() {
	// Decr counter
	mm.count--

	if mm.count < 1 {
		// Perform all queued evictions
		for i := 0; i < len(mm.evict); i++ {
			mm.evict[i]()
		}

		// Notify all waiting goroutines
		for i := 0; i < len(mm.queue); i++ {
			mm.queue[i]()
		}

		// Reset the map state
		mm.evict = nil
		mm.queue = nil
		mm.state = stateUnlockd
	}

	// Finally, unlock
	mm.mapMu.Unlock()
}

// RLockMap acquires a read lock over the entire map, returning a lock state for acquiring key read locks.
// Please note that the 'unlock()' function will block until all keys locked from this state are unlocked.
func (mm *MutexMap) RLockMap() *LockState {
	return mm.getMapLock(lockTypeRead)
}

// LockMap acquires a write lock over the entire map, returning a lock state for acquiring key read/write locks.
// Please note that the 'unlock()' function will block until all keys locked from this state are unlocked.
func (mm *MutexMap) LockMap() *LockState {
	return mm.getMapLock(lockTypeWrite)
}

// RLock acquires a mutex read lock for supplied key, returning an RUnlock function.
func (mm *MutexMap) RLock(key string) (runlock func()) {
	return mm.getLock(key, lockTypeRead)
}

// Lock acquires a mutex write lock for supplied key, returning an Unlock function.
func (mm *MutexMap) Lock(key string) (unlock func()) {
	return mm.getLock(key, lockTypeWrite)
}

// getLock will fetch lock of provided type, for given key, returning unlock function.
func (mm *MutexMap) getLock(key string, lt locktype) func() {
	// Spin until achieve lock
	mm.spinLock(func() bool {
		return permitLockType(mm.state, lt) &&
			mm.count < mm.maxmu // not overloaded
	})

	// Perform actual mutex lock
	return mm.lockMutex(key, lt)
}

// getMapLock will acquire a map lock of provided type, returning a LockState session.
func (mm *MutexMap) getMapLock(lt locktype) *LockState {
	// Spin until achieve lock
	mm.spinLock(func() bool {
		return permitLockType(mm.state, lt|lockTypeMap) &&
			mm.count < mm.maxmu // not overloaded
	})

	// Incr counter
	mm.count++

	switch {
	// Set read lock state
	case lt&lockTypeRead != 0:
		mm.state = stateRLocked

	// Set write lock state
	case lt&lockTypeWrite != 0:
		mm.state = stateLocked

	default:
		panic("unexpected lock type")
	}

	// Unlock + return
	mm.mapMu.Unlock()
	return &LockState{
		mmap: mm,
		ltyp: lt,
	}
}

// LockState represents a window to a locked MutexMap.
type LockState struct {
	wait sync.WaitGroup
	mmap *MutexMap
	done uint32
	ltyp locktype
}

// Lock: see MutexMap.Lock() definition. Will panic if map only read locked.
func (st *LockState) Lock(key string) (unlock func()) {
	return st.getLock(key, lockTypeWrite)
}

// RLock: see MutexMap.RLock() definition.
func (st *LockState) RLock(key string) (runlock func()) {
	return st.getLock(key, lockTypeRead)
}

// UnlockMap will close this state and release the currently locked map.
func (st *LockState) UnlockMap() {
	// Set state to finished (or panic if already done)
	if !atomic.CompareAndSwapUint32(&st.done, 0, 1) {
		panic("called UnlockMap() on expired state")
	}

	// Wait until done
	st.wait.Wait()

	// Async reset map
	st.mmap.mapMu.Lock()
	go st.mmap.onUnlock()
}

// getLock: see MutexMap.getLock() definition.
func (st *LockState) getLock(key string, lt locktype) func() {
	st.wait.Add(1) // track lock

	// Check if closed, or if write lock is allowed
	if atomic.LoadUint32(&st.done) == 1 {
		panic("map lock closed")
	} else if lt&lockTypeWrite != 0 &&
		st.ltyp&lockTypeWrite == 0 {
		panic("called .Lock() on rlocked map")
	}

	// Spin until achieve map lock
	st.mmap.spinLock(func() bool {
		return st.mmap.count < st.mmap.maxmu
	}) // i.e. not overloaded

	// Perform actual mutex lock
	unlock := st.mmap.lockMutex(key, lt)

	return func() {
		unlock()
		st.wait.Done()
	}
}
