package storage

import (
	"context"
	"io"
	"sync/atomic"

	"codeberg.org/gruf/go-bytes"
	"codeberg.org/gruf/go-iotools"
	"github.com/cornelk/hashmap"
)

// MemoryStorage is a storage implementation that simply stores key-value
// pairs in a Go map in-memory. The map is protected by a mutex.
type MemoryStorage struct {
	ow bool // overwrites
	fs *hashmap.Map[string, []byte]
	st uint32
}

// OpenMemory opens a new MemoryStorage instance with internal map starting size.
func OpenMemory(size int, overwrites bool) *MemoryStorage {
	if size <= 0 {
		size = 8
	}
	return &MemoryStorage{
		fs: hashmap.NewSized[string, []byte](uintptr(size)),
		ow: overwrites,
	}
}

// Clean implements Storage.Clean().
func (st *MemoryStorage) Clean(ctx context.Context) error {
	// Check store open
	if st.closed() {
		return ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}

// ReadBytes implements Storage.ReadBytes().
func (st *MemoryStorage) ReadBytes(ctx context.Context, key string) ([]byte, error) {
	// Check store open
	if st.closed() {
		return nil, ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Check for key in store
	b, ok := st.fs.Get(key)
	if !ok {
		return nil, ErrNotFound
	}

	// Create return copy
	return copyb(b), nil
}

// ReadStream implements Storage.ReadStream().
func (st *MemoryStorage) ReadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	// Check store open
	if st.closed() {
		return nil, ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Check for key in store
	b, ok := st.fs.Get(key)
	if !ok {
		return nil, ErrNotFound
	}

	// Create io.ReadCloser from 'b' copy
	r := bytes.NewReader(copyb(b))
	return iotools.NopReadCloser(r), nil
}

// WriteBytes implements Storage.WriteBytes().
func (st *MemoryStorage) WriteBytes(ctx context.Context, key string, b []byte) (int, error) {
	// Check store open
	if st.closed() {
		return 0, ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	// Check for key that already exists
	if _, ok := st.fs.Get(key); ok && !st.ow {
		return 0, ErrAlreadyExists
	}

	// Write key copy to store
	st.fs.Set(key, copyb(b))
	return len(b), nil
}

// WriteStream implements Storage.WriteStream().
func (st *MemoryStorage) WriteStream(ctx context.Context, key string, r io.Reader) (int64, error) {
	// Check store open
	if st.closed() {
		return 0, ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	// Check for key that already exists
	if _, ok := st.fs.Get(key); ok && !st.ow {
		return 0, ErrAlreadyExists
	}

	// Read all from reader
	b, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}

	// Write key to store
	st.fs.Set(key, b)
	return int64(len(b)), nil
}

// Stat implements Storage.Stat().
func (st *MemoryStorage) Stat(ctx context.Context, key string) (bool, error) {
	// Check store open
	if st.closed() {
		return false, ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return false, err
	}

	// Check for key in store
	_, ok := st.fs.Get(key)
	return ok, nil
}

// Remove implements Storage.Remove().
func (st *MemoryStorage) Remove(ctx context.Context, key string) error {
	// Check store open
	if st.closed() {
		return ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return err
	}

	// Attempt to delete key
	ok := st.fs.Del(key)
	if !ok {
		return ErrNotFound
	}

	return nil
}

// WalkKeys implements Storage.WalkKeys().
func (st *MemoryStorage) WalkKeys(ctx context.Context, opts WalkKeysOptions) error {
	// Check store open
	if st.closed() {
		return ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return err
	}

	var err error

	// Nil check func
	_ = opts.WalkFn

	// Pass each key in map to walk function
	st.fs.Range(func(key string, val []byte) bool {
		err = opts.WalkFn(ctx, Entry{
			Key:  key,
			Size: int64(len(val)),
		})
		return (err == nil)
	})

	return err
}

// Close implements Storage.Close().
func (st *MemoryStorage) Close() error {
	atomic.StoreUint32(&st.st, 1)
	return nil
}

// closed returns whether MemoryStorage is closed.
func (st *MemoryStorage) closed() bool {
	return (atomic.LoadUint32(&st.st) == 1)
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
