package mutexes

import (
	"runtime"
	"sync"
)

const (
	// default values.
	defaultWake = 1024
)

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
	queue *sync.WaitGroup
	qucnt int32

	mumap map[string]*rwmutexState
	mpool rwmutexPool
	evict []*rwmutexState

	count int32
	maxmu int32
	wake  int32

	mapmu sync.Mutex
	state uint8

	// NOTE:
	// The 2 main thread synchronization mechanisms in use here are the sync.Mutex
	// and the sync.WaitGroup. We take advantage of these instead of using simpler
	// CAS-spin-locks as we gain the implemented goroutine parking measures on "spin".
	// This is important as when when a goroutine blocks on acquiring a mutex lock,
	// it may get parked and allow another scheduled goroutine to be released that
	// actually releases the map from map / write locked state.
}

// NewMap returns a new MutexMap instance with provided max no. open mutexes.
func NewMap(max, wake int32) *MutexMap {
	var mm MutexMap
	mm.Init(max, wake)
	return &mm
}

// Init initializes the MutexMap with given max open locks and wake modulus.
func (mm *MutexMap) Init(max, wake int32) {
	mm.mapmu.Lock()
	if mm.mumap == nil {
		mm.queue = new(sync.WaitGroup)
		mm.mumap = make(map[string]*rwmutexState)
	}
	mm.set(max, wake)
	mm.mapmu.Unlock()
}

// SET sets the MutexMap max open locks and wake modulus, returns current values.
// For values less than zero defaults are set, and zero is non-op.
func (mm *MutexMap) SET(max, wake int32) (int32, int32) {
	mm.mapmu.Lock()
	max, wake = mm.set(max, wake)
	mm.mapmu.Unlock()
	return max, wake
}

// set contains the actual logic for MutexMap.SET, without mutex protection.
func (mm *MutexMap) set(max, wake int32) (int32, int32) {
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
		mm.maxmu = mm.wake * int32(procs)

	// Set supplied max
	case max > 0:
		mm.maxmu = max
	}

	// Fetch current values
	return mm.maxmu, mm.wake
}

// spinLock will wait (using a mutex to sleep thread) until conditional returns true.
func (mm *MutexMap) spinLock(cond func() bool) {
	for {
		// Acquire map lock
		mm.mapmu.Lock()

		// Perform check
		if cond() {
			return
		}

		// Current queue ptr
		queue := mm.queue

		// Queue ourselves
		queue.Add(1)
		mm.qucnt++

		// Unlock map
		mm.mapmu.Unlock()

		// Wait on notify
		mm.queue.Wait()
	}
}

// lock will acquire a lock of given type on the 'mutex' at key.
func (mm *MutexMap) lock(key string, lt uint8) func() {
	var mu *rwmutexState
	var ok bool

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
		// No mutex for key, alloc new
		mu = mm.mpool.Acquire()
		mm.mumap[key] = mu

		// Set our key
		mu.key = key

		// Queue mutex for eviction
		mm.evict = append(mm.evict, mu)
	}

	// Lock mutex
	mu.Lock(lt)

	// Unlock map
	mm.mapmu.Unlock()

	return func() {
		mm.mapmu.Lock()
		mu.Unlock()
		mm.cleanup()
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

	// Calculate current wake modulus
	wakemod := mm.count % mm.wake

	if mm.count != 0 && wakemod != 0 {
		// Fast path => no cleanup.
		// Unlock, return early
		mm.mapmu.Unlock()
		return
	}

	// Launch cleanup goroutine via
	// outlined function below, this
	// allows inlining of clean().
	mm.goCleanup(wakemod)
}

// goCleanup launches the slow part of the cleanup routine in a separate
// goroutine, using given calculated wake modulus. This releases all queued
// goroutines and evicts any mutexes currently awaiting.
func (mm *MutexMap) goCleanup(wakemod int32) {
	go func() {
		if wakemod == 0 {
			// Release queued goroutines
			mm.queue.Add(-int(mm.qucnt))

			// Allocate new queue and reset
			mm.queue = new(sync.WaitGroup)
			mm.qucnt = 0
		}

		if mm.count == 0 {
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
		}

		// Unlock map
		mm.mapmu.Unlock()
	}()
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
