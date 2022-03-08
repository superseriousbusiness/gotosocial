package mutexes

import (
	"runtime"
	"sync"
	"sync/atomic"
)

const (
	// possible lock types.
	lockTypeRead  = uint8(1) << 0
	lockTypeWrite = uint8(1) << 1
	lockTypeMap   = uint8(1) << 2

	// possible mutexmap states.
	stateUnlockd = uint8(0)
	stateRLocked = uint8(1)
	stateLocked  = uint8(2)
	stateInUse   = uint8(3)

	// default values.
	defaultWake = 1024
)

// acquireState attempts to acquire required map state for lockType.
func acquireState(state uint8, lt uint8) (uint8, bool) {
	switch state {
	// Unlocked state
	// (all allowed)
	case stateUnlockd:

	// Keys locked, no state lock.
	// (don't allow map locks)
	case stateInUse:
		if lt&lockTypeMap != 0 {
			return 0, false
		}

	// Read locked
	// (only allow read locks)
	case stateRLocked:
		if lt&lockTypeRead == 0 {
			return 0, false
		}

	// Write locked
	// (none allowed)
	case stateLocked:
		return 0, false

	// shouldn't reach here
	default:
		panic("unexpected state")
	}

	switch {
	// If unlocked and not a map
	// lock request, set in use
	case lt&lockTypeMap == 0:
		if state == stateUnlockd {
			state = stateInUse
		}

	// Set read lock state
	case lt&lockTypeRead != 0:
		state = stateRLocked

	// Set write lock state
	case lt&lockTypeWrite != 0:
		state = stateLocked

	default:
		panic("unexpected lock type")
	}

	return state, true
}

// MutexMap is a structure that allows read / write locking key, performing
// as you'd expect a map[string]*sync.RWMutex to perform. The differences
// being that the entire map can itself be read / write locked, it uses memory
// pooling for the mutex (not quite) structures, and it is self-evicting. The
// core configurations of maximum no. open locks and wake modulus* are user
// definable.
//
// * The wake modulus is the number that the current number of open locks is
// modulused against to determine how often to notify sleeping goroutines.
// These are goroutines that are attempting to lock a key / whole map and are
// awaiting a permissible state (.e.g no key write locks allowed when the
// map is read locked).
type MutexMap struct {
	qpool pool
	queue []*sync.Mutex

	mumap map[string]*rwmutex
	mpool pool
	evict []*rwmutex

	count int32
	maxmu int32
	wake  int32

	mapmu sync.Mutex
	state uint8
}

// NewMap returns a new MutexMap instance with provided max no. open mutexes.
func NewMap(max, wake int32) MutexMap {
	// Determine wake mod.
	if wake < 1 {
		wake = defaultWake
	}

	// Determine max no. mutexes
	if max < 1 {
		procs := runtime.GOMAXPROCS(0)
		max = wake * int32(procs)
	}

	return MutexMap{
		qpool: pool{
			alloc: func() interface{} {
				return &sync.Mutex{}
			},
		},
		mumap: make(map[string]*rwmutex, max),
		mpool: pool{
			alloc: func() interface{} {
				return &rwmutex{}
			},
		},
		maxmu: max,
		wake:  wake,
	}
}

// MAX sets the MutexMap max open locks and wake modulus, returns current values.
// For values less than zero defaults are set, and zero is non-op.
func (mm *MutexMap) SET(max, wake int32) (int32, int32) {
	mm.mapmu.Lock()

	switch {
	// Set default wake
	case wake < 0:
		mm.wake = defaultWake

	// Set supplied wake
	case wake > 0:
		mm.wake = wake
	}

	switch {
	// Set default max
	case max < 0:
		procs := runtime.GOMAXPROCS(0)
		mm.maxmu = wake * int32(procs)

	// Set supplied max
	case max > 0:
		mm.maxmu = max
	}

	// Fetch values
	max = mm.maxmu
	wake = mm.wake

	mm.mapmu.Unlock()
	return max, wake
}

// spinLock will wait (using a mutex to sleep thread) until conditional returns true.
func (mm *MutexMap) spinLock(cond func() bool) {
	var mu *sync.Mutex

	for {
		// Acquire map lock
		mm.mapmu.Lock()

		if cond() {
			// Release mu if needed
			if mu != nil {
				mm.qpool.Release(mu)
			}
			return
		}

		// Alloc mu if needed
		if mu == nil {
			v := mm.qpool.Acquire()
			mu = v.(*sync.Mutex)
		}

		// Queue ourselves
		mm.queue = append(mm.queue, mu)
		mu.Lock()

		// Unlock map
		mm.mapmu.Unlock()

		// Wait on notify
		mu.Lock()
		mu.Unlock()
	}
}

// lock will acquire a lock of given type on the 'mutex' at key.
func (mm *MutexMap) lock(key string, lt uint8) func() {
	var ok bool
	var mu *rwmutex

	// Spin lock until returns true
	mm.spinLock(func() bool {
		// Check not overloaded
		if !(mm.count < mm.maxmu) {
			return false
		}

		// Attempt to acquire usable map state
		state, ok := acquireState(mm.state, lt)
		if !ok {
			return false
		}

		// Update state
		mm.state = state

		// Ensure mutex at key
		// is in lockable state
		mu, ok = mm.mumap[key]
		return !ok || mu.CanLock(lt)
	})

	// Incr count
	mm.count++

	if !ok {
		// No mutex found for key

		// Alloc from pool
		v := mm.mpool.Acquire()
		mu = v.(*rwmutex)
		mm.mumap[key] = mu

		// Set our key
		mu.key = key

		// Queue for eviction
		mm.evict = append(mm.evict, mu)
	}

	// Lock mutex
	mu.Lock(lt)

	// Unlock map
	mm.mapmu.Unlock()

	return func() {
		mm.mapmu.Lock()
		mu.Unlock()
		go mm.cleanup()
	}
}

// lockMap will lock the whole map under given lock type.
func (mm *MutexMap) lockMap(lt uint8) {
	// Spin lock until returns true
	mm.spinLock(func() bool {
		// Attempt to acquire usable map state
		state, ok := acquireState(mm.state, lt)
		if !ok {
			return false
		}

		// Update state
		mm.state = state

		return true
	})

	// Incr count
	mm.count++

	// State acquired, unlock
	mm.mapmu.Unlock()
}

// cleanup is performed as the final stage of unlocking a locked key / map state, finally unlocks map.
func (mm *MutexMap) cleanup() {
	// Decr count
	mm.count--

	if mm.count%mm.wake == 0 {
		// Notify queued routines
		for _, mu := range mm.queue {
			mu.Unlock()
		}

		// Reset queue
		mm.queue = mm.queue[:0]
	}

	if mm.count < 1 {
		// Perform evictions
		for _, mu := range mm.evict {
			key := mu.key
			mu.key = ""
			delete(mm.mumap, key)
			mm.mpool.Release(mu)
		}

		// Reset map state
		mm.evict = mm.evict[:0]
		mm.state = stateUnlockd
		mm.mpool.GC()
		mm.qpool.GC()
	}

	// Unlock map
	mm.mapmu.Unlock()
}

// RLockMap acquires a read lock over the entire map, returning a lock state for acquiring key read locks.
// Please note that the 'unlock()' function will block until all keys locked from this state are unlocked.
func (mm *MutexMap) RLockMap() *LockState {
	mm.lockMap(lockTypeRead | lockTypeMap)
	return &LockState{
		mmap: mm,
		ltyp: lockTypeRead,
	}
}

// LockMap acquires a write lock over the entire map, returning a lock state for acquiring key read/write locks.
// Please note that the 'unlock()' function will block until all keys locked from this state are unlocked.
func (mm *MutexMap) LockMap() *LockState {
	mm.lockMap(lockTypeWrite | lockTypeMap)
	return &LockState{
		mmap: mm,
		ltyp: lockTypeWrite,
	}
}

// RLock acquires a mutex read lock for supplied key, returning an RUnlock function.
func (mm *MutexMap) RLock(key string) (runlock func()) {
	return mm.lock(key, lockTypeRead)
}

// Lock acquires a mutex write lock for supplied key, returning an Unlock function.
func (mm *MutexMap) Lock(key string) (unlock func()) {
	return mm.lock(key, lockTypeWrite)
}

// LockState represents a window to a locked MutexMap.
type LockState struct {
	wait sync.WaitGroup
	mmap *MutexMap
	done uint32
	ltyp uint8
}

// Lock: see MutexMap.Lock() definition. Will panic if map only read locked.
func (st *LockState) Lock(key string) (unlock func()) {
	return st.lock(key, lockTypeWrite)
}

// RLock: see MutexMap.RLock() definition.
func (st *LockState) RLock(key string) (runlock func()) {
	return st.lock(key, lockTypeRead)
}

// lock: see MutexMap.lock() definition.
func (st *LockState) lock(key string, lt uint8) func() {
	st.wait.Add(1) // track lock

	if atomic.LoadUint32(&st.done) == 1 {
		panic("called (r)lock on unlocked state")
	} else if lt&lockTypeWrite != 0 &&
		st.ltyp&lockTypeWrite == 0 {
		panic("called lock on rlocked map")
	}

	var ok bool
	var mu *rwmutex

	// Spin lock until returns true
	st.mmap.spinLock(func() bool {
		// Check not overloaded
		if !(st.mmap.count < st.mmap.maxmu) {
			return false
		}

		// Ensure mutex at key
		// is in lockable state
		mu, ok = st.mmap.mumap[key]
		return !ok || mu.CanLock(lt)
	})

	// Incr count
	st.mmap.count++

	if !ok {
		// No mutex found for key

		// Alloc from pool
		v := st.mmap.mpool.Acquire()
		mu = v.(*rwmutex)
		st.mmap.mumap[key] = mu

		// Set our key
		mu.key = key

		// Queue for eviction
		st.mmap.evict = append(st.mmap.evict, mu)
	}

	// Lock mutex
	mu.Lock(lt)

	// Unlock map
	st.mmap.mapmu.Unlock()

	return func() {
		st.mmap.mapmu.Lock()
		mu.Unlock()
		go st.mmap.cleanup()
		st.wait.Add(-1)
	}
}

// UnlockMap will close this state and release the currently locked map.
func (st *LockState) UnlockMap() {
	if !atomic.CompareAndSwapUint32(&st.done, 0, 1) {
		panic("called unlockmap on expired state")
	}
	st.wait.Wait()
	st.mmap.mapmu.Lock()
	go st.mmap.cleanup()
}

// rwmutex is a very simple *representation* of a read-write
// mutex, though not one in implementation. it works by
// tracking the lock state for a given map key, which is
// protected by the map's mutex.
type rwmutex struct {
	rcnt uint32
	lock uint8
	key  string
}

func (mu *rwmutex) CanLock(lt uint8) bool {
	return mu.lock == 0 ||
		(mu.lock&lockTypeRead != 0 && lt&lockTypeRead != 0)
}

func (mu *rwmutex) Lock(lt uint8) {
	mu.lock = lt
	if lt&lockTypeRead != 0 {
		mu.rcnt++
	}
}

func (mu *rwmutex) Unlock() {
	mu.rcnt--
	if mu.rcnt == 0 {
		mu.lock = 0
	}
}
