package cache

import "time"

// Cache represents a TTL cache with customizable callbacks, it
// exists here to abstract away the "unsafe" methods in the case that
// you do not want your own implementation atop TTLCache{}
type Cache interface {
	// SetEvictionCallback sets the eviction callback to the provided hook
	SetEvictionCallback(hook Hook)

	// SetInvalidateCallback sets the invalidate callback to the provided hook
	SetInvalidateCallback(hook Hook)

	// SetTTL sets the cache item TTL. Update can be specified to force updates of existing items in
	// the cache, this will simply add the change in TTL to their current expiry time
	SetTTL(ttl time.Duration, update bool)

	// Get fetches the value with key from the cache, extending its TTL
	Get(key string) (interface{}, bool)

	// Put attempts to place the value at key in the cache, doing nothing if
	// a value with this key already exists. Returned bool is success state
	Put(key string, value interface{}) bool

	// Set places the value at key in the cache. This will overwrite any
	// existing value, and call the update callback so. Existing values
	// will have their TTL extended upon update
	Set(key string, value interface{})

	// Has checks the cache for a value with key, this will not update TTL
	Has(key string) bool

	// Invalidate deletes a value from the cache, calling the invalidate callback
	Invalidate(key string) bool

	// Clear empties the cache, calling the invalidate callback
	Clear()

	// Size returns the current size of the cache
	Size() int
}

// New returns a new initialized Cache
func New() Cache {
	c := TTLCache{}
	c.Init()
	return &c
}
