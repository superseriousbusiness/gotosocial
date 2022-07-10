package cache

import (
	"sync"
	"time"
)

// TTLCache is the underlying Cache implementation, providing both the base
// Cache interface and access to "unsafe" methods so that you may build your
// customized caches ontop of this structure.
type TTLCache[Key comparable, Value any] struct {
	cache   map[Key](*entry[Value])
	evict   Hook[Key, Value] // the evict hook is called when an item is evicted from the cache, includes manual delete
	invalid Hook[Key, Value] // the invalidate hook is called when an item's data in the cache is invalidated
	ttl     time.Duration    // ttl is the item TTL
	stop    func()           // stop is the cancel function for the scheduled eviction routine
	mu      sync.Mutex       // mu protects TTLCache for concurrent access
}

// Init performs Cache initialization. MUST be called.
func (c *TTLCache[K, V]) Init() {
	c.cache = make(map[K](*entry[V]), 100)
	c.evict = emptyHook[K, V]
	c.invalid = emptyHook[K, V]
	c.ttl = time.Minute * 5
}

func (c *TTLCache[K, V]) Start(freq time.Duration) (ok bool) {
	// Nothing to start
	if freq <= 0 {
		return false
	}

	// Safely start
	c.mu.Lock()

	if ok = c.stop == nil; ok {
		// Not yet running, schedule us
		c.stop = schedule(c.sweep, freq)
	}

	// Done with lock
	c.mu.Unlock()

	return
}

func (c *TTLCache[K, V]) Stop() (ok bool) {
	// Safely stop
	c.mu.Lock()

	if ok = c.stop != nil; ok {
		// We're running, cancel evicts
		c.stop()
		c.stop = nil
	}

	// Done with lock
	c.mu.Unlock()

	return
}

// sweep attempts to evict expired items (with callback!) from cache.
func (c *TTLCache[K, V]) sweep(now time.Time) {
	// Lock and defer unlock (in case of hook panic)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Sweep the cache for old items!
	for key, item := range c.cache {
		if now.After(item.expiry) {
			c.evict(key, item.value)
			delete(c.cache, key)
		}
	}
}

// Lock locks the cache mutex.
func (c *TTLCache[K, V]) Lock() {
	c.mu.Lock()
}

// Unlock unlocks the cache mutex.
func (c *TTLCache[K, V]) Unlock() {
	c.mu.Unlock()
}

func (c *TTLCache[K, V]) SetEvictionCallback(hook Hook[K, V]) {
	// Ensure non-nil hook
	if hook == nil {
		hook = emptyHook[K, V]
	}

	// Safely set evict hook
	c.mu.Lock()
	c.evict = hook
	c.mu.Unlock()
}

func (c *TTLCache[K, V]) SetInvalidateCallback(hook Hook[K, V]) {
	// Ensure non-nil hook
	if hook == nil {
		hook = emptyHook[K, V]
	}

	// Safely set invalidate hook
	c.mu.Lock()
	c.invalid = hook
	c.mu.Unlock()
}

func (c *TTLCache[K, V]) SetTTL(ttl time.Duration, update bool) {
	// Safely update TTL
	c.mu.Lock()
	diff := ttl - c.ttl
	c.ttl = ttl

	if update {
		// Update existing cache entries
		for _, entry := range c.cache {
			entry.expiry.Add(diff)
		}
	}

	// We're done
	c.mu.Unlock()
}

func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	value, ok := c.GetUnsafe(key)
	c.mu.Unlock()
	return value, ok
}

// GetUnsafe is the mutex-unprotected logic for Cache.Get().
func (c *TTLCache[K, V]) GetUnsafe(key K) (V, bool) {
	item, ok := c.cache[key]
	if !ok {
		var value V
		return value, false
	}
	item.expiry = time.Now().Add(c.ttl)
	return item.value, true
}

func (c *TTLCache[K, V]) Put(key K, value V) bool {
	c.mu.Lock()
	success := c.PutUnsafe(key, value)
	c.mu.Unlock()
	return success
}

// PutUnsafe is the mutex-unprotected logic for Cache.Put().
func (c *TTLCache[K, V]) PutUnsafe(key K, value V) bool {
	// If already cached, return
	if _, ok := c.cache[key]; ok {
		return false
	}

	// Create new cached item
	c.cache[key] = &entry[V]{
		value:  value,
		expiry: time.Now().Add(c.ttl),
	}

	return true
}

func (c *TTLCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock() // defer in case of hook panic
	c.SetUnsafe(key, value)
}

// SetUnsafe is the mutex-unprotected logic for Cache.Set(), it calls externally-set functions.
func (c *TTLCache[K, V]) SetUnsafe(key K, value V) {
	item, ok := c.cache[key]
	if ok {
		// call invalidate hook
		c.invalid(key, item.value)
	} else {
		// alloc new item
		item = &entry[V]{}
		c.cache[key] = item
	}

	// Update the item + expiry
	item.value = value
	item.expiry = time.Now().Add(c.ttl)
}

func (c *TTLCache[K, V]) CAS(key K, cmp V, swp V) bool {
	c.mu.Lock()
	ok := c.CASUnsafe(key, cmp, swp)
	c.mu.Unlock()
	return ok
}

// CASUnsafe is the mutex-unprotected logic for Cache.CAS().
func (c *TTLCache[K, V]) CASUnsafe(key K, cmp V, swp V) bool {
	// Check for item
	item, ok := c.cache[key]
	if !ok || !Compare(item.value, cmp) {
		return false
	}

	// Invalidate item
	c.invalid(key, item.value)

	// Update item + expiry
	item.value = swp
	item.expiry = time.Now().Add(c.ttl)

	return ok
}

func (c *TTLCache[K, V]) Swap(key K, swp V) V {
	c.mu.Lock()
	old := c.SwapUnsafe(key, swp)
	c.mu.Unlock()
	return old
}

// SwapUnsafe is the mutex-unprotected logic for Cache.Swap().
func (c *TTLCache[K, V]) SwapUnsafe(key K, swp V) V {
	// Check for item
	item, ok := c.cache[key]
	if !ok {
		var value V
		return value
	}

	// invalidate old item
	c.invalid(key, item.value)
	old := item.value

	// update item + expiry
	item.value = swp
	item.expiry = time.Now().Add(c.ttl)

	return old
}

func (c *TTLCache[K, V]) Has(key K) bool {
	c.mu.Lock()
	ok := c.HasUnsafe(key)
	c.mu.Unlock()
	return ok
}

// HasUnsafe is the mutex-unprotected logic for Cache.Has().
func (c *TTLCache[K, V]) HasUnsafe(key K) bool {
	_, ok := c.cache[key]
	return ok
}

func (c *TTLCache[K, V]) Invalidate(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.InvalidateUnsafe(key)
}

// InvalidateUnsafe is mutex-unprotected logic for Cache.Invalidate().
func (c *TTLCache[K, V]) InvalidateUnsafe(key K) bool {
	// Check if we have item with key
	item, ok := c.cache[key]
	if !ok {
		return false
	}

	// Call hook, remove from cache
	c.invalid(key, item.value)
	delete(c.cache, key)
	return true
}

func (c *TTLCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ClearUnsafe()
}

// ClearUnsafe is mutex-unprotected logic for Cache.Clean().
func (c *TTLCache[K, V]) ClearUnsafe() {
	for key, item := range c.cache {
		c.invalid(key, item.value)
		delete(c.cache, key)
	}
}

func (c *TTLCache[K, V]) Size() int {
	c.mu.Lock()
	sz := c.SizeUnsafe()
	c.mu.Unlock()
	return sz
}

// SizeUnsafe is mutex unprotected logic for Cache.Size().
func (c *TTLCache[K, V]) SizeUnsafe() int {
	return len(c.cache)
}

// entry represents an item in the cache, with
// it's currently calculated expiry time.
type entry[Value any] struct {
	value  Value
	expiry time.Time
}
