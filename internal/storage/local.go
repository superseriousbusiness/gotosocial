package storage

import (
	"net/url"

	"codeberg.org/gruf/go-store/kv"
)

type Local struct {
	*kv.KVStore
}

func (l *Local) URL(key string) *url.URL {
	return nil
}
