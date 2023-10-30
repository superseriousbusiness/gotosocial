package mutexes

import (
	"sync"
	"sync/atomic"
)

var (
	// possible lock types.
	lockTypeRead  = uint8(1) << 0
	lockTypeWrite = uint8(1) << 1
	lockTypeMap   = uint8(1) << 2

	// possible mutexmap states.
	stateUnlockd = uint8(0)
	stateRLocked = uint8(1)
	stateLocked  = uint8(2)
	stateInUse   = uint8(3)
)

// acquireState attempts to acquire required map state for lockType,
// returns the updated state and whether the lock type was acquired.
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

// rwmutexState is a very simple *representation* of a read-write
// mutex, though not one in implementation. it works by
// tracking the lock state for a given map key, which is
// protected by the map's mutex.
type rwmutexState struct {
	rcnt int32  // read lock count
	lock uint8  // lock type
	key  string // map key
}

func (mu *rwmutexState) CanLock(lt uint8) bool {
	return mu.lock == 0 ||
		(mu.lock&lockTypeRead != 0 && lt&lockTypeRead != 0)
}

func (mu *rwmutexState) Lock(lt uint8) {
	// Set lock type
	mu.lock = lt

	if lt&lockTypeRead != 0 {
		// RLock, increment
		mu.rcnt++
	}
}

func (mu *rwmutexState) Unlock() {
	if mu.rcnt > 0 {
		// RUnlock
		mu.rcnt--
	}

	if mu.rcnt == 0 {
		// Total unlock
		mu.lock = 0
	}
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

	var mu *rwmutexState
	var ok bool

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

		// Alloc mu from pool
		mu = st.mmap.mpool.Acquire()
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
		st.mmap.cleanup()
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
	st.mmap.cleanup()
}
