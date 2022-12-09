package ttl

import (
	"sync"
	"time"

	"codeberg.org/gruf/go-maps"
)

// Entry represents an item in the cache, with it's currently calculated Expiry time.
type Entry[Key comparable, Value any] struct {
	Key    Key
	Value  Value
	Expiry time.Time
}

// Cache is the underlying Cache implementation, providing both the base Cache interface and unsafe access to underlying map to allow flexibility in building your own.
type Cache[Key comparable, Value any] struct {
	// TTL is the cache item TTL.
	TTL time.Duration

	// Evict is the hook that is called when an item is evicted from the cache, includes manual delete.
	Evict func(*Entry[Key, Value])

	// Invalid is the hook that is called when an item's data in the cache is invalidated.
	Invalid func(*Entry[Key, Value])

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

	if ok = c.stop == nil; ok {
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

	if ok = c.stop != nil; ok {
		// We're running, cancel evicts
		c.stop()
		c.stop = nil
	}

	// Done with lock
	c.Unlock()

	return
}

// Sweep attempts to evict expired items (with callback!) from cache.
func (c *Cache[K, V]) Sweep(now time.Time) {
	var after int

	// Sweep within lock
	c.Lock()
	defer c.Unlock()

	// Sentinel value
	after = -1

	// The cache will be ordered by expiry date, we iterate until we reach the index of
	// the youngest item that hsa expired, as all succeeding items will also be expired.
	c.Cache.RangeIf(0, c.Cache.Len(), func(i int, _ K, item *Entry[K, V]) bool {
		if now.After(item.Expiry) {
			after = i

			// All older than this (including) can be dropped
			return false
		}

		// Continue looping
		return true
	})

	if after == -1 {
		// No Truncation needed
		return
	}

	// Truncate items, calling eviction hook
	c.truncate(c.Cache.Len()-after, c.Evict)
}

// SetEvictionCallback: implements cache.Cache's SetEvictionCallback().
func (c *Cache[K, V]) SetEvictionCallback(hook func(*Entry[K, V])) {
	// Ensure non-nil hook
	if hook == nil {
		hook = func(*Entry[K, V]) {}
	}

	// Update within lock
	c.Lock()
	defer c.Unlock()

	// Update hook
	c.Evict = hook
}

// SetInvalidateCallback: implements cache.Cache's SetInvalidateCallback().
func (c *Cache[K, V]) SetInvalidateCallback(hook func(*Entry[K, V])) {
	// Ensure non-nil hook
	if hook == nil {
		hook = func(*Entry[K, V]) {}
	}

	// Update within lock
	c.Lock()
	defer c.Unlock()

	// Update hook
	c.Invalid = hook
}

// SetTTL: implements cache.Cache's SetTTL().
func (c *Cache[K, V]) SetTTL(ttl time.Duration, update bool) {
	if ttl < 0 {
		panic("ttl must be greater than zero")
	}

	// Update within lock
	c.Lock()
	defer c.Unlock()

	// Set updated TTL
	diff := ttl - c.TTL
	c.TTL = ttl

	if update {
		// Update existing cache entries with new expiry time
		c.Cache.Range(0, c.Cache.Len(), func(i int, key K, item *Entry[K, V]) {
			item.Expiry = item.Expiry.Add(diff)
		})
	}
}

// Get: implements cache.Cache's Get().
func (c *Cache[K, V]) Get(key K) (V, bool) {
	// Read within lock
	c.Lock()
	defer c.Unlock()

	// Check for item in cache
	item, ok := c.Cache.Get(key)
	if !ok {
		var value V
		return value, false
	}

	// Update item expiry and return
	item.Expiry = time.Now().Add(c.TTL)
	return item.Value, true
}

// Add: implements cache.Cache's Add().
func (c *Cache[K, V]) Add(key K, value V) bool {
	// Write within lock
	c.Lock()
	defer c.Unlock()

	// If already cached, return
	if c.Cache.Has(key) {
		return false
	}

	// Alloc new item
	item := c.alloc()
	item.Key = key
	item.Value = value
	item.Expiry = time.Now().Add(c.TTL)

	var hook func(K, *Entry[K, V])

	if c.Evict != nil {
		// Pass evicted entry to user hook
		hook = func(_ K, item *Entry[K, V]) {
			c.Evict(item)
		}
	}

	// Place new item in the map with hook
	c.Cache.SetWithHook(key, item, hook)

	return true
}

// Set: implements cache.Cache's Set().
func (c *Cache[K, V]) Set(key K, value V) {
	// Write within lock
	c.Lock()
	defer c.Unlock()

	// Check if already exists
	item, ok := c.Cache.Get(key)

	if ok {
		if c.Invalid != nil {
			// Invalidate existing
			c.Invalid(item)
		}
	} else {
		// Allocate new item
		item = c.alloc()
		item.Key = key
		c.Cache.Set(key, item)
	}

	// Update the item value + expiry
	item.Expiry = time.Now().Add(c.TTL)
	item.Value = value
}

// CAS: implements cache.Cache's CAS().
func (c *Cache[K, V]) CAS(key K, old V, new V, cmp func(V, V) bool) bool {
	// CAS within lock
	c.Lock()
	defer c.Unlock()

	// Check for item in cache
	item, ok := c.Cache.Get(key)
	if !ok || !cmp(item.Value, old) {
		return false
	}

	if c.Invalid != nil {
		// Invalidate item
		c.Invalid(item)
	}

	// Update item + Expiry
	item.Value = new
	item.Expiry = time.Now().Add(c.TTL)

	return ok
}

// Swap: implements cache.Cache's Swap().
func (c *Cache[K, V]) Swap(key K, swp V) V {
	// Swap within lock
	c.Lock()
	defer c.Unlock()

	// Check for item in cache
	item, ok := c.Cache.Get(key)
	if !ok {
		var value V
		return value
	}

	if c.Invalid != nil {
		// invalidate old
		c.Invalid(item)
	}

	old := item.Value

	// update item + Expiry
	item.Value = swp
	item.Expiry = time.Now().Add(c.TTL)

	return old
}

// Has: implements cache.Cache's Has().
func (c *Cache[K, V]) Has(key K) bool {
	c.Lock()
	ok := c.Cache.Has(key)
	c.Unlock()
	return ok
}

// Invalidate: implements cache.Cache's Invalidate().
func (c *Cache[K, V]) Invalidate(key K) bool {
	// Delete within lock
	c.Lock()
	defer c.Unlock()

	// Check if we have item with key
	item, ok := c.Cache.Get(key)
	if !ok {
		return false
	}

	// Remove from cache map
	_ = c.Cache.Delete(key)

	if c.Invalid != nil {
		// Invalidate item
		c.Invalid(item)
	}

	// Return item to pool
	c.free(item)

	return true
}

// Clear: implements cache.Cache's Clear().
func (c *Cache[K, V]) Clear() {
	c.Lock()
	defer c.Unlock()
	c.truncate(c.Cache.Len(), c.Invalid)
}

// Len: implements cache.Cache's Len().
func (c *Cache[K, V]) Len() int {
	c.Lock()
	l := c.Cache.Len()
	c.Unlock()
	return l
}

// Cap: implements cache.Cache's Cap().
func (c *Cache[K, V]) Cap() int {
	c.Lock()
	l := c.Cache.Cap()
	c.Unlock()
	return l
}

// truncate will call Cache.Truncate(sz), and if provided a hook will temporarily store deleted items before passing them to the hook. This is required in order to prevent cache writes during .Truncate().
func (c *Cache[K, V]) truncate(sz int, hook func(*Entry[K, V])) {
	if hook == nil {
		// No hook was provided, we can simply truncate and free items immediately.
		c.Cache.Truncate(sz, func(_ K, item *Entry[K, V]) { c.free(item) })
		return
	}

	// Store list of deleted items for later callbacks
	deleted := make([]*Entry[K, V], 0, sz)

	// Truncate and store list of deleted items
	c.Cache.Truncate(sz, func(_ K, item *Entry[K, V]) {
		deleted = append(deleted, item)
	})

	// Pass each deleted to hook, then free
	for _, item := range deleted {
		hook(item)
		c.free(item)
	}
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

// free will reset entry fields and place back in pool.
func (c *Cache[K, V]) free(e *Entry[K, V]) {
	var (
		zk K
		zv V
	)
	e.Key = zk
	e.Value = zv
	e.Expiry = time.Time{}
	c.pool = append(c.pool, e)
}
