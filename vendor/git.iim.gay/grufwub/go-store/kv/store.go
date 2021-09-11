package kv

import (
	"io"
	"sync"

	"git.iim.gay/grufwub/go-mutexes"
	"git.iim.gay/grufwub/go-store/storage"
	"git.iim.gay/grufwub/go-store/util"
)

// KVStore is a very simple, yet performant key-value store
type KVStore struct {
	mutexMap mutexes.MutexMap // mutexMap is a map of keys to mutexes to protect file access
	mutex    sync.RWMutex     // mutex is the total store mutex
	storage  storage.Storage  // storage is the underlying storage
}

func OpenFile(path string, cfg *storage.DiskConfig) (*KVStore, error) {
	// Attempt to open disk storage
	storage, err := storage.OpenFile(path, cfg)
	if err != nil {
		return nil, err
	}

	// Return new KVStore
	return OpenStorage(storage)
}

func OpenBlock(path string, cfg *storage.BlockConfig) (*KVStore, error) {
	// Attempt to open block storage
	storage, err := storage.OpenBlock(path, cfg)
	if err != nil {
		return nil, err
	}

	// Return new KVStore
	return OpenStorage(storage)
}

func OpenStorage(storage storage.Storage) (*KVStore, error) {
	// Perform initial storage clean
	err := storage.Clean()
	if err != nil {
		return nil, err
	}

	// Return new KVStore
	return &KVStore{
		mutexMap: mutexes.NewMap(mutexes.NewRW),
		mutex:    sync.RWMutex{},
		storage:  storage,
	}, nil
}

// Get fetches the bytes for supplied key in the store
func (st *KVStore) Get(key string) ([]byte, error) {
	// Acquire store read lock
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	// Pass to unprotected fn
	return st.get(key)
}

func (st *KVStore) get(key string) ([]byte, error) {
	// Acquire read lock for key
	runlock := st.mutexMap.RLock(key)
	defer runlock()

	// Read file bytes
	return st.storage.ReadBytes(key)
}

// GetStream fetches a ReadCloser for the bytes at the supplied key location in the store
func (st *KVStore) GetStream(key string) (io.ReadCloser, error) {
	// Acquire store read lock
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	// Pass to unprotected fn
	return st.getStream(key)
}

func (st *KVStore) getStream(key string) (io.ReadCloser, error) {
	// Acquire read lock for key
	runlock := st.mutexMap.RLock(key)

	// Attempt to open stream for read
	rd, err := st.storage.ReadStream(key)
	if err != nil {
		runlock()
		return nil, err
	}

	// Wrap readcloser in our own callback closer
	return util.ReadCloserWithCallback(rd, runlock), nil
}

// Put places the bytes at the supplied key location in the store
func (st *KVStore) Put(key string, value []byte) error {
	// Acquire store write lock
	st.mutex.Lock()
	defer st.mutex.Unlock()

	// Pass to unprotected fn
	return st.put(key, value)
}

func (st *KVStore) put(key string, value []byte) error {
	// Acquire write lock for key
	unlock := st.mutexMap.Lock(key)
	defer unlock()

	// Write file bytes
	return st.storage.WriteBytes(key, value)
}

// PutStream writes the bytes from the supplied Reader at the supplied key location in the store
func (st *KVStore) PutStream(key string, r io.Reader) error {
	// Acquire store write lock
	st.mutex.Lock()
	defer st.mutex.Unlock()

	// Pass to unprotected fn
	return st.putStream(key, r)
}

func (st *KVStore) putStream(key string, r io.Reader) error {
	// Acquire write lock for key
	unlock := st.mutexMap.Lock(key)
	defer unlock()

	// Write file stream
	return st.storage.WriteStream(key, r)
}

// Has checks whether the supplied key exists in the store
func (st *KVStore) Has(key string) (bool, error) {
	// Acquire store read lock
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	// Pass to unprotected fn
	return st.has(key)
}

func (st *KVStore) has(key string) (bool, error) {
	// Acquire read lock for key
	runlock := st.mutexMap.RLock(key)
	defer runlock()

	// Stat file on disk
	return st.storage.Stat(key)
}

// Delete removes the supplied key-value pair from the store
func (st *KVStore) Delete(key string) error {
	// Acquire store write lock
	st.mutex.Lock()
	defer st.mutex.Unlock()

	// Pass to unprotected fn
	return st.delete(key)
}

func (st *KVStore) delete(key string) error {
	// Acquire write lock for key
	unlock := st.mutexMap.Lock(key)
	defer unlock()

	// Remove file from disk
	return st.storage.Remove(key)
}

// Iterator returns an Iterator for key-value pairs in the store, using supplied match function
func (st *KVStore) Iterator(matchFn func(string) bool) (*KVIterator, error) {
	// If no function, match all
	if matchFn == nil {
		matchFn = func(string) bool { return true }
	}

	// Get store read lock
	st.mutex.RLock()

	// Setup the walk keys function
	entries := []storage.StorageEntry{}
	walkFn := func(entry storage.StorageEntry) {
		// Ignore unmatched entries
		if !matchFn(entry.Key()) {
			return
		}

		// Add to entries
		entries = append(entries, entry)
	}

	// Walk keys in the storage
	err := st.storage.WalkKeys(&storage.WalkKeysOptions{WalkFn: walkFn})
	if err != nil {
		st.mutex.RUnlock()
		return nil, err
	}

	// Return new iterator
	return &KVIterator{
		store:   st,
		entries: entries,
		index:   -1,
		key:     "",
		onClose: st.mutex.RUnlock,
	}, nil
}

// Read provides a read-only window to the store, holding it in a read-locked state until
// the supplied function returns
func (st *KVStore) Read(do func(*StateRO)) {
	// Get store read lock
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	// Create new store state (defer close)
	state := &StateRO{store: st}
	defer state.close()

	// Pass state
	do(state)
}

// Update provides a read-write window to the store, holding it in a read-write-locked state
// until the supplied functions returns
func (st *KVStore) Update(do func(*StateRW)) {
	// Get store lock
	st.mutex.Lock()
	defer st.mutex.Unlock()

	// Create new store state (defer close)
	state := &StateRW{store: st}
	defer state.close()

	// Pass state
	do(state)
}
