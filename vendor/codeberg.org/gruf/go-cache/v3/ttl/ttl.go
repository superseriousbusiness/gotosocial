package ttl

import (
	"sync"
	"time"
	_ "unsafe"

	"codeberg.org/gruf/go-maps"
)

// Entry represents an item in the cache, with it's currently calculated Expiry time.
type Entry[Key comparable, Value any] struct {
	Key    Key
	Value  Value
	Expiry uint64
}

// Cache is the underlying Cache implementation, providing both the base Cache interface and unsafe access to underlying map to allow flexibility in building your own.
type Cache[Key comparable, Value any] struct {
	// TTL is the cache item TTL.
	TTL time.Duration

	// Evict is the hook that is called when an item is evicted from the cache.
	Evict func(Key, Value)

	// Invalid is the hook that is called when an item's data in the cache is invalidated, includes Add/Set.
	Invalid func(Key, Value)

	// Cache is the underlying hashmap used for this cache.
	Cache maps.LRUMap[Key, *Entry[Key, Value]]

	// stop is the eviction routine cancel func.
	stop func()

	// pool is a memory pool of entry objects.
	pool []*Entry[Key, Value]

	// Embedded mutex.
	sync.Mutex
}

// New returns a new initialized Cache with given initial length, maximum capacity and item TTL.
func New[K comparable, V any](len, cap int, ttl time.Duration) *Cache[K, V] {
	c := new(Cache[K, V])
	c.Init(len, cap, ttl)
	return c
}

// Init will initialize this cache with given initial length, maximum capacity and item TTL.
func (c *Cache[K, V]) Init(len, cap int, ttl time.Duration) {
	if ttl <= 0 {
		// Default duration
		ttl = time.Second * 5
	}
	c.TTL = ttl
	c.SetEvictionCallback(nil)
	c.SetInvalidateCallback(nil)
	c.Cache.Init(len, cap)
}

// Start: implements cache.Cache's Start().
func (c *Cache[K, V]) Start(freq time.Duration) (ok bool) {
	// Nothing to start
	if freq <= 0 {
		return false
	}

	// Safely start
	c.Lock()

	if ok = (c.stop == nil); ok {
		// Not yet running, schedule us
		c.stop = schedule(c.Sweep, freq)
	}

	// Done with lock
	c.Unlock()

	return
}

// Stop: implements cache.Cache's Stop().
func (c *Cache[K, V]) Stop() (ok bool) {
	// Safely stop
	c.Lock()

	if ok = (c.stop != nil); ok {
		// We're running, cancel evicts
		c.stop()
		c.stop = nil
	}

	// Done with lock
	c.Unlock()

	return
}

// Sweep attempts to evict expired items (with callback!) from cache.
func (c *Cache[K, V]) Sweep(_ time.Time) {
	var (
		// evicted key-values.
		kvs []kv[K, V]

		// hook func ptrs.
		evict func(K, V)

		// get current nanoseconds.
		now = runtime_nanotime()
	)

	c.locked(func() {
		if c.TTL <= 0 {
			// sweep is
			// disabled
			return
		}

		// Sentinel value
		after := -1

		// The cache will be ordered by expiry date, we iterate until we reach the index of
		// the youngest item that hsa expired, as all succeeding items will also be expired.
		c.Cache.RangeIf(0, c.Cache.Len(), func(i int, _ K, item *Entry[K, V]) bool {
			if now > item.Expiry {
				after = i

				// evict all older items
				// than this (inclusive)
				return false
			}

			// cont. loop.
			return true
		})

		if after == -1 {
			// No Truncation needed
			return
		}

		// Set hook func ptr.
		evict = c.Evict

		// Truncate determined size.
		sz := c.Cache.Len() - after
		kvs = c.truncate(sz, evict)
	})

	if evict != nil {
		for x := range kvs {
			// Pass to eviction hook.
			evict(kvs[x].K, kvs[x].V)
		}
	}
}

// SetEvictionCallback: implements cache.Cache's SetEvictionCallback().
func (c *Cache[K, V]) SetEvictionCallback(hook func(K, V)) {
	c.locked(func() {
		c.Evict = hook
	})
}

// SetInvalidateCallback: implements cache.Cache's SetInvalidateCallback().
func (c *Cache[K, V]) SetInvalidateCallback(hook func(K, V)) {
	c.locked(func() {
		c.Invalid = hook
	})
}

// SetTTL: implements cache.Cache's SetTTL().
func (c *Cache[K, V]) SetTTL(ttl time.Duration, update bool) {
	c.locked(func() {
		// Set updated TTL
		diff := ttl - c.TTL
		c.TTL = ttl

		if update {
			// Update existing cache entries with new expiry time
			c.Cache.Range(0, c.Cache.Len(), func(i int, _ K, item *Entry[K, V]) {
				item.Expiry += uint64(diff)
			})
		}
	})
}

// Get: implements cache.Cache's Get().
func (c *Cache[K, V]) Get(key K) (V, bool) {
	var (
		// did exist in cache?
		ok bool

		// cached value.
		v V
	)

	c.locked(func() {
		var item *Entry[K, V]

		// Check for item in cache
		item, ok = c.Cache.Get(key)
		if !ok {
			return
		}

		// Update fetched's expiry
		item.Expiry = c.expiry()

		// Set value.
		v = item.Value
	})

	return v, ok
}

// Add: implements cache.Cache's Add().
func (c *Cache[K, V]) Add(key K, value V) bool {
	var (
		// did exist in cache?
		ok bool

		// was entry evicted?
		ev bool

		// evicted key values.
		evcK K
		evcV V

		// hook func ptrs.
		evict func(K, V)
	)

	c.locked(func() {
		// Check if in cache.
		ok = c.Cache.Has(key)
		if ok {
			return
		}

		// Alloc new entry.
		new := c.alloc()
		new.Expiry = c.expiry()
		new.Key = key
		new.Value = value

		// Add new entry to cache and catched any evicted item.
		c.Cache.SetWithHook(key, new, func(_ K, item *Entry[K, V]) {
			evcK = item.Key
			evcV = item.Value
			ev = true
			c.free(item)
		})

		// Set hook func ptr.
		evict = c.Evict
	})

	if ev && evict != nil {
		// Pass to eviction hook.
		evict(evcK, evcV)
	}

	return !ok
}

// Set: implements cache.Cache's Set().
func (c *Cache[K, V]) Set(key K, value V) {
	var (
		// did exist in cache?
		ok bool

		// was entry evicted?
		ev bool

		// old value.
		oldV V

		// evicted key values.
		evcK K
		evcV V

		// hook func ptrs.
		invalid func(K, V)
		evict   func(K, V)
	)

	c.locked(func() {
		var item *Entry[K, V]

		// Check for item in cache
		item, ok = c.Cache.Get(key)

		if ok {
			// Set old value.
			oldV = item.Value

			// Update the existing item.
			item.Expiry = c.expiry()
			item.Value = value
		} else {
			// Alloc new entry.
			new := c.alloc()
			new.Expiry = c.expiry()
			new.Key = key
			new.Value = value

			// Add new entry to cache and catched any evicted item.
			c.Cache.SetWithHook(key, new, func(_ K, item *Entry[K, V]) {
				evcK = item.Key
				evcV = item.Value
				ev = true
				c.free(item)
			})
		}

		// Set hook func ptrs.
		invalid = c.Invalid
		evict = c.Evict
	})

	if ok && invalid != nil {
		// Pass to invalidate hook.
		invalid(key, oldV)
	}

	if ev && evict != nil {
		// Pass to eviction hook.
		evict(evcK, evcV)
	}
}

// CAS: implements cache.Cache's CAS().
func (c *Cache[K, V]) CAS(key K, old V, new V, cmp func(V, V) bool) bool {
	var (
		// did exist in cache?
		ok bool

		// swapped value.
		oldV V

		// hook func ptrs.
		invalid func(K, V)
	)

	c.locked(func() {
		var item *Entry[K, V]

		// Check for item in cache
		item, ok = c.Cache.Get(key)
		if !ok {
			return
		}

		// Perform the comparison
		if !cmp(old, item.Value) {
			return
		}

		// Set old value.
		oldV = item.Value

		// Update value + expiry.
		item.Expiry = c.expiry()
		item.Value = new

		// Set hook func ptr.
		invalid = c.Invalid
	})

	if ok && invalid != nil {
		// Pass to invalidate hook.
		invalid(key, oldV)
	}

	return ok
}

// Swap: implements cache.Cache's Swap().
func (c *Cache[K, V]) Swap(key K, swp V) V {
	var (
		// did exist in cache?
		ok bool

		// swapped value.
		oldV V

		// hook func ptrs.
		invalid func(K, V)
	)

	c.locked(func() {
		var item *Entry[K, V]

		// Check for item in cache
		item, ok = c.Cache.Get(key)
		if !ok {
			return
		}

		// Set old value.
		oldV = item.Value

		// Update value + expiry.
		item.Expiry = c.expiry()
		item.Value = swp

		// Set hook func ptr.
		invalid = c.Invalid
	})

	if ok && invalid != nil {
		// Pass to invalidate hook.
		invalid(key, oldV)
	}

	return oldV
}

// Has: implements cache.Cache's Has().
func (c *Cache[K, V]) Has(key K) (ok bool) {
	c.locked(func() {
		ok = c.Cache.Has(key)
	})
	return
}

// Invalidate: implements cache.Cache's Invalidate().
func (c *Cache[K, V]) Invalidate(key K) (ok bool) {
	var (
		// old value.
		oldV V

		// hook func ptrs.
		invalid func(K, V)
	)

	c.locked(func() {
		var item *Entry[K, V]

		// Check for item in cache
		item, ok = c.Cache.Get(key)
		if !ok {
			return
		}

		// Set old value.
		oldV = item.Value

		// Remove from cache map
		_ = c.Cache.Delete(key)

		// Free entry
		c.free(item)

		// Set hook func ptrs.
		invalid = c.Invalid
	})

	if ok && invalid != nil {
		// Pass to invalidate hook.
		invalid(key, oldV)
	}

	return
}

// InvalidateAll: implements cache.Cache's InvalidateAll().
func (c *Cache[K, V]) InvalidateAll(keys ...K) (ok bool) {
	var (
		// invalidated kvs.
		kvs []kv[K, V]

		// hook func ptrs.
		invalid func(K, V)
	)

	// Allocate a slice for invalidated.
	kvs = make([]kv[K, V], 0, len(keys))

	c.locked(func() {
		for _, key := range keys {
			var item *Entry[K, V]

			// Check for item in cache
			item, ok = c.Cache.Get(key)
			if !ok {
				return
			}

			// Append this old value to slice
			kvs = append(kvs, kv[K, V]{
				K: key,
				V: item.Value,
			})

			// Remove from cache map
			_ = c.Cache.Delete(key)

			// Free entry
			c.free(item)
		}

		// Set hook func ptrs.
		invalid = c.Invalid
	})

	if invalid != nil {
		for x := range kvs {
			// Pass to invalidate hook.
			invalid(kvs[x].K, kvs[x].V)
		}
	}

	return
}

// Clear: implements cache.Cache's Clear().
func (c *Cache[K, V]) Clear() {
	var (
		// deleted key-values.
		kvs []kv[K, V]

		// hook func ptrs.
		invalid func(K, V)
	)

	c.locked(func() {
		// Set hook func ptr.
		invalid = c.Invalid

		// Truncate the entire cache length.
		kvs = c.truncate(c.Cache.Len(), invalid)
	})

	if invalid != nil {
		for x := range kvs {
			// Pass to invalidate hook.
			invalid(kvs[x].K, kvs[x].V)
		}
	}
}

// Len: implements cache.Cache's Len().
func (c *Cache[K, V]) Len() (l int) {
	c.locked(func() { l = c.Cache.Len() })
	return
}

// Cap: implements cache.Cache's Cap().
func (c *Cache[K, V]) Cap() (l int) {
	c.locked(func() { l = c.Cache.Cap() })
	return
}

func (c *Cache[K, V]) locked(fn func()) {
	c.Lock()
	fn()
	c.Unlock()
}

// truncate will truncate the cache by given size, returning deleted items.
func (c *Cache[K, V]) truncate(sz int, hook func(K, V)) []kv[K, V] {
	if hook == nil {
		// No hook to execute, simply free all truncated entries.
		c.Cache.Truncate(sz, func(_ K, e *Entry[K, V]) { c.free(e) })
		return nil
	}

	// Allocate a slice for deleted k-v pairs.
	deleted := make([]kv[K, V], 0, sz)

	c.Cache.Truncate(sz, func(_ K, item *Entry[K, V]) {
		// Store key-value pair for later access.
		deleted = append(deleted, kv[K, V]{
			K: item.Key,
			V: item.Value,
		})

		// Free entry.
		c.free(item)
	})

	return deleted
}

// alloc will acquire cache entry from pool, or allocate new.
func (c *Cache[K, V]) alloc() *Entry[K, V] {
	if len(c.pool) == 0 {
		return &Entry[K, V]{}
	}
	idx := len(c.pool) - 1
	e := c.pool[idx]
	c.pool = c.pool[:idx]
	return e
}

// clone allocates a new Entry and copies all info from passed Entry.
func (c *Cache[K, V]) clone(e *Entry[K, V]) *Entry[K, V] {
	e2 := c.alloc()
	e2.Key = e.Key
	e2.Value = e.Value
	e2.Expiry = e.Expiry
	return e2
}

// free will reset entry fields and place back in pool.
func (c *Cache[K, V]) free(e *Entry[K, V]) {
	var (
		zk K
		zv V
	)
	e.Expiry = 0
	e.Key = zk
	e.Value = zv
	c.pool = append(c.pool, e)
}

//go:linkname runtime_nanotime runtime.nanotime
func runtime_nanotime() uint64

// expiry returns an the next expiry time to use for an entry,
// which is equivalent to time.Now().Add(ttl), or zero if disabled.
func (c *Cache[K, V]) expiry() uint64 {
	if ttl := c.TTL; ttl > 0 {
		return runtime_nanotime() +
			uint64(c.TTL)
	}
	return 0
}

type kv[K comparable, V any] struct {
	K K
	V V
}
