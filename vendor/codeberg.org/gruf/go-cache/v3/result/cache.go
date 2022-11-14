package result

import (
	"reflect"
	"time"

	"codeberg.org/gruf/go-cache/v3/ttl"
)

// Lookup represents a struct object lookup method in the cache.
type Lookup struct {
	// Name is a period ('.') separated string
	// of struct fields this Key encompasses.
	Name string

	// AllowZero indicates whether to accept and cache
	// under zero value keys, otherwise ignore them.
	AllowZero bool
}

// Cache provides a means of caching value structures, along with
// the results of attempting to load them. An example usecase of this
// cache would be in wrapping a database, allowing caching of sql.ErrNoRows.
type Cache[Value any] struct {
	cache   ttl.Cache[int64, result[Value]] // underlying result cache
	lookups structKeys                      // pre-determined struct lookups
	copy    func(Value) Value               // copies a Value type
	next    int64                           // update key counter
}

// New returns a new initialized Cache, with given lookups and underlying value copy function.
func New[Value any](lookups []Lookup, copy func(Value) Value) *Cache[Value] {
	return NewSized(lookups, copy, 64)
}

// NewSized returns a new initialized Cache, with given lookups, underlying value copy function and provided capacity.
func NewSized[Value any](lookups []Lookup, copy func(Value) Value, cap int) *Cache[Value] {
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
	c.lookups = make([]structKey, len(lookups))

	for i, lookup := range lookups {
		// Generate keyed field info for lookup
		c.lookups[i] = genStructKey(lookup, t)
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
			pkeys := key.key.pkeys
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
			// Delete key->pkey lookup
			pkeys := key.key.pkeys
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

// Load will attempt to load an existing result from the cacche for the given lookup and key parts, else calling the load function and caching that result.
func (c *Cache[Value]) Load(lookup string, load func() (Value, error), keyParts ...any) (Value, error) {
	var (
		zero Value
		res  result[Value]
	)

	// Get lookup key info by name.
	keyInfo := c.lookups.get(lookup)

	// Generate cache key string.
	ckey := genKey(keyParts...)

	// Acquire cache lock
	c.cache.Lock()

	// Look for primary key for cache key
	pkey, ok := keyInfo.pkeys[ckey]

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
			res.Keys = []cachedKey{{
				key:   keyInfo,
				value: ckey,
			}}
		} else {
			// This was a successful load, generate keys.
			res.Keys = c.lookups.generate(res.Value)
		}

		// Acquire cache lock.
		c.cache.Lock()
		defer c.cache.Unlock()

		// Cache this result
		c.storeResult(res)
	}

	// Catch and return error
	if res.Error != nil {
		return zero, res.Error
	}

	// Return a copy of value from cache
	return c.copy(res.Value), nil
}

// Store will call the given store function, and on success store the value in the cache as a positive result.
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

	// Cache this result
	c.storeResult(result)

	return nil
}

// Has checks the cache for a positive result under the given lookup and key parts.
func (c *Cache[Value]) Has(lookup string, keyParts ...any) bool {
	var res result[Value]

	// Get lookup key type by name.
	keyType := c.lookups.get(lookup)

	// Generate cache key string.
	ckey := genKey(keyParts...)

	// Acquire cache lock
	c.cache.Lock()

	// Look for primary key for cache key
	pkey, ok := keyType.pkeys[ckey]

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

// Invalidate will invalidate any result from the cache found under given lookup and key parts.
func (c *Cache[Value]) Invalidate(lookup string, keyParts ...any) {
	// Get lookup key type by name.
	keyType := c.lookups.get(lookup)

	// Generate cache key string.
	ckey := genKey(keyParts...)

	// Look for primary key for cache key
	c.cache.Lock()
	pkey, ok := keyType.pkeys[ckey]
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

// Len returns the current length of the cache.
func (c *Cache[Value]) Len() int {
	return c.cache.Cache.Len()
}

// Cap returns the maximum capacity of this result cache.
func (c *Cache[Value]) Cap() int {
	return c.cache.Cache.Cap()
}

func (c *Cache[Value]) storeResult(res result[Value]) {
	for _, key := range res.Keys {
		pkeys := key.key.pkeys

		// Look for cache primary key
		pkey, ok := pkeys[key.value]

		if ok {
			// Get the overlapping result with this key.
			entry, _ := c.cache.Cache.Get(pkey)

			// From conflicting entry, drop this key, this
			// will prevent eviction cleanup key confusion.
			entry.Value.Keys.drop(key.key.name)

			if len(entry.Value.Keys) == 0 {
				// We just over-wrote the only lookup key for
				// this value, so we drop its primary key too
				c.cache.Cache.Delete(pkey)
			}
		}
	}

	// Get primary key
	pkey := c.next
	c.next++

	// Store all primary key lookups
	for _, key := range res.Keys {
		pkeys := key.key.pkeys
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
}

type result[Value any] struct {
	// keys accessible under
	Keys cacheKeys

	// cached value
	Value Value

	// cached error
	Error error
}
