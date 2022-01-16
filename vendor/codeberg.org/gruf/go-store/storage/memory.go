package storage

import (
	"io"
	"sync"

	"codeberg.org/gruf/go-bytes"
	"codeberg.org/gruf/go-store/util"
)

// MemoryStorage is a storage implementation that simply stores key-value
// pairs in a Go map in-memory. The map is protected by a mutex.
type MemoryStorage struct {
	ow bool // overwrites
	fs map[string][]byte
	mu sync.Mutex
}

// OpenMemory opens a new MemoryStorage instance with internal map of 'size'.
func OpenMemory(size int, overwrites bool) *MemoryStorage {
	return &MemoryStorage{
		fs: make(map[string][]byte, size),
		mu: sync.Mutex{},
		ow: overwrites,
	}
}

// Clean implements Storage.Clean().
func (st *MemoryStorage) Clean() error {
	return nil
}

// ReadBytes implements Storage.ReadBytes().
func (st *MemoryStorage) ReadBytes(key string) ([]byte, error) {
	// Safely check store
	st.mu.Lock()
	b, ok := st.fs[key]
	st.mu.Unlock()

	// Return early if not exist
	if !ok {
		return nil, ErrNotFound
	}

	// Create return copy
	return bytes.Copy(b), nil
}

// ReadStream implements Storage.ReadStream().
func (st *MemoryStorage) ReadStream(key string) (io.ReadCloser, error) {
	// Safely check store
	st.mu.Lock()
	b, ok := st.fs[key]
	st.mu.Unlock()

	// Return early if not exist
	if !ok {
		return nil, ErrNotFound
	}

	// Create io.ReadCloser from 'b' copy
	b = bytes.Copy(b)
	r := bytes.NewReader(b)
	return util.NopReadCloser(r), nil
}

// WriteBytes implements Storage.WriteBytes().
func (st *MemoryStorage) WriteBytes(key string, b []byte) error {
	// Safely check store
	st.mu.Lock()
	_, ok := st.fs[key]

	// Check for already exist
	if ok && !st.ow {
		st.mu.Unlock()
		return ErrAlreadyExists
	}

	// Write + unlock
	st.fs[key] = bytes.Copy(b)
	st.mu.Unlock()
	return nil
}

// WriteStream implements Storage.WriteStream().
func (st *MemoryStorage) WriteStream(key string, r io.Reader) error {
	// Read all from reader
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// Write to storage
	return st.WriteBytes(key, b)
}

// Stat implements Storage.Stat().
func (st *MemoryStorage) Stat(key string) (bool, error) {
	st.mu.Lock()
	_, ok := st.fs[key]
	st.mu.Unlock()
	return ok, nil
}

// Remove implements Storage.Remove().
func (st *MemoryStorage) Remove(key string) error {
	// Safely check store
	st.mu.Lock()
	_, ok := st.fs[key]

	// Check in store
	if !ok {
		st.mu.Unlock()
		return ErrNotFound
	}

	// Delete + unlock
	delete(st.fs, key)
	st.mu.Unlock()
	return nil
}

// Close implements Storage.Close().
func (st *MemoryStorage) Close() error {
	return nil
}

// WalkKeys implements Storage.WalkKeys().
func (st *MemoryStorage) WalkKeys(opts WalkKeysOptions) error {
	// Safely walk storage keys
	st.mu.Lock()
	for key := range st.fs {
		opts.WalkFn(entry(key))
	}
	st.mu.Unlock()

	return nil
}
