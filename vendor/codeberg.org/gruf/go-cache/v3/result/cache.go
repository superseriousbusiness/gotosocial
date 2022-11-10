package result

import (
	"reflect"
	"time"

	"codeberg.org/gruf/go-cache/v3/ttl"
)

// Cache ...
type Cache[Value any] struct {
	cache   ttl.Cache[int64, result[Value]] // underlying result cache
	lookups structKeys                      // pre-determined struct lookups
	copy    func(Value) Value               // copies a Value type
	next    int64                           // update key counter
}

// New returns a new initialized Cache, with given lookups and underlying value copy function.
func New[Value any](lookups []string, copy func(Value) Value) *Cache[Value] {
	return NewSized(lookups, copy, 64)
}

// NewSized returns a new initialized Cache, with given lookups, underlying value copy function and provided capacity.
func NewSized[Value any](lookups []string, copy func(Value) Value, cap int) *Cache[Value] {
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

	// Allocate new cache object
	c := &Cache[Value]{copy: copy}
	c.lookups = make([]keyFields, len(lookups))

	for i, lookup := range lookups {
		// Generate keyed field info for lookup
		c.lookups[i].pkeys = make(map[string]int64, cap)
		c.lookups[i].lookup = lookup
		c.lookups[i].populate(t)
	}

	// Create and initialize underlying cache
	c.cache.Init(0, cap, 0)
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
	c.cache.SetEvictionCallback(func(item *ttl.Entry[int64, result[Value]]) {
		for _, key := range item.Value.Keys {
			// Delete key->pkey lookup
			pkeys := key.fields.pkeys
			delete(pkeys, key.value)
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
	c.cache.SetInvalidateCallback(func(item *ttl.Entry[int64, result[Value]]) {
		for _, key := range item.Value.Keys {
			if key.fields != nil {
				// Delete key->pkey lookup
				pkeys := key.fields.pkeys
				delete(pkeys, key.value)
			}
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
	var (
		zero Value
		res  result[Value]
	)

	// Get lookup map by name.
	lmap := c.getLookup(lookup)

	// Generate cache key string.
	ckey := genkey(keyParts...)

	// Acquire cache lock
	c.cache.Lock()

	// Look for primary key
	pkey, ok := lmap[ckey]

	if ok {
		// Fetch the result for primary key
		entry, _ := c.cache.Cache.Get(pkey)
		res = entry.Value
	}

	// Done with lock
	c.cache.Unlock()

	if !ok {
		// Generate new result from fresh load.
		res.Value, res.Error = load()

		if res.Error != nil {
			// This load returned an error, only
			// store this item under provided key.
			res.Keys = []cacheKey{{value: ckey}}
		} else {
			// This was a successful load, generate keys.
			res.Keys = c.lookups.generate(res.Value)
		}

		// Acquire cache lock.
		c.cache.Lock()
		defer c.cache.Unlock()

		// Attempt to cache this result.
		if key, ok := c.storeResult(res); !ok {
			return zero, ConflictError{key}
		}
	}

	// Catch and return error
	if res.Error != nil {
		return zero, res.Error
	}

	// Return a copy of value from cache
	return c.copy(res.Value), nil
}

// Store ...
func (c *Cache[Value]) Store(value Value, store func() error) error {
	// Attempt to store this value.
	if err := store(); err != nil {
		return err
	}

	// Prepare cached result.
	result := result[Value]{
		Keys:  c.lookups.generate(value),
		Value: c.copy(value),
		Error: nil,
	}

	// Acquire cache lock.
	c.cache.Lock()
	defer c.cache.Unlock()

	// Attempt to cache result, only return conflict
	// error if the appropriate flag has been set.
	if key, ok := c.storeResult(result); !ok {
		return ConflictError{key}
	}

	return nil
}

// Has ...
func (c *Cache[Value]) Has(lookup string, keyParts ...any) bool {
	var res result[Value]

	// Get lookup map by name.
	lmap := c.getLookup(lookup)

	// Generate cache key string.
	ckey := genkey(keyParts...)

	// Acquire cache lock
	c.cache.Lock()

	// Look for primary key
	pkey, ok := lmap[ckey]

	if ok {
		// Fetch the result for primary key
		entry, _ := c.cache.Cache.Get(pkey)
		res = entry.Value
	}

	// Done with lock
	c.cache.Unlock()

	// Check for non-error result.
	return ok && (res.Error == nil)
}

// Invalidate ...
func (c *Cache[Value]) Invalidate(lookup string, keyParts ...any) {
	// Get lookup map by name.
	lmap := c.getLookup(lookup)

	// Generate cache key string.
	ckey := genkey(keyParts...)

	// Look for primary key
	c.cache.Lock()
	pkey, ok := lmap[ckey]
	c.cache.Unlock()

	if !ok {
		return
	}

	// Invalid by primary key
	c.cache.Invalidate(pkey)
}

// Clear empties the cache, calling the invalidate callback.
func (c *Cache[Value]) Clear() {
	c.cache.Clear()
}

// Len ...
func (c *Cache[Value]) Len() int {
	return c.cache.Cache.Len()
}

// Cap ...
func (c *Cache[Value]) Cap() int {
	return c.cache.Cache.Cap()
}

func (c *Cache[Value]) getLookup(name string) map[string]int64 {
	for _, l := range c.lookups {
		// Find lookup map with name
		if l.lookup == name {
			return l.pkeys
		}
	}
	panic("invalid lookup: " + name)
}

func (c *Cache[Value]) storeResult(res result[Value]) (string, bool) {
	for _, key := range res.Keys {
		pkeys := key.fields.pkeys

		// Look for cache primary key
		pkey, ok := pkeys[key.value]

		if ok {
			// Look for overlap with non error keys,
			// as an overlap for some but not all keys
			// could produce inconsistent results.
			entry, _ := c.cache.Cache.Get(pkey)
			if entry.Value.Error == nil {
				return key.value, false
			}
		}
	}

	// Get primary key
	pkey := c.next
	c.next++

	// Store all primary key lookups
	for _, key := range res.Keys {
		pkeys := key.fields.pkeys
		pkeys[key.value] = pkey
	}

	// Store main entry under primary key, using evict hook if needed
	c.cache.Cache.SetWithHook(pkey, &ttl.Entry[int64, result[Value]]{
		Expiry: time.Now().Add(c.cache.TTL),
		Key:    pkey,
		Value:  res,
	}, func(_ int64, item *ttl.Entry[int64, result[Value]]) {
		c.cache.Evict(item)
	})

	return "", true
}

type result[Value any] struct {
	// keys accessible under
	Keys []cacheKey

	// cached value
	Value Value

	// cached error
	Error error
}
