package kv

import (
	"context"
	"errors"
	"io"

	"codeberg.org/gruf/go-mutexes"
)

// ErrStateClosed is returned on further calls to states after calling Release().
var ErrStateClosed = errors.New("store/kv: state closed")

// StateRO provides a read-only window to the store. While this
// state is active during the Read() function window, the entire
// store will be read-locked. The state is thread-safe for concurrent
// use UNTIL the moment that your supplied function to Read() returns.
type StateRO struct {
	store *KVStore
	state *mutexes.LockState
}

// Get: see KVStore.Get(). Returns error if state already closed.
func (st *StateRO) Get(ctx context.Context, key string) ([]byte, error) {
	if st.store == nil {
		return nil, ErrStateClosed
	}
	return st.store.get(st.state.RLock, ctx, key)
}

// GetStream: see KVStore.GetStream(). Returns error if state already closed.
func (st *StateRO) GetStream(ctx context.Context, key string) (io.ReadCloser, error) {
	if st.store == nil {
		return nil, ErrStateClosed
	}
	return st.store.getStream(st.state.RLock, ctx, key)
}

// Has: see KVStore.Has(). Returns error if state already closed.
func (st *StateRO) Has(ctx context.Context, key string) (bool, error) {
	if st.store == nil {
		return false, ErrStateClosed
	}
	return st.store.has(st.state.RLock, ctx, key)
}

// Release will release the store read-lock, and close this state.
func (st *StateRO) Release() {
	st.state.UnlockMap()
	st.state = nil
	st.store = nil
}

// StateRW provides a read-write window to the store. While this
// state is active during the Update() function window, the entire
// store will be locked. The state is thread-safe for concurrent
// use UNTIL the moment that your supplied function to Update() returns.
type StateRW struct {
	store *KVStore
	state *mutexes.LockState
}

// Get: see KVStore.Get(). Returns error if state already closed.
func (st *StateRW) Get(ctx context.Context, key string) ([]byte, error) {
	if st.store == nil {
		return nil, ErrStateClosed
	}
	return st.store.get(st.state.RLock, ctx, key)
}

// GetStream: see KVStore.GetStream(). Returns error if state already closed.
func (st *StateRW) GetStream(ctx context.Context, key string) (io.ReadCloser, error) {
	if st.store == nil {
		return nil, ErrStateClosed
	}
	return st.store.getStream(st.state.RLock, ctx, key)
}

// Put: see KVStore.Put(). Returns error if state already closed.
func (st *StateRW) Put(ctx context.Context, key string, value []byte) error {
	if st.store == nil {
		return ErrStateClosed
	}
	return st.store.put(st.state.Lock, ctx, key, value)
}

// PutStream: see KVStore.PutStream(). Returns error if state already closed.
func (st *StateRW) PutStream(ctx context.Context, key string, r io.Reader) error {
	if st.store == nil {
		return ErrStateClosed
	}
	return st.store.putStream(st.state.Lock, ctx, key, r)
}

// Has: see KVStore.Has(). Returns error if state already closed.
func (st *StateRW) Has(ctx context.Context, key string) (bool, error) {
	if st.store == nil {
		return false, ErrStateClosed
	}
	return st.store.has(st.state.RLock, ctx, key)
}

// Delete: see KVStore.Delete(). Returns error if state already closed.
func (st *StateRW) Delete(ctx context.Context, key string) error {
	if st.store == nil {
		return ErrStateClosed
	}
	return st.store.delete(st.state.Lock, ctx, key)
}

// Release will release the store lock, and close this state.
func (st *StateRW) Release() {
	st.state.UnlockMap()
	st.state = nil
	st.store = nil
}
