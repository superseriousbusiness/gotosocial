package storage

import (
	"context"
	"io"
)

// Storage defines a means of storing and accessing key value pairs
type Storage interface {

	// ReadBytes returns the byte value for key in storage.
	ReadBytes(ctx context.Context, key string) ([]byte, error)

	// ReadStream returns an io.ReadCloser for the data at key in the storage.
	ReadStream(ctx context.Context, key string) (io.ReadCloser, error)

	// WriteBytes writes the supplied data at key in the storage.
	WriteBytes(ctx context.Context, key string, data []byte) (int, error)

	// WriteStream writes the bytes from supplied reader at key in the storage.
	WriteStream(ctx context.Context, key string, stream io.Reader) (int64, error)

	// Stat checks if the supplied key is in the storage.
	Stat(ctx context.Context, key string) (*Entry, error)

	// Remove attempts to remove the supplied key-value pair from storage.
	Remove(ctx context.Context, key string) error

	// Clean removes unused values and unclutters the storage (e.g. removing empty folders)
	Clean(ctx context.Context) error

	// WalkKeys walks available keys using opts in storage implementation.
	WalkKeys(ctx context.Context, opts WalkKeysOpts) error
}

// Entry represents a key in a Storage{} implementation,
// with any associated metadata that may have been set.
type Entry struct {

	// Key is this entry's
	// unique storage key.
	Key string

	// Size is the size of
	// this entry in storage.
	Size int64
}

// WalkKeysOpts are arguments provided
// to a storage WalkKeys() implementation.
type WalkKeysOpts struct {

	// Prefix can be used to filter entries
	// by the given key prefix, for example
	// only those under a subdirectory. This
	// is preferred over Filter() function.
	Prefix string

	// Filter can be used to filter entries
	// by any custom metric before before they
	// are passed to Step() function. E.g.
	// filter storage entries by regexp.
	Filter func(string) bool

	// Step is called for each entry during
	// WalkKeys, error triggers early return.
	Step func(Entry) error
}
