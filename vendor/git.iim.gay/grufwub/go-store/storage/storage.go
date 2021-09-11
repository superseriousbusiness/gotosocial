package storage

import (
	"io"
)

// StorageEntry defines a key in Storage
type StorageEntry interface {
	// Key returns the storage entry's key
	Key() string
}

// entry is the simplest possible StorageEntry
type entry string

func (e entry) Key() string {
	return string(e)
}

// Storage defines a means of storing and accessing key value pairs
type Storage interface {
	// Clean removes unused values and unclutters the storage (e.g. removing empty folders)
	Clean() error

	// ReadBytes returns the byte value for key in storage
	ReadBytes(key string) ([]byte, error)

	// ReadStream returns an io.ReadCloser for the value bytes at key in the storage
	ReadStream(key string) (io.ReadCloser, error)

	// WriteBytes writes the supplied value bytes at key in the storage
	WriteBytes(key string, value []byte) error

	// WriteStream writes the bytes from supplied reader at key in the storage
	WriteStream(key string, r io.Reader) error

	// Stat checks if the supplied key is in the storage
	Stat(key string) (bool, error)

	// Remove attempts to remove the supplied key-value pair from storage
	Remove(key string) error

	// WalkKeys walks the keys in the storage
	WalkKeys(opts *WalkKeysOptions) error
}

// WalkKeysOptions defines how to walk the keys in a storage implementation
type WalkKeysOptions struct {
	// WalkFn is the function to apply on each StorageEntry
	WalkFn func(StorageEntry)
}
