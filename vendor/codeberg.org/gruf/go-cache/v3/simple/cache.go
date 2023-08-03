package simple

import (
	"sync"

	"codeberg.org/gruf/go-maps"
)

// Entry represents an item in the cache.
type Entry struct {
	Key   any
	Value any
}

// Cache is the underlying Cache implementation, providing both the base Cache interface and unsafe access to underlying map to allow flexibility in building your own.
type Cache[Key comparable, Value any] struct {
	// Evict is the hook that is called when an item is evicted from the cache.
	Evict func(Key, Value)

	// Invalid is the hook that is called when an item's data in the cache is invalidated, includes Add/Set.
	Invalid func(Key, Value)

	// Cache is the underlying hashmap used for this cache.
	Cache maps.LRUMap[Key, *Entry]

	// Embedded mutex.
	sync.Mutex
}

// New returns a new initialized Cache with given initial length, maximum capacity and item TTL.
func New[K comparable, V any](len, cap int) *Cache[K, V] {
	c := new(Cache[K, V])
	c.Init(len, cap)
	return c
}

// Init will initialize this cache with given initial length, maximum capacity and item TTL.
func (c *Cache[K, V]) Init(len, cap int) {
	c.SetEvictionCallback(nil)
	c.SetInvalidateCallback(nil)
	c.Cache.Init(len, cap)
}

// SetEvictionCallback: implements cache.Cache's SetEvictionCallback().
func (c *Cache[K, V]) SetEvictionCallback(hook func(K, V)) {
	c.locked(func() { c.Evict = hook })
}

// SetInvalidateCallback: implements cache.Cache's SetInvalidateCallback().
func (c *Cache[K, V]) SetInvalidateCallback(hook func(K, V)) {
	c.locked(func() { c.Invalid = hook })
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
		var item *Entry

		// Check for item in cache
		item, ok = c.Cache.Get(key)
		if !ok {
			return
		}

		// Set item value.
		v = item.Value.(V)
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
		new := getEntry()
		new.Key = key
		new.Value = value

		// Add new entry to cache and catched any evicted item.
		c.Cache.SetWithHook(key, new, func(_ K, item *Entry) {
			evcK = item.Key.(K)
			evcV = item.Value.(V)
			ev = true
			putEntry(item)
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
		var item *Entry

		// Check for item in cache
		item, ok = c.Cache.Get(key)

		if ok {
			// Set old value.
			oldV = item.Value.(V)

			// Update the existing item.
			item.Value = value
		} else {
			// Alloc new entry.
			new := getEntry()
			new.Key = key
			new.Value = value

			// Add new entry to cache and catched any evicted item.
			c.Cache.SetWithHook(key, new, func(_ K, item *Entry) {
				evcK = item.Key.(K)
				evcV = item.Value.(V)
				ev = true
				putEntry(item)
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
		var item *Entry

		// Check for item in cache
		item, ok = c.Cache.Get(key)
		if !ok {
			return
		}

		// Set old value.
		oldV = item.Value.(V)

		// Perform the comparison
		if !cmp(old, oldV) {
			var zero V
			oldV = zero
			return
		}

		// Update value.
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
		var item *Entry

		// Check for item in cache
		item, ok = c.Cache.Get(key)
		if !ok {
			return
		}

		// Set old value.
		oldV = item.Value.(V)

		// Update value.
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
		var item *Entry

		// Check for item in cache
		item, ok = c.Cache.Get(key)
		if !ok {
			return
		}

		// Set old value.
		oldV = item.Value.(V)

		// Remove from cache map
		_ = c.Cache.Delete(key)

		// Free entry
		putEntry(item)

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
		// deleted items.
		items []*Entry

		// hook func ptrs.
		invalid func(K, V)
	)

	// Allocate a slice for invalidated.
	items = make([]*Entry, 0, len(keys))

	c.locked(func() {
		for x := range keys {
			var item *Entry

			// Check for item in cache
			item, ok = c.Cache.Get(keys[x])
			if !ok {
				continue
			}

			// Append this old value.
			items = append(items, item)

			// Remove from cache map
			_ = c.Cache.Delete(keys[x])
		}

		// Set hook func ptrs.
		invalid = c.Invalid
	})

	if invalid != nil {
		for x := range items {
			// Pass to invalidate hook.
			k := items[x].Key.(K)
			v := items[x].Value.(V)
			invalid(k, v)

			// Free this entry.
			putEntry(items[x])
		}
	}

	return
}

// Clear: implements cache.Cache's Clear().
func (c *Cache[K, V]) Clear() { c.Trim(100) }

// Trim will truncate the cache to ensure it stays within given percentage of total capacity.
func (c *Cache[K, V]) Trim(perc float64) {
	var (
		// deleted items
		items []*Entry

		// hook func ptrs.
		invalid func(K, V)
	)

	c.locked(func() {
		// Calculate number of cache items to truncate.
		max := (perc / 100) * float64(c.Cache.Cap())
		diff := c.Cache.Len() - int(max)
		if diff <= 0 {
			return
		}

		// Set hook func ptr.
		invalid = c.Invalid

		// Truncate by calculated length.
		items = c.truncate(diff, invalid)
	})

	if invalid != nil {
		for x := range items {
			// Pass to invalidate hook.
			k := items[x].Key.(K)
			v := items[x].Value.(V)
			invalid(k, v)

			// Free this entry.
			putEntry(items[x])
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

// locked performs given function within mutex lock (NOTE: UNLOCK IS NOT DEFERRED).
func (c *Cache[K, V]) locked(fn func()) {
	c.Lock()
	fn()
	c.Unlock()
}

// truncate will truncate the cache by given size, returning deleted items.
func (c *Cache[K, V]) truncate(sz int, hook func(K, V)) []*Entry {
	if hook == nil {
		// No hook to execute, simply release all truncated entries.
		c.Cache.Truncate(sz, func(_ K, item *Entry) { putEntry(item) })
		return nil
	}

	// Allocate a slice for deleted.
	deleted := make([]*Entry, 0, sz)

	// Truncate and catch all deleted k-v pairs.
	c.Cache.Truncate(sz, func(_ K, item *Entry) {
		deleted = append(deleted, item)
	})

	return deleted
}
