package kv

import (
	"errors"
	"io"

	"codeberg.org/gruf/go-mutexes"
)

var ErrStateClosed = errors.New("store/kv: state closed")

// StateRO provides a read-only window to the store. While this
// state is active during the Read() function window, the entire
// store will be read-locked. The state is thread-safe for concurrent
// use UNTIL the moment that your supplied function to Read() returns,
// then the state has zero guarantees
type StateRO struct {
	store *KVStore
	state *mutexes.LockState
}

func (st *StateRO) Get(key string) ([]byte, error) {
	// Check not closed
	if st.store == nil {
		return nil, ErrStateClosed
	}

	// Pass request to store
	return st.store.get(st.state.RLock, key)
}

func (st *StateRO) GetStream(key string) (io.ReadCloser, error) {
	// Check not closed
	if st.store == nil {
		return nil, ErrStateClosed
	}

	// Pass request to store
	return st.store.getStream(st.state.RLock, key)
}

func (st *StateRO) Has(key string) (bool, error) {
	// Check not closed
	if st.store == nil {
		return false, ErrStateClosed
	}

	// Pass request to store
	return st.store.has(st.state.RLock, key)
}

func (st *StateRO) Release() {
	st.state.UnlockMap()
	st.store = nil
}

// StateRW provides a read-write window to the store. While this
// state is active during the Update() function window, the entire
// store will be locked. The state is thread-safe for concurrent
// use UNTIL the moment that your supplied function to Update() returns,
// then the state has zero guarantees
type StateRW struct {
	store *KVStore
	state *mutexes.LockState
}

func (st *StateRW) Get(key string) ([]byte, error) {
	// Check not closed
	if st.store == nil {
		return nil, ErrStateClosed
	}

	// Pass request to store
	return st.store.get(st.state.RLock, key)
}

func (st *StateRW) GetStream(key string) (io.ReadCloser, error) {
	// Check not closed
	if st.store == nil {
		return nil, ErrStateClosed
	}

	// Pass request to store
	return st.store.getStream(st.state.RLock, key)
}

func (st *StateRW) Put(key string, value []byte) error {
	// Check not closed
	if st.store == nil {
		return ErrStateClosed
	}

	// Pass request to store
	return st.store.put(st.state.Lock, key, value)
}

func (st *StateRW) PutStream(key string, r io.Reader) error {
	// Check not closed
	if st.store == nil {
		return ErrStateClosed
	}

	// Pass request to store
	return st.store.putStream(st.state.Lock, key, r)
}

func (st *StateRW) Has(key string) (bool, error) {
	// Check not closed
	if st.store == nil {
		return false, ErrStateClosed
	}

	// Pass request to store
	return st.store.has(st.state.RLock, key)
}

func (st *StateRW) Delete(key string) error {
	// Check not closed
	if st.store == nil {
		return ErrStateClosed
	}

	// Pass request to store
	return st.store.delete(st.state.Lock, key)
}

func (st *StateRW) Release() {
	st.state.UnlockMap()
	st.store = nil
}
