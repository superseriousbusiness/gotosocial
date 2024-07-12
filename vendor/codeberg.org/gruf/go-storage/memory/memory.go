package memory

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"

	"codeberg.org/gruf/go-storage"

	"codeberg.org/gruf/go-storage/internal"
)

// ensure MemoryStorage conforms to storage.Storage.
var _ storage.Storage = (*MemoryStorage)(nil)

// MemoryStorage is a storage implementation that simply stores key-value
// pairs in a Go map in-memory. The map is protected by a mutex.
type MemoryStorage struct {
	ow bool // overwrites
	fs map[string][]byte
	mu sync.Mutex
}

// Open opens a new MemoryStorage instance with internal map starting size.
func Open(size int, overwrites bool) *MemoryStorage {
	return &MemoryStorage{
		ow: overwrites,
		fs: make(map[string][]byte, size),
	}
}

// Clean: implements Storage.Clean().
func (st *MemoryStorage) Clean(ctx context.Context) error {
	// Check context still valid
	if err := ctx.Err(); err != nil {
		return err
	}

	// Lock map.
	st.mu.Lock()

	// Resize map to only necessary size in-mem.
	fs := make(map[string][]byte, len(st.fs))
	for key, val := range st.fs {
		fs[key] = val
	}
	st.fs = fs

	// Done with lock.
	st.mu.Unlock()

	return nil
}

// ReadBytes: implements Storage.ReadBytes().
func (st *MemoryStorage) ReadBytes(ctx context.Context, key string) ([]byte, error) {
	// Check context still valid.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Lock map.
	st.mu.Lock()

	// Check key in store.
	b, ok := st.fs[key]
	if ok {

		// COPY bytes.
		b = copyb(b)
	}

	// Done with lock.
	st.mu.Unlock()

	if !ok {
		return nil, internal.ErrWithKey(storage.ErrNotFound, key)
	}

	return b, nil
}

// ReadStream: implements Storage.ReadStream().
func (st *MemoryStorage) ReadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	// Read value data from store.
	b, err := st.ReadBytes(ctx, key)
	if err != nil {
		return nil, err
	}

	// Wrap in readcloser.
	r := bytes.NewReader(b)
	return io.NopCloser(r), nil
}

// WriteBytes: implements Storage.WriteBytes().
func (st *MemoryStorage) WriteBytes(ctx context.Context, key string, b []byte) (int, error) {
	// Check context still valid
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	// Lock map.
	st.mu.Lock()

	// Check key in store.
	_, ok := st.fs[key]

	if ok && !st.ow {
		// Done with lock.
		st.mu.Unlock()

		// Overwrites are disabled, return existing key error.
		return 0, internal.ErrWithKey(storage.ErrAlreadyExists, key)
	}

	// Write copy to store.
	st.fs[key] = copyb(b)

	// Done with lock.
	st.mu.Unlock()

	return len(b), nil
}

// WriteStream: implements Storage.WriteStream().
func (st *MemoryStorage) WriteStream(ctx context.Context, key string, r io.Reader) (int64, error) {
	// Read all from reader.
	b, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}

	// Write in-memory data to store.
	n, err := st.WriteBytes(ctx, key, b)
	return int64(n), err
}

// Stat: implements Storage.Stat().
func (st *MemoryStorage) Stat(ctx context.Context, key string) (*storage.Entry, error) {
	// Check context still valid
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Lock map.
	st.mu.Lock()

	// Check key in store.
	b, ok := st.fs[key]

	// Get entry size.
	sz := int64(len(b))

	// Done with lock.
	st.mu.Unlock()

	if !ok {
		return nil, nil
	}

	return &storage.Entry{
		Key:  key,
		Size: sz,
	}, nil
}

// Remove: implements Storage.Remove().
func (st *MemoryStorage) Remove(ctx context.Context, key string) error {
	// Check context still valid
	if err := ctx.Err(); err != nil {
		return err
	}

	// Lock map.
	st.mu.Lock()

	// Check key in store.
	_, ok := st.fs[key]

	if ok {
		// Delete store key.
		delete(st.fs, key)
	}

	// Done with lock.
	st.mu.Unlock()

	if !ok {
		return internal.ErrWithKey(storage.ErrNotFound, key)
	}

	return nil
}

// WalkKeys: implements Storage.WalkKeys().
func (st *MemoryStorage) WalkKeys(ctx context.Context, opts storage.WalkKeysOpts) error {
	if opts.Step == nil {
		panic("nil step fn")
	}

	// Check context still valid.
	if err := ctx.Err(); err != nil {
		return err
	}

	var err error

	// Lock map.
	st.mu.Lock()

	// Ensure unlocked.
	defer st.mu.Unlock()

	// Range all key-vals in hash map.
	for key, val := range st.fs {
		// Check for filtered prefix.
		if opts.Prefix != "" &&
			!strings.HasPrefix(key, opts.Prefix) {
			continue // ignore
		}

		// Check for filtered key.
		if opts.Filter != nil &&
			!opts.Filter(key) {
			continue // ignore
		}

		// Pass to provided step func.
		err = opts.Step(storage.Entry{
			Key:  key,
			Size: int64(len(val)),
		})
		if err != nil {
			return err
		}
	}

	return err
}

// copyb returns a copy of byte-slice b.
func copyb(b []byte) []byte {
	if b == nil {
		return nil
	}
	p := make([]byte, len(b))
	_ = copy(p, b)
	return p
}
