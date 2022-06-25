package storage

import (
	"io"

	"codeberg.org/gruf/go-store/kv"
)

// Driver implements the functionality to store and retrieve blobs
// (images,video,audio)
type Driver interface {
	Get(key string) ([]byte, error)
	GetStream(key string) (io.ReadCloser, error)
	PutStream(key string, r io.Reader) error
	Put(key string, value []byte) error
	Delete(key string) error
	Iterator(matchFn func(string) bool) (*kv.KVIterator, error)
}
