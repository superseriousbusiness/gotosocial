package storage

import (
	"errors"
	"io"
	"net/url"
)

var (
	ErrNotSupported = errors.New("driver does not suppport functionality")
)

// Driver implements the functionality to store and retrieve blobs
// (images,video,audio)
type Driver interface {
	Get(key string) ([]byte, error)
	GetStream(key string) (io.ReadCloser, error)
	PutStream(key string, r io.Reader) error
	Put(key string, value []byte) error
	Delete(key string) error
	URL(key string) *url.URL
}
