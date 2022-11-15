package cache

import (
	"time"

	ttlcache "codeberg.org/gruf/go-cache/v3/ttl"
)

// Cache represents a TTL cache with customizable callbacks, it exists here to abstract away the "unsafe" methods in the case that you do not want your own implementation atop ttl.Cache{}.
type Cache[Key comparable, Value any] interface {
	// Start will start the cache background eviction routine with given sweep frequency. If already running or a freq <= 0 provided, this is a no-op. This will block until the eviction routine has started.
	Start(freq time.Duration) bool

	// Stop will stop cache background eviction routine. If not running this is a no-op. This will block until the eviction routine has stopped.
	Stop() bool

	// SetEvictionCallback sets the eviction callback to the provided hook.
	SetEvictionCallback(hook func(*ttlcache.Entry[Key, Value]))

	// SetInvalidateCallback sets the invalidate callback to the provided hook.
	SetInvalidateCallback(hook func(*ttlcache.Entry[Key, Value]))

	// SetTTL sets the cache item TTL. Update can be specified to force updates of existing items in the cache, this will simply add the change in TTL to their current expiry time.
	SetTTL(ttl time.Duration, update bool)

	// Get fetches the value with key from the cache, extending its TTL.
	Get(key Key) (value Value, ok bool)

	// Add attempts to place the value at key in the cache, doing nothing if a value with this key already exists. Returned bool is success state.
	Add(key Key, value Value) bool

	// Set places the value at key in the cache. This will overwrite any existing value, and call the update callback so. Existing values will have their TTL extended upon update.
	Set(key Key, value Value)

	// CAS will attempt to perform a CAS operation on 'key', using provided old and new values, and comparator function. Returned bool is success.
	CAS(key Key, old, new Value, cmp func(Value, Value) bool) bool

	// Swap will attempt to perform a swap on 'key', replacing the value there and returning the existing value. If no value exists for key, this will set the value and return the zero value for V.
	Swap(key Key, swp Value) Value

	// Has checks the cache for a value with key, this will not update TTL.
	Has(key Key) bool

	// Invalidate deletes a value from the cache, calling the invalidate callback.
	Invalidate(key Key) bool

	// Clear empties the cache, calling the invalidate callback.
	Clear()

	// Len returns the current length of the cache.
	Len() int

	// Cap returns the maximum capacity of the cache.
	Cap() int
}

// New returns a new initialized Cache with given initial length, maximum capacity and item TTL.
func New[K comparable, V any](len, cap int, ttl time.Duration) Cache[K, V] {
	return ttlcache.New[K, V](len, cap, ttl)
}
