package result

import (
	"context"
	"reflect"
	"time"

	"codeberg.org/gruf/go-cache/v3/ttl"
	"codeberg.org/gruf/go-errors/v2"
)

// Lookup represents a struct object lookup method in the cache.
type Lookup struct {
	// Name is a period ('.') separated string
	// of struct fields this Key encompasses.
	Name string

	// AllowZero indicates whether to accept and cache
	// under zero value keys, otherwise ignore them.
	AllowZero bool

	// Multi allows specifying a key capable of storing
	// multiple results. Note this only supports invalidate.
	Multi bool

	// TODO: support toggling case sensitive lookups.
	// CaseSensitive bool
}

// Cache provides a means of caching value structures, along with
// the results of attempting to load them. An example usecase of this
// cache would be in wrapping a database, allowing caching of sql.ErrNoRows.
type Cache[Value any] struct {
	cache   ttl.Cache[int64, result[Value]] // underlying result cache
	invalid func(Value)                     // store unwrapped invalidate callback.
	lookups structKeys                      // pre-determined struct lookups
	ignore  func(error) bool                // determines cacheable errors
	copy    func(Value) Value               // copies a Value type
	next    int64                           // update key counter
}

// New returns a new initialized Cache, with given lookups, underlying value copy function and provided capacity.
func New[Value any](lookups []Lookup, copy func(Value) Value, cap int) *Cache[Value] {
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
		// Create keyed field info for lookup
		c.lookups[i] = newStructKey(lookup, t)
	}

	// Create and initialize underlying cache
	c.cache.Init(0, cap, 0)
	c.SetEvictionCallback(nil)
	c.SetInvalidateCallback(nil)
	c.IgnoreErrors(nil)
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
	c.cache.SetEvictionCallback(func(pkey int64, res result[Value]) {
		c.cache.Lock()
		for _, key := range res.Keys {
			// Delete key->pkey lookup
			pkeys := key.info.pkeys
			delete(pkeys, key.key)
		}
		c.cache.Unlock()

		if res.Error != nil {
			// Skip error hooks
			return
		}

		// Call user hook.
		hook(res.Value)
	})
}

// SetInvalidateCallback sets the invalidate callback to the provided hook.
func (c *Cache[Value]) SetInvalidateCallback(hook func(Value)) {
	if hook == nil {
		// Ensure non-nil hook.
		hook = func(Value) {}
	} // store hook.
	c.invalid = hook
	c.cache.SetInvalidateCallback(func(pkey int64, res result[Value]) {
		c.cache.Lock()
		for _, key := range res.Keys {
			// Delete key->pkey lookup
			pkeys := key.info.pkeys
			delete(pkeys, key.key)
		}
		c.cache.Unlock()

		if res.Error != nil {
			// Skip error hooks
			return
		}

		// Call user hook.
		hook(res.Value)
	})
}

// IgnoreErrors allows setting a function hook to determine which error types should / not be cached.
func (c *Cache[Value]) IgnoreErrors(ignore func(error) bool) {
	if ignore == nil {
		ignore = func(err error) bool {
			return errors.Comparable(
				err,
				context.Canceled,
				context.DeadlineExceeded,
			)
		}
	}
	c.cache.Lock()
	c.ignore = ignore
	c.cache.Unlock()
}

// Load will attempt to load an existing result from the cacche for the given lookup and key parts, else calling the provided load function and caching the result.
func (c *Cache[Value]) Load(lookup string, load func() (Value, error), keyParts ...any) (Value, error) {
	var (
		zero Value
		res  result[Value]
		ok   bool
	)

	// Get lookup key info by name.
	keyInfo := c.lookups.get(lookup)
	if !keyInfo.unique {
		panic("non-unique lookup does not support load: " + lookup)
	}

	// Generate cache key string.
	ckey := keyInfo.genKey(keyParts)

	// Acquire cache lock
	c.cache.Lock()

	// Look for primary cache key
	pkeys := keyInfo.pkeys[ckey]

	if ok = (len(pkeys) > 0); ok {
		var entry *ttl.Entry[int64, result[Value]]

		// Fetch the result for primary key
		entry, ok = c.cache.Cache.Get(pkeys[0])
		if ok {
			// Since the invalidation / eviction hooks acquire a mutex
			// lock separately, and only at this point are the pkeys
			// updated, there is a chance that a primary key may return
			// no matching entry. Hence we have to check for it here.
			res = entry.Value
		}
	}

	// Done with lock
	c.cache.Unlock()

	if !ok {
		// Generate fresh result.
		value, err := load()

		if err != nil {
			if c.ignore(err) {
				// don't cache this error type
				return zero, err
			}

			// Store error result.
			res.Error = err

			// This load returned an error, only
			// store this item under provided key.
			res.Keys = []cacheKey{{
				info: keyInfo,
				key:  ckey,
			}}
		} else {
			// Store value result.
			res.Value = value

			// This was a successful load, generate keys.
			res.Keys = c.lookups.generate(res.Value)
		}

		// Acquire cache lock.
		c.cache.Lock()
		defer c.cache.Unlock()

		// Cache result
		c.store(res)
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

	// Cache result
	c.store(result)

	// Call invalidate.
	c.invalid(value)

	return nil
}

// Has checks the cache for a positive result under the given lookup and key parts.
func (c *Cache[Value]) Has(lookup string, keyParts ...any) bool {
	var res result[Value]
	var ok bool

	// Get lookup key info by name.
	keyInfo := c.lookups.get(lookup)
	if !keyInfo.unique {
		panic("non-unique lookup does not support has: " + lookup)
	}

	// Generate cache key string.
	ckey := keyInfo.genKey(keyParts)

	// Acquire cache lock
	c.cache.Lock()

	// Look for primary key for cache key
	pkeys := keyInfo.pkeys[ckey]

	if ok = (len(pkeys) > 0); ok {
		var entry *ttl.Entry[int64, result[Value]]

		// Fetch the result for primary key
		entry, ok = c.cache.Cache.Get(pkeys[0])
		if ok {
			// Since the invalidation / eviction hooks acquire a mutex
			// lock separately, and only at this point are the pkeys
			// updated, there is a chance that a primary key may return
			// no matching entry. Hence we have to check for it here.
			res = entry.Value
		}
	}

	// Done with lock
	c.cache.Unlock()

	// Check for non-error result.
	return ok && (res.Error == nil)
}

// Invalidate will invalidate any result from the cache found under given lookup and key parts.
func (c *Cache[Value]) Invalidate(lookup string, keyParts ...any) {
	// Get lookup key info by name.
	keyInfo := c.lookups.get(lookup)

	// Generate cache key string.
	ckey := keyInfo.genKey(keyParts)

	// Look for primary key for cache key
	c.cache.Lock()
	pkeys := keyInfo.pkeys[ckey]
	delete(keyInfo.pkeys, ckey)
	c.cache.Unlock()

	// Invalidate all primary keys.
	c.cache.InvalidateAll(pkeys...)
}

// Clear empties the cache, calling the invalidate callback.
func (c *Cache[Value]) Clear() { c.cache.Clear() }

// store will cache this result under all of its required cache keys.
func (c *Cache[Value]) store(res result[Value]) {
	// Get primary key
	pnext := c.next
	c.next++
	if pnext > c.next {
		panic("cache primary key overflow")
	}

	for _, key := range res.Keys {
		// Look for cache primary keys.
		pkeys := key.info.pkeys[key.key]

		if key.info.unique && len(pkeys) > 0 {
			for _, conflict := range pkeys {
				// Get the overlapping result with this key.
				entry, _ := c.cache.Cache.Get(conflict)

				// From conflicting entry, drop this key, this
				// will prevent eviction cleanup key confusion.
				entry.Value.Keys.drop(key.info.name)

				if len(entry.Value.Keys) == 0 {
					// We just over-wrote the only lookup key for
					// this value, so we drop its primary key too.
					c.cache.Cache.Delete(conflict)
				}
			}

			// Drop existing.
			pkeys = pkeys[:0]
		}

		// Store primary key lookup.
		pkeys = append(pkeys, pnext)
		key.info.pkeys[key.key] = pkeys
	}

	// Store main entry under primary key, using evict hook if needed
	c.cache.Cache.SetWithHook(pnext, &ttl.Entry[int64, result[Value]]{
		Expiry: time.Now().Add(c.cache.TTL),
		Key:    pnext,
		Value:  res,
	}, func(_ int64, item *ttl.Entry[int64, result[Value]]) {
		c.cache.Evict(item.Key, item.Value)
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
