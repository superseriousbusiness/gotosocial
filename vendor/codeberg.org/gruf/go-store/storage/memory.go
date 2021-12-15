package storage

import (
	"io"

	"codeberg.org/gruf/go-bytes"
	"codeberg.org/gruf/go-store/util"
)

// MemoryStorage is a storage implementation that simply stores key-value
// pairs in a Go map in-memory. The map is protected by a mutex.
type MemoryStorage struct {
	fs map[string][]byte
}

// OpenMemory opens a new MemoryStorage instance with internal map of 'size'.
func OpenMemory(size int) *MemoryStorage {
	return &MemoryStorage{
		fs: make(map[string][]byte, size),
	}
}

// Clean implements Storage.Clean().
func (st *MemoryStorage) Clean() error {
	return nil
}

// ReadBytes implements Storage.ReadBytes().
func (st *MemoryStorage) ReadBytes(key string) ([]byte, error) {
	b, ok := st.fs[key]
	if !ok {
		return nil, ErrNotFound
	}
	return bytes.Copy(b), nil
}

// ReadStream implements Storage.ReadStream().
func (st *MemoryStorage) ReadStream(key string) (io.ReadCloser, error) {
	b, ok := st.fs[key]
	if !ok {
		return nil, ErrNotFound
	}
	b = bytes.Copy(b)
	r := bytes.NewReader(b)
	return util.NopReadCloser(r), nil
}

// WriteBytes implements Storage.WriteBytes().
func (st *MemoryStorage) WriteBytes(key string, b []byte) error {
	_, ok := st.fs[key]
	if ok {
		return ErrAlreadyExists
	}
	st.fs[key] = bytes.Copy(b)
	return nil
}

// WriteStream implements Storage.WriteStream().
func (st *MemoryStorage) WriteStream(key string, r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return st.WriteBytes(key, b)
}

// Stat implements Storage.Stat().
func (st *MemoryStorage) Stat(key string) (bool, error) {
	_, ok := st.fs[key]
	return ok, nil
}

// Remove implements Storage.Remove().
func (st *MemoryStorage) Remove(key string) error {
	_, ok := st.fs[key]
	if !ok {
		return ErrNotFound
	}
	delete(st.fs, key)
	return nil
}

// WalkKeys implements Storage.WalkKeys().
func (st *MemoryStorage) WalkKeys(opts WalkKeysOptions) error {
	for key := range st.fs {
		opts.WalkFn(entry(key))
	}
	return nil
}
