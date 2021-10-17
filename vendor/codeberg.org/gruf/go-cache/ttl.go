package cache

import (
	"sync"
	"time"

	"codeberg.org/gruf/go-nowish"
)

const (
	// clockPrecision is the precision of the cacheClock
	clockPrecision = time.Millisecond * 100

	// maxLockAttempts is the max no. attempted mutex locks by the reaper before
	// enforcing a lock at the next attempt
	maxLockAttempts = 5
)

var (
	// caches is the global tracked slice of Cache objects, used by the reaper
	caches = []*TTLCache{}

	// cacheClock is the cache-entry clock, used for TTL checking
	cacheClock = nowish.Clock{}

	// globalOnce protects the cacheClock and reaper from multiple attempted starts
	globalOnce = sync.Once{}

	// globalMutex protects the caches slice
	globalMutex = sync.Mutex{}
)

// TTLCache is the underlying Cache implementation, providing both the base
// Cache interface and access to "unsafe" methods so that you may build your
// customized caches ontop of this structure
type TTLCache struct {
	cache   map[string]*entry
	evict   Hook // the evict hook is called when an item is evicted from the cache, includes manual delete
	invalid Hook // the invalidate hook is called when an item's data in the cache is invalidated
	ttl     time.Duration
	mutex   mutex
}

// Init performs Cache initialization, this MUST be called
func (c *TTLCache) Init() {
	// Initialize the cache itself
	c.cache = make(map[string]*entry, 100)
	c.evict = emptyHook
	c.invalid = emptyHook
	c.ttl = time.Minute * 5

	// Add to global caches
	globalMutex.Lock()
	caches = append(caches, c)
	globalMutex.Unlock()

	// Ensure reaper & clock started
	globalOnce.Do(func() {
		cacheClock.Start(clockPrecision)
		go reaper()
	})
}

// sweep attempts to evict expired items (with callback!) from cache
func (c *TTLCache) sweep() {
	// Attempt to get cache lock
	if !c.mutex.AttemptLock() {
		return
	}

	// Defer in case of hook panic
	defer c.mutex.Unlock()

	// Fetch current time for TTL check
	now := cacheClock.Now()

	// Sweep the cache for old items!
	for key, item := range c.cache {
		if now.After(item.expiry) {
			c.evict(key, item.value)
			delete(c.cache, key)
		}
	}
}

// Lock locks the cache mutex
func (c *TTLCache) Lock() {
	c.mutex.Lock()
}

// Unlock unlocks the cache mutex
func (c *TTLCache) Unlock() {
	c.mutex.Unlock()
}

func (c *TTLCache) SetEvictionCallback(hook Hook) {
	// Ensure non-nil hook
	if hook == nil {
		hook = emptyHook
	}

	// Safely set evict hook
	c.Lock()
	c.evict = hook
	c.Unlock()
}

func (c *TTLCache) SetInvalidateCallback(hook Hook) {
	// Ensure non-nil hook
	if hook == nil {
		hook = emptyHook
	}

	// Safely set invalidate hook
	c.Lock()
	c.invalid = hook
	c.Unlock()
}

func (c *TTLCache) SetTTL(ttl time.Duration, update bool) {
	if ttl < clockPrecision*10 && ttl > 0 {
		// A zero TTL means nothing expires,
		// but other small values we check to
		// ensure they won't be lost by our
		// unprecise cache clock
		panic("ttl too close to cache clock precision")
	}

	// Safely update TTL
	c.Lock()
	diff := ttl - c.ttl
	c.ttl = ttl

	if update {
		// Update existing cache entries
		for _, entry := range c.cache {
			entry.expiry.Add(diff)
		}
	}

	// We're done
	c.Unlock()
}

func (c *TTLCache) Get(key string) (interface{}, bool) {
	c.Lock()
	value, ok := c.GetUnsafe(key)
	c.Unlock()
	return value, ok
}

// GetUnsafe is the mutex-unprotected logic for Cache.Get()
func (c *TTLCache) GetUnsafe(key string) (interface{}, bool) {
	item, ok := c.cache[key]
	if !ok {
		return nil, false
	}
	item.expiry = cacheClock.Now().Add(c.ttl)
	return item.value, true
}

func (c *TTLCache) Put(key string, value interface{}) bool {
	c.Lock()
	success := c.PutUnsafe(key, value)
	c.Unlock()
	return success
}

// PutUnsafe is the mutex-unprotected logic for Cache.Put()
func (c *TTLCache) PutUnsafe(key string, value interface{}) bool {
	// If already cached, return
	if _, ok := c.cache[key]; ok {
		return false
	}

	// Create new cached item
	c.cache[key] = &entry{
		value:  value,
		expiry: cacheClock.Now().Add(c.ttl),
	}

	return true
}

func (c *TTLCache) Set(key string, value interface{}) {
	c.Lock()
	defer c.Unlock() // defer in case of hook panic
	c.SetUnsafe(key, value)
}

// SetUnsafe is the mutex-unprotected logic for Cache.Set(), it calls externally-set functions
func (c *TTLCache) SetUnsafe(key string, value interface{}) {
	item, ok := c.cache[key]
	if ok {
		// call invalidate hook
		c.invalid(key, item.value)
	} else {
		// alloc new item
		item = &entry{}
		c.cache[key] = item
	}

	// Update the item + expiry
	item.value = value
	item.expiry = cacheClock.Now().Add(c.ttl)
}

func (c *TTLCache) Has(key string) bool {
	c.Lock()
	ok := c.HasUnsafe(key)
	c.Unlock()
	return ok
}

// HasUnsafe is the mutex-unprotected logic for Cache.Has()
func (c *TTLCache) HasUnsafe(key string) bool {
	_, ok := c.cache[key]
	return ok
}

func (c *TTLCache) Invalidate(key string) bool {
	c.Lock()
	defer c.Unlock()
	return c.InvalidateUnsafe(key)
}

// InvalidateUnsafe is mutex-unprotected logic for Cache.Invalidate()
func (c *TTLCache) InvalidateUnsafe(key string) bool {
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

func (c *TTLCache) Clear() {
	c.Lock()
	defer c.Unlock()
	c.ClearUnsafe()
}

// ClearUnsafe is mutex-unprotected logic for Cache.Clean()
func (c *TTLCache) ClearUnsafe() {
	for key, item := range c.cache {
		c.invalid(key, item.value)
		delete(c.cache, key)
	}
}

func (c *TTLCache) Size() int {
	c.Lock()
	sz := c.SizeUnsafe()
	c.Unlock()
	return sz
}

// SizeUnsafe is mutex unprotected logic for Cache.Size()
func (c *TTLCache) SizeUnsafe() int {
	return len(c.cache)
}

// entry represents an item in the cache, with
// it's currently calculated expiry time
type entry struct {
	value  interface{}
	expiry time.Time
}
