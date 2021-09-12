package kv

import (
	"git.iim.gay/grufwub/go-errors"
	"git.iim.gay/grufwub/go-store/storage"
)

var ErrIteratorClosed = errors.Define("store/kv: iterator closed")

// KVIterator provides a read-only iterator to all the key-value
// pairs in a KVStore. While the iterator is open the store is read
// locked, you MUST release the iterator when you are finished with
// it.
//
// Please note:
// - individual iterators are NOT concurrency safe, though it is safe to
// have multiple iterators running concurrently
type KVIterator struct {
	store   *KVStore // store is the linked KVStore
	entries []storage.StorageEntry
	index   int
	key     string
	onClose func()
}

// Next attempts to set the next key-value pair, the
// return value is if there was another pair remaining
func (i *KVIterator) Next() bool {
	next := i.index + 1
	if next >= len(i.entries) {
		i.key = ""
		return false
	}
	i.key = i.entries[next].Key()
	i.index = next
	return true
}

// Key returns the next key from the store
func (i *KVIterator) Key() string {
	return i.key
}

// Release releases the KVIterator and KVStore's read lock
func (i *KVIterator) Release() {
	// Reset key, path, entries
	i.store = nil
	i.key = ""
	i.entries = nil

	// Perform requested callback
	i.onClose()
}

// Value returns the next value from the KVStore
func (i *KVIterator) Value() ([]byte, error) {
	// Check store isn't closed
	if i.store == nil {
		return nil, ErrIteratorClosed
	}

	// Attempt to fetch from store
	return i.store.get(i.key)
}
