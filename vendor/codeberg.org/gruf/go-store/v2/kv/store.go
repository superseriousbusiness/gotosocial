package kv

import (
	"context"
	"io"

	"codeberg.org/gruf/go-iotools"
	"codeberg.org/gruf/go-mutexes"
	"codeberg.org/gruf/go-store/v2/storage"
)

// KVStore is a very simple, yet performant key-value store
type KVStore struct {
	mu mutexes.MutexMap // map of keys to mutexes to protect key access
	st storage.Storage  // underlying storage implementation
}

func OpenDisk(path string, cfg *storage.DiskConfig) (*KVStore, error) {
	// Attempt to open disk storage
	storage, err := storage.OpenDisk(path, cfg)
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

func OpenMemory(overwrites bool) *KVStore {
	return New(storage.OpenMemory(100, overwrites))
}

func OpenS3(endpoint string, bucket string, cfg *storage.S3Config) (*KVStore, error) {
	// Attempt to open S3 storage
	storage, err := storage.OpenS3(endpoint, bucket, cfg)
	if err != nil {
		return nil, err
	}

	// Return new KVStore
	return OpenStorage(storage)
}

// OpenStorage will return a new KVStore instance based on Storage, performing an initial storage.Clean().
func OpenStorage(storage storage.Storage) (*KVStore, error) {
	// Perform initial storage clean
	err := storage.Clean(context.Background())
	if err != nil {
		return nil, err
	}

	// Return new KVStore
	return New(storage), nil
}

// New will simply return a new KVStore instance based on Storage.
func New(storage storage.Storage) *KVStore {
	if storage == nil {
		panic("nil storage")
	}
	return &KVStore{
		mu: mutexes.NewMap(-1, -1),
		st: storage,
	}
}

// RLock acquires a read-lock on supplied key, returning unlock function.
func (st *KVStore) RLock(key string) (runlock func()) {
	return st.mu.RLock(key)
}

// Lock acquires a write-lock on supplied key, returning unlock function.
func (st *KVStore) Lock(key string) (unlock func()) {
	return st.mu.Lock(key)
}

// Get fetches the bytes for supplied key in the store.
func (st *KVStore) Get(ctx context.Context, key string) ([]byte, error) {
	return st.get(st.RLock, ctx, key)
}

// get performs the underlying logic for KVStore.Get(), using supplied read lock func to allow use with states.
func (st *KVStore) get(rlock func(string) func(), ctx context.Context, key string) ([]byte, error) {
	// Acquire read lock for key
	runlock := rlock(key)
	defer runlock()

	// Read file bytes from storage
	return st.st.ReadBytes(ctx, key)
}

// GetStream fetches a ReadCloser for the bytes at the supplied key in the store.
func (st *KVStore) GetStream(ctx context.Context, key string) (io.ReadCloser, error) {
	return st.getStream(st.RLock, ctx, key)
}

// getStream performs the underlying logic for KVStore.GetStream(), using supplied read lock func to allow use with states.
func (st *KVStore) getStream(rlock func(string) func(), ctx context.Context, key string) (io.ReadCloser, error) {
	// Acquire read lock for key
	runlock := rlock(key)

	// Attempt to open stream for read
	rd, err := st.st.ReadStream(ctx, key)
	if err != nil {
		runlock()
		return nil, err
	}

	var unlocked bool

	// Wrap readcloser to call our own callback
	return iotools.ReadCloser(rd, iotools.CloserFunc(func() error {
		if !unlocked {
			unlocked = true
			defer runlock()
		}
		return rd.Close()
	})), nil
}

// Put places the bytes at the supplied key in the store.
func (st *KVStore) Put(ctx context.Context, key string, value []byte) (int, error) {
	return st.put(st.Lock, ctx, key, value)
}

// put performs the underlying logic for KVStore.Put(), using supplied lock func to allow use with states.
func (st *KVStore) put(lock func(string) func(), ctx context.Context, key string, value []byte) (int, error) {
	// Acquire write lock for key
	unlock := lock(key)
	defer unlock()

	// Write file bytes to storage
	return st.st.WriteBytes(ctx, key, value)
}

// PutStream writes the bytes from the supplied Reader at the supplied key in the store.
func (st *KVStore) PutStream(ctx context.Context, key string, r io.Reader) (int64, error) {
	return st.putStream(st.Lock, ctx, key, r)
}

// putStream performs the underlying logic for KVStore.PutStream(), using supplied lock func to allow use with states.
func (st *KVStore) putStream(lock func(string) func(), ctx context.Context, key string, r io.Reader) (int64, error) {
	// Acquire write lock for key
	unlock := lock(key)
	defer unlock()

	// Write file stream to storage
	return st.st.WriteStream(ctx, key, r)
}

// Has checks whether the supplied key exists in the store.
func (st *KVStore) Has(ctx context.Context, key string) (bool, error) {
	return st.has(st.RLock, ctx, key)
}

// has performs the underlying logic for KVStore.Has(), using supplied read lock func to allow use with states.
func (st *KVStore) has(rlock func(string) func(), ctx context.Context, key string) (bool, error) {
	// Acquire read lock for key
	runlock := rlock(key)
	defer runlock()

	// Stat file in storage
	return st.st.Stat(ctx, key)
}

// Delete removes value at supplied key from the store.
func (st *KVStore) Delete(ctx context.Context, key string) error {
	return st.delete(st.Lock, ctx, key)
}

// delete performs the underlying logic for KVStore.Delete(), using supplied lock func to allow use with states.
func (st *KVStore) delete(lock func(string) func(), ctx context.Context, key string) error {
	// Acquire write lock for key
	unlock := lock(key)
	defer unlock()

	// Remove file from storage
	return st.st.Remove(ctx, key)
}

// Iterator returns an Iterator for key-value pairs in the store, using supplied match function
func (st *KVStore) Iterator(ctx context.Context, matchFn func(string) bool) (*Iterator, error) {
	if matchFn == nil {
		// By default simply match all keys
		matchFn = func(string) bool { return true }
	}

	// Get store read lock state
	state := st.mu.RLockMap()

	var entries []storage.Entry

	walkFn := func(ctx context.Context, entry storage.Entry) error {
		// Ignore unmatched entries
		if !matchFn(entry.Key) {
			return nil
		}

		// Add to entries
		entries = append(entries, entry)
		return nil
	}

	// Collate keys in storage with our walk function
	err := st.st.WalkKeys(ctx, storage.WalkKeysOptions{WalkFn: walkFn})
	if err != nil {
		state.UnlockMap()
		return nil, err
	}

	// Return new iterator
	return &Iterator{
		store:   st,
		state:   state,
		entries: entries,
		index:   -1,
		key:     "",
	}, nil
}

// Read provides a read-only window to the store, holding it in a read-locked state until release.
func (st *KVStore) Read() *StateRO {
	state := st.mu.RLockMap()
	return &StateRO{store: st, state: state}
}

// ReadFn provides a read-only window to the store, holding it in a read-locked state until fn return..
func (st *KVStore) ReadFn(fn func(*StateRO)) {
	// Acquire read-only state
	state := st.Read()
	defer state.Release()

	// Pass to fn
	fn(state)
}

// Update provides a read-write window to the store, holding it in a write-locked state until release.
func (st *KVStore) Update() *StateRW {
	state := st.mu.LockMap()
	return &StateRW{store: st, state: state}
}

// UpdateFn provides a read-write window to the store, holding it in a write-locked state until fn return.
func (st *KVStore) UpdateFn(fn func(*StateRW)) {
	// Acquire read-write state
	state := st.Update()
	defer state.Release()

	// Pass to fn
	fn(state)
}

// Close will close the underlying storage, the mutex map locking (e.g. RLock(), Lock()) will continue to function.
func (st *KVStore) Close() error {
	return st.st.Close()
}
