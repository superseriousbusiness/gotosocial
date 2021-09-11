package kv

import (
	"io"

	"git.iim.gay/grufwub/go-errors"
)

var ErrStateClosed = errors.Define("store/kv: state closed")

// StateRO provides a read-only window to the store. While this
// state is active during the Read() function window, the entire
// store will be read-locked. The state is thread-safe for concurrent
// use UNTIL the moment that your supplied function to Read() returns,
// then the state has zero guarantees
type StateRO struct {
	store *KVStore
}

func (st *StateRO) Get(key string) ([]byte, error) {
	// Check not closed
	if st.store == nil {
		return nil, ErrStateClosed
	}

	// Pass request to store
	return st.store.get(key)
}

func (st *StateRO) GetStream(key string) (io.ReadCloser, error) {
	// Check not closed
	if st.store == nil {
		return nil, ErrStateClosed
	}

	// Pass request to store
	return st.store.getStream(key)
}

func (st *StateRO) Has(key string) (bool, error) {
	// Check not closed
	if st.store == nil {
		return false, ErrStateClosed
	}

	// Pass request to store
	return st.store.has(key)
}

func (st *StateRO) close() {
	st.store = nil
}

// StateRW provides a read-write window to the store. While this
// state is active during the Update() function window, the entire
// store will be locked. The state is thread-safe for concurrent
// use UNTIL the moment that your supplied function to Update() returns,
// then the state has zero guarantees
type StateRW struct {
	store *KVStore
}

func (st *StateRW) Get(key string) ([]byte, error) {
	// Check not closed
	if st.store == nil {
		return nil, ErrStateClosed
	}

	// Pass request to store
	return st.store.get(key)
}

func (st *StateRW) GetStream(key string) (io.ReadCloser, error) {
	// Check not closed
	if st.store == nil {
		return nil, ErrStateClosed
	}

	// Pass request to store
	return st.store.getStream(key)
}

func (st *StateRW) Put(key string, value []byte) error {
	// Check not closed
	if st.store == nil {
		return ErrStateClosed
	}

	// Pass request to store
	return st.store.put(key, value)
}

func (st *StateRW) PutStream(key string, r io.Reader) error {
	// Check not closed
	if st.store == nil {
		return ErrStateClosed
	}

	// Pass request to store
	return st.store.putStream(key, r)
}

func (st *StateRW) Has(key string) (bool, error) {
	// Check not closed
	if st.store == nil {
		return false, ErrStateClosed
	}

	// Pass request to store
	return st.store.has(key)
}

func (st *StateRW) Delete(key string) error {
	// Check not closed
	if st.store == nil {
		return ErrStateClosed
	}

	// Pass request to store
	return st.store.delete(key)
}

func (st *StateRW) close() {
	st.store = nil
}
