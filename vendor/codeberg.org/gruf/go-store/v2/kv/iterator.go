package kv

import (
	"context"
	"errors"

	"codeberg.org/gruf/go-mutexes"
	"codeberg.org/gruf/go-store/v2/storage"
)

var ErrIteratorClosed = errors.New("store/kv: iterator closed")

// Iterator provides a read-only iterator to all the key-value
// pairs in a KVStore. While the iterator is open the store is read
// locked, you MUST release the iterator when you are finished with
// it.
//
// Please note:
// individual iterators are NOT concurrency safe, though it is safe to
// have multiple iterators running concurrently.
type Iterator struct {
	store   *KVStore // store is the linked KVStore
	state   *mutexes.LockState
	entries []storage.Entry
	index   int
	key     string
}

// Next attempts to fetch the next key-value pair, the
// return value indicates whether another pair remains.
func (i *Iterator) Next() bool {
	next := i.index + 1
	if next >= len(i.entries) {
		i.key = ""
		return false
	}
	i.key = i.entries[next].Key
	i.index = next
	return true
}

// Key returns the current iterator key.
func (i *Iterator) Key() string {
	return i.key
}

// Value returns the current iterator value at key.
func (i *Iterator) Value(ctx context.Context) ([]byte, error) {
	if i.store == nil {
		return nil, ErrIteratorClosed
	}
	return i.store.get(i.state.RLock, ctx, i.key)
}

// Release will release the store read-lock, and close this iterator.
func (i *Iterator) Release() {
	i.state.UnlockMap()
	i.state = nil
	i.store = nil
	i.key = ""
	i.entries = nil
	i.index = 0
}
