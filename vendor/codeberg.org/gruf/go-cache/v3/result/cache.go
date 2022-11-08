package result

import (
	"reflect"
	"time"

	"codeberg.org/gruf/go-cache/v3/ttl"
)

// Cache ...
type Cache[Value any] struct {
	cache ttl.Cache[string, result[Value]] // underlying result cache
	keys  structKeys                       // pre-determined generic type struct keys
	copy  func(Value) Value                // copies a Value type
}

// New ...
func New[Value any](lookups []string, copy func(Value) Value) *Cache[Value] {
	return NewSized(lookups, copy, 64)
}

// NewSized ...
func NewSized[Value any](lookups []string, copy func(Value) Value, sz int) *Cache[Value] {
	var z Value

	// Determine generic type
	t := reflect.TypeOf(z)

	// Iteratively deref pointer type
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	// Ensure that this is a struct type
	if t.Kind() != reflect.Struct {
		panic("generic parameter type must be struct (or ptr to)")
	}

	// Preallocate a slice of keyed fields info
	keys := make([]keyFields, len(lookups))

	for i, lookup := range lookups {
		// Generate keyed field info for lookup
		keys[i] = keyFields{prefix: lookup}
		keys[i].populate(t)
	}

	// Create and initialize
	c := &Cache[Value]{keys: keys, copy: copy}
	c.cache.Init(0, 100, 0)
	c.SetEvictionCallback(nil)
	c.SetInvalidateCallback(nil)
	return c
}

// Start will start the cache background eviction routine with given sweep frequency. If already
// running or a freq <= 0 provided, this is a no-op. This will block until eviction routine started.
func (c *Cache[Value]) Start(freq time.Duration) bool {
	return c.cache.Start(freq)
}

// Stop will stop cache background eviction routine. If not running this
// is a no-op. This will block until the eviction routine has stopped.
func (c *Cache[Value]) Stop() bool {
	return c.cache.Stop()
}

// SetTTL sets the cache item TTL. Update can be specified to force updates of existing items
// in the cache, this will simply add the change in TTL to their current expiry time.
func (c *Cache[Value]) SetTTL(ttl time.Duration, update bool) {
	c.cache.SetTTL(ttl, update)
}

// SetEvictionCallback sets the eviction callback to the provided hook.
func (c *Cache[Value]) SetEvictionCallback(hook func(Value)) {
	if hook == nil {
		// Ensure non-nil hook.
		hook = func(Value) {}
	}
	c.cache.SetEvictionCallback(func(item *ttl.Entry[string, result[Value]]) {
		for i := range item.Value.Keys {
			// This is "us", already deleted.
			if item.Value.Keys[i].value == item.Key {
				continue
			}

			// Manually delete this extra cache key.
			c.cache.Cache.Delete(item.Value.Keys[i].value)
		}

		if item.Value.Error != nil {
			// Skip error hooks
			return
		}

		// Call user hook.
		hook(item.Value.Value)
	})
}

// SetInvalidateCallback sets the invalidate callback to the provided hook.
func (c *Cache[Value]) SetInvalidateCallback(hook func(Value)) {
	if hook == nil {
		// Ensure non-nil hook.
		hook = func(Value) {}
	}
	c.cache.SetInvalidateCallback(func(item *ttl.Entry[string, result[Value]]) {
		for i := range item.Value.Keys {
			// This is "us", already deleted.
			if item.Value.Keys[i].value == item.Key {
				continue
			}

			// Manually delete this extra cache key.
			c.cache.Cache.Delete(item.Value.Keys[i].value)
		}

		if item.Value.Error != nil {
			// Skip error hooks
			return
		}

		// Call user hook.
		hook(item.Value.Value)
	})
}

// Load ...
func (c *Cache[Value]) Load(lookup string, load func() (Value, error), keyParts ...any) (Value, error) {
	var zero Value

	// Generate cache key string.
	ckey := genkey(lookup, keyParts...)

	// Look for existing result in cache.
	result, ok := c.cache.Get(ckey)

	if !ok {
		// Generate new result from fresh load.
		result.Value, result.Error = load()

		if result.Error != nil {
			// This load returned an error, only
			// store this item under provided key.
			result.Keys = []cacheKey{{value: ckey}}
		} else {
			// This was a successful load, generate keys.
			result.Keys = c.keys.generate(result.Value)
		}

		// Acquire cache lock.
		c.cache.Lock()
		defer c.cache.Unlock()

		// Attempt to cache result, only return conflict
		// error if the appropriate flag has been set.
		if key, ok := c.store(result); !ok {
			return zero, ConflictError{key}
		}
	}

	// Catch and return error
	if result.Error != nil {
		return zero, result.Error
	}

	// Return a copy of value from cache
	return c.copy(result.Value), nil
}

// Store ...
func (c *Cache[Value]) Store(value Value, store func() error) error {
	// Attempt to store this value.
	if err := store(); err != nil {
		return err
	}

	// Prepare cached result.
	result := result[Value]{
		Keys:  c.keys.generate(value),
		Value: c.copy(value),
		Error: nil,
	}

	// Acquire cache lock.
	c.cache.Lock()
	defer c.cache.Unlock()

	// Attempt to cache result, only return conflict
	// error if the appropriate flag has been set.
	if key, ok := c.store(result); !ok {
		return ConflictError{key}
	}

	return nil
}

// store will store a given result in the cache, returning the key string
// and 'false' on any conflict. Note this function MUST be called within
// the underlying cache's mutex lock as it makes calls to TTLCache{}.__Unsafe().
func (c *Cache[Value]) store(r result[Value]) (string, bool) {
	// Check for overlapy with any NON-ERROR keys, as an
	// overlap will cause say one but not all of
	// an item's keys to produce unexpected results.
	for _, key := range r.Keys {
		if entry, ok := c.cache.Cache.Get(key.value); ok {
			if entry.Value.Error == nil {
				return key.value, false
			}
		}
	}

	// Determine cached result expiry time
	expiry := time.Now().Add(c.cache.TTL)

	// Store this result under all keys.
	for _, key := range r.Keys {
		c.cache.Cache.Set(key.value, &ttl.Entry[string, result[Value]]{
			Key:    key.value,
			Value:  r,
			Expiry: expiry,
		})
	}

	return "", true
}

// Has ...
func (c *Cache[Value]) Has(lookup string, keyParts ...any) bool {
	// Generate cache key string.
	ckey := genkey(lookup, keyParts...)

	// Check for non-error result.
	result, ok := c.cache.Get(ckey)
	return ok && (result.Error == nil)
}

// Invalidate ...
func (c *Cache[Value]) Invalidate(lookup string, keyParts ...any) {
	// Generate cache key string.
	ckey := genkey(lookup, keyParts...)

	// Invalidate this key from cache.
	c.cache.Invalidate(ckey)
}

// Clear empties the cache, calling the invalidate callback.
func (cache *Cache[Value]) Clear() {
	cache.cache.Clear()
}

type result[Value any] struct {
	// keys accessible under
	Keys []cacheKey

	// cached value
	Value Value

	// cached error
	Error error
}
