package mangler

import (
	"sync/atomic"
	"unsafe"

	"codeberg.org/gruf/go-xunsafe"
)

var manglers cache

// cache is a concurrency-safe map[xunsafe.TypeInfo]Mangler
// cache, designed for heavy reads but with unfortunately expensive
// writes. it is designed such that after some initial load period
// in which functions are cached by types, all future ops are reads.
type cache struct{ p unsafe.Pointer }

// Get will check cache for mangler func under key.
func (c *cache) Get(t xunsafe.TypeInfo) Mangler {
	if p := c.load(); p != nil {
		return (*p)[t]
	}
	return nil
}

// Put will place given mangler func in cache under key, if not already exists.
func (c *cache) Put(t xunsafe.TypeInfo, fn Mangler) {
	for {
		p := c.load()

		var cache map[xunsafe.TypeInfo]Mangler

		if p != nil {
			if _, ok := (*p)[t]; ok {
				return
			}

			cache = make(map[xunsafe.TypeInfo]Mangler, len(*p)+1)
			for key, value := range *p {
				cache[key] = value
			}
		} else {
			cache = make(map[xunsafe.TypeInfo]Mangler, 1)
		}

		cache[t] = fn

		if c.cas(p, &cache) {
			return
		}
	}
}

// load is a typed wrapper around atomic.LoadPointer().
func (c *cache) load() *map[xunsafe.TypeInfo]Mangler {
	return (*map[xunsafe.TypeInfo]Mangler)(atomic.LoadPointer(&c.p))
}

// cas is a typed wrapper around atomic.CompareAndSwapPointer().
func (c *cache) cas(old, new *map[xunsafe.TypeInfo]Mangler) bool {
	return atomic.CompareAndSwapPointer(&c.p, unsafe.Pointer(old), unsafe.Pointer(new))
}
