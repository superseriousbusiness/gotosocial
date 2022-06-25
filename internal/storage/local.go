package storage

import "codeberg.org/gruf/go-store/kv"

type Local struct {
	*kv.KVStore
}
