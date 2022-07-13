package atomics

import "sync"

// State provides user-friendly means of performing atomic-like
// operations on a uint32 state, and allowing callbacks on successful
// state change. This is a bit of a misnomer being where it is, as it
// actually uses a mutex under-the-hood.
type State struct {
	mutex sync.Mutex
	state uint32
}

// Store will update State value safely within mutex lock.
func (st *State) Store(val uint32) {
	st.mutex.Lock()
	st.state = val
	st.mutex.Unlock()
}

// Load will get value of State safely within mutex lock.
func (st *State) Load() uint32 {
	st.mutex.Lock()
	state := st.state
	st.mutex.Unlock()
	return state
}

// WithLock performs fn within State mutex lock, useful if you want
// to just use State's mutex for locking instead of creating another.
func (st *State) WithLock(fn func()) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	fn()
}

// Update performs fn within State mutex lock, with the current state
// value provided as an argument, and return value used to update state.
func (st *State) Update(fn func(state uint32) uint32) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.state = fn(st.state)
}

// CAS performs a compare-and-swap on State, calling fn on success. Success value is also returned.
func (st *State) CAS(cmp, swp uint32, fn func()) (ok bool) {
	// Acquire lock
	st.mutex.Lock()
	defer st.mutex.Unlock()

	// Perform CAS operation, fn() on success
	if ok = (st.state == cmp); ok {
		st.state = swp
		fn()
	}

	return
}
