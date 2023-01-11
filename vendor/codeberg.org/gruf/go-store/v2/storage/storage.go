package storage

import (
	"context"
	"io"
)

// Storage defines a means of storing and accessing key value pairs
type Storage interface {
	// ReadBytes returns the byte value for key in storage
	ReadBytes(ctx context.Context, key string) ([]byte, error)

	// ReadStream returns an io.ReadCloser for the value bytes at key in the storage
	ReadStream(ctx context.Context, key string) (io.ReadCloser, error)

	// WriteBytes writes the supplied value bytes at key in the storage
	WriteBytes(ctx context.Context, key string, value []byte) (int, error)

	// WriteStream writes the bytes from supplied reader at key in the storage
	WriteStream(ctx context.Context, key string, r io.Reader) (int64, error)

	// Stat checks if the supplied key is in the storage
	Stat(ctx context.Context, key string) (bool, error)

	// Remove attempts to remove the supplied key-value pair from storage
	Remove(ctx context.Context, key string) error

	// Close will close the storage, releasing any file locks
	Close() error

	// Clean removes unused values and unclutters the storage (e.g. removing empty folders)
	Clean(ctx context.Context) error

	// WalkKeys walks the keys in the storage
	WalkKeys(ctx context.Context, opts WalkKeysOptions) error
}

// Entry represents a key in a Storage{} implementation,
// with any associated metadata that may have been set.
type Entry struct {
	// Key is this entry's unique storage key.
	Key string

	// Size is the size of this entry in storage.
	// Note that size < 0 indicates unknown.
	Size int64
}

// WalkKeysOptions defines how to walk the keys in a storage implementation
type WalkKeysOptions struct {
	// WalkFn is the function to apply on each StorageEntry
	WalkFn func(context.Context, Entry) error
}
