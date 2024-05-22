package storage

import (
	"context"
	"io"
)

// Storage defines a means of accessing and storing
// data to some abstracted underlying mechanism. Whether
// that be in-memory, an on-disk filesystem or S3 bucket.
type Storage interface {

	// ReadBytes returns the data located at key (e.g. filepath) in storage.
	ReadBytes(ctx context.Context, key string) ([]byte, error)

	// ReadStream returns an io.ReadCloser for the data at key (e.g. filepath) in storage.
	ReadStream(ctx context.Context, key string) (io.ReadCloser, error)

	// WriteBytes writes the supplied data at key (e.g. filepath) in storage.
	WriteBytes(ctx context.Context, key string, data []byte) (int, error)

	// WriteStream writes the supplied data stream at key (e.g. filepath) in storage.
	WriteStream(ctx context.Context, key string, stream io.Reader) (int64, error)

	// Stat returns details about key (e.g. filepath) in storage, nil indicates not found.
	Stat(ctx context.Context, key string) (*Entry, error)

	// Remove will remove data at key from storage.
	Remove(ctx context.Context, key string) error

	// Clean in simple terms performs a clean of underlying
	// storage mechanism. For memory implementations this may
	// compact the underlying hashmap, for disk filesystems
	// this may remove now-unused directories.
	Clean(ctx context.Context) error

	// WalkKeys walks available keys using opts in storage.
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
