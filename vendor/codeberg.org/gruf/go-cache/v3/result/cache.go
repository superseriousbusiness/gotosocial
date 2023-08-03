package result

import (
	"context"
	"fmt"
	"os"
	"reflect"
	_ "unsafe"

	"codeberg.org/gruf/go-cache/v3/simple"
	"codeberg.org/gruf/go-errors/v2"
)

type result struct {
	// keys accessible under
	Keys cacheKeys

	// cached value
	Value any

	// cached error
	Error error
}

// getResultValue is a safe way of casting and fetching result value.
func getResultValue[T any](res *result) T {
	v, ok := res.Value.(T)
	if !ok {
		fmt.Fprintf(os.Stderr, "!! BUG: unexpected value type in result: %T\n", res.Value)
	}
	return v
}

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
}

// Cache provides a means of caching value structures, along with
// the results of attempting to load them. An example usecase of this
// cache would be in wrapping a database, allowing caching of sql.ErrNoRows.
type Cache[T any] struct {
	cache   simple.Cache[int64, *result] // underlying result cache
	lookups structKeys                   // pre-determined struct lookups
	invalid func(T)                      // store unwrapped invalidate callback.
	ignore  func(error) bool             // determines cacheable errors
	copy    func(T) T                    // copies a Value type
	next    int64                        // update key counter
}

// New returns a new initialized Cache, with given lookups, underlying value copy function and provided capacity.
func New[T any](lookups []Lookup, copy func(T) T, cap int) *Cache[T] {
	var z T

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
	c := &Cache[T]{copy: copy}
	c.lookups = make([]structKey, len(lookups))

	for i, lookup := range lookups {
		// Create keyed field info for lookup
		c.lookups[i] = newStructKey(lookup, t)
	}

	// Create and initialize underlying cache
	c.cache.Init(0, cap)
	c.SetEvictionCallback(nil)
	c.SetInvalidateCallback(nil)
	c.IgnoreErrors(nil)
	return c
}

// SetEvictionCallback sets the eviction callback to the provided hook.
func (c *Cache[T]) SetEvictionCallback(hook func(T)) {
	if hook == nil {
		// Ensure non-nil hook.
		hook = func(T) {}
	}
	c.cache.SetEvictionCallback(func(pkey int64, res *result) {
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

		// Free result and call hook.
		v := getResultValue[T](res)
		putResult(res)
		hook(v)
	})
}

// SetInvalidateCallback sets the invalidate callback to the provided hook.
func (c *Cache[T]) SetInvalidateCallback(hook func(T)) {
	if hook == nil {
		// Ensure non-nil hook.
		hook = func(T) {}
	} // store hook.
	c.invalid = hook
	c.cache.SetInvalidateCallback(func(pkey int64, res *result) {
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

		// Free result and call hook.
		v := getResultValue[T](res)
		putResult(res)
		hook(v)
	})
}

// IgnoreErrors allows setting a function hook to determine which error types should / not be cached.
func (c *Cache[T]) IgnoreErrors(ignore func(error) bool) {
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
func (c *Cache[T]) Load(lookup string, load func() (T, error), keyParts ...any) (T, error) {
	var (
		zero T
		res  *result
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

	if len(pkeys) > 0 {
		// Fetch the result for primary key
		entry, ok := c.cache.Cache.Get(pkeys[0])
		if ok {
			// Since the invalidation / eviction hooks acquire a mutex
			// lock separately, and only at this point are the pkeys
			// updated, there is a chance that a primary key may return
			// no matching entry. Hence we have to check for it here.
			res = entry.Value.(*result)
		}
	}

	// Done with lock
	c.cache.Unlock()

	if res == nil {
		// Generate fresh result.
		value, err := load()

		if err != nil {
			if c.ignore(err) {
				// don't cache this error type
				return zero, err
			}

			// Alloc result.
			res = getResult()

			// Store error result.
			res.Error = err

			// This load returned an error, only
			// store this item under provided key.
			res.Keys = []cacheKey{{
				info: keyInfo,
				key:  ckey,
			}}
		} else {
			// Alloc result.
			res = getResult()

			// Store value result.
			res.Value = value

			// This was a successful load, generate keys.
			res.Keys = c.lookups.generate(res.Value)
		}

		var evict func()

		// Lock cache.
		c.cache.Lock()

		defer func() {
			// Unlock cache.
			c.cache.Unlock()

			if evict != nil {
				// Call evict.
				evict()
			}
		}()

		// Store result in cache.
		evict = c.store(res)
	}

	// Catch and return error
	if res.Error != nil {
		return zero, res.Error
	}

	// Return a copy of value from cache
	return c.copy(getResultValue[T](res)), nil
}

// Store will call the given store function, and on success store the value in the cache as a positive result.
func (c *Cache[T]) Store(value T, store func() error) error {
	// Attempt to store this value.
	if err := store(); err != nil {
		return err
	}

	// Prepare cached result.
	result := getResult()
	result.Keys = c.lookups.generate(value)
	result.Value = c.copy(value)
	result.Error = nil

	var evict func()

	// Lock cache.
	c.cache.Lock()

	defer func() {
		// Unlock cache.
		c.cache.Unlock()

		if evict != nil {
			// Call evict.
			evict()
		}

		// Call invalidate.
		c.invalid(value)
	}()

	// Store result in cache.
	evict = c.store(result)

	return nil
}

// Has checks the cache for a positive result under the given lookup and key parts.
func (c *Cache[T]) Has(lookup string, keyParts ...any) bool {
	var res *result

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

	if len(pkeys) > 0 {
		// Fetch the result for primary key
		entry, ok := c.cache.Cache.Get(pkeys[0])
		if ok {
			// Since the invalidation / eviction hooks acquire a mutex
			// lock separately, and only at this point are the pkeys
			// updated, there is a chance that a primary key may return
			// no matching entry. Hence we have to check for it here.
			res = entry.Value.(*result)
		}
	}

	// Done with lock
	c.cache.Unlock()

	// Check for result AND non-error result.
	return (res != nil && res.Error == nil)
}

// Invalidate will invalidate any result from the cache found under given lookup and key parts.
func (c *Cache[T]) Invalidate(lookup string, keyParts ...any) {
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

// Clear empties the cache, calling the invalidate callback where necessary.
func (c *Cache[T]) Clear() { c.Trim(100) }

// Trim ensures the cache stays within percentage of total capacity, truncating where necessary.
func (c *Cache[T]) Trim(perc float64) {
	c.cache.Trim(perc)
}

// store will cache this result under all of its required cache keys.
func (c *Cache[T]) store(res *result) (evict func()) {
	var toEvict []struct {
		pkey int64
		res  *result
	}

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
				res := entry.Value.(*result)

				// From conflicting entry, drop this key, this
				// will prevent eviction cleanup key confusion.
				res.Keys.drop(key.info.name)

				if len(res.Keys) == 0 {
					// We just over-wrote the only lookup key for
					// this value, so we drop its primary key too.
					c.cache.Cache.Delete(conflict)

					// Add to later evictions slice.
					toEvict = append(toEvict, struct {
						pkey int64
						res  *result
					}{
						pkey: conflict,
						res:  res,
					})
				}
			}

			// Drop existing.
			pkeys = pkeys[:0]
		}

		// Store primary key lookup.
		pkeys = append(pkeys, pnext)
		key.info.pkeys[key.key] = pkeys
	}

	// Store main entry under primary key, catch evicted.
	c.cache.Cache.SetWithHook(pnext, &simple.Entry{
		Key:   pnext,
		Value: res,
	}, func(pkey int64, item *simple.Entry) {
		toEvict = append(toEvict, struct {
			pkey int64
			res  *result
		}{
			pkey: pkey,
			res:  item.Value.(*result),
		})
	})

	if len(toEvict) == 0 {
		// none evicted.
		return nil
	}

	return func() {
		for _, evict := range toEvict {
			// Call evict hook on each entry.
			c.cache.Evict(evict.pkey, evict.res)
		}
	}
}
