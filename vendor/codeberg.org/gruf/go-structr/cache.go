package structr

import (
	"context"
	"errors"
	"sync"
)

// DefaultIgnoreErr is the default function used to
// ignore (i.e. not cache) incoming error results during
// Load() calls. By default ignores context pkg errors.
func DefaultIgnoreErr(err error) bool {
	return errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded)
}

// Config defines config variables
// for initializing a struct cache.
type Config[StructType any] struct {

	// Indices defines indices to create
	// in the Cache for the receiving
	// generic struct type parameter.
	Indices []IndexConfig

	// MaxSize defines the maximum number
	// of results allowed in the Cache at
	// one time, before old results start
	// getting evicted.
	MaxSize int

	// IgnoreErr defines which errors to
	// ignore (i.e. not cache) returned
	// from load function callback calls.
	// This may be left as nil, on which
	// DefaultIgnoreErr will be used.
	IgnoreErr func(error) bool

	// CopyValue provides a means of copying
	// cached values, to ensure returned values
	// do not share memory with those in cache.
	CopyValue func(StructType) StructType

	// Invalidate is called when cache values
	// (NOT errors) are invalidated, either
	// as the values passed to Put() / Store(),
	// or by the keys by calls to Invalidate().
	Invalidate func(StructType)
}

// Cache provides a structure cache with automated
// indexing and lookups by any initialization-defined
// combination of fields (as long as serialization is
// supported by codeberg.org/gruf/go-mangler). This
// also supports caching of negative results by errors
// returned from the LoadOne() series of functions.
type Cache[StructType any] struct {

	// indices used in storing passed struct
	// types by user defined sets of fields.
	indices []Index[StructType]

	// keeps track of all indexed results,
	// in order of last recently used (LRU).
	lruList list

	// max cache size, imposes size
	// limit on the lruList in order
	// to evict old entries.
	maxSize int

	// hook functions.
	ignore  func(error) bool
	copy    func(StructType) StructType
	invalid func(StructType)

	// protective mutex, guards:
	// - Cache{}.lruList
	// - Index{}.data
	// - Cache{} hook fns
	mutex sync.Mutex
}

// Init initializes the cache with given configuration
// including struct fields to index, and necessary fns.
func (c *Cache[T]) Init(config Config[T]) {
	if len(config.Indices) == 0 {
		panic("no indices provided")
	}

	if config.IgnoreErr == nil {
		config.IgnoreErr = DefaultIgnoreErr
	}

	if config.CopyValue == nil {
		panic("copy value function must be provided")
	}

	if config.MaxSize < 2 {
		panic("minimum cache size is 2 for LRU to work")
	}

	// Safely copy over
	// provided config.
	c.mutex.Lock()
	c.indices = make([]Index[T], len(config.Indices))
	for i, cfg := range config.Indices {
		init_index(&c.indices[i], cfg, config.MaxSize)
	}
	c.ignore = config.IgnoreErr
	c.copy = config.CopyValue
	c.invalid = config.Invalidate
	c.maxSize = config.MaxSize
	c.mutex.Unlock()
}

// Index selects index with given name from cache, else panics.
func (c *Cache[T]) Index(name string) *Index[T] {
	for i := range c.indices {
		if c.indices[i].name == name {
			return &c.indices[i]
		}
	}
	panic("unknown index: " + name)
}

// GetOne fetches one value from the cache stored under index, using key generated from key parts.
// Note that given number of key parts MUST match expected number and types of the given index name.
func (c *Cache[T]) GetOne(index string, key ...any) (T, bool) {
	return c.GetOneBy(c.Index(index), key...)
}

// GetOneBy fetches value from cache stored under index, using precalculated index key.
func (c *Cache[T]) GetOneBy(index *Index[T], key ...any) (T, bool) {
	if index == nil {
		panic("no index given")
	} else if !is_unique(index.flags) {
		panic("cannot get one by non-unique index")
	}
	values := c.GetBy(index, key)
	if len(values) == 0 {
		var zero T
		return zero, false
	}
	return values[0], true
}

// Get fetches values from the cache stored under index, using keys generated from given key parts.
// Note that each number of key parts MUST match expected number and types of the given index name.
func (c *Cache[T]) Get(index string, keys ...[]any) []T {
	return c.GetBy(c.Index(index), keys...)
}

// GetBy fetches values from the cache stored under index, using precalculated index keys.
func (c *Cache[T]) GetBy(index *Index[T], keys ...[]any) []T {
	if index == nil {
		panic("no index given")
	}

	// Acquire hasher.
	h := get_hasher()

	// Acquire lock.
	c.mutex.Lock()

	// Check cache init.
	if c.copy == nil {
		c.mutex.Unlock()
		panic("not initialized")
	}

	// Preallocate expected ret slice.
	values := make([]T, 0, len(keys))

	for _, key := range keys {

		// Generate sum from provided key.
		sum, ok := index_hash(index, h, key)
		if !ok {
			continue
		}

		// Get indexed results list at key.
		list := index_get(index, sum, key)
		if list == nil {
			continue
		}

		// Concatenate all *values* from non-err cached results.
		list_rangefn(list, func(e *list_elem) {
			entry := (*index_entry)(e.data)
			res := entry.result

			switch value := res.data.(type) {
			case T:
				// Append value COPY.
				value = c.copy(value)
				values = append(values, value)

			case error:
				// Don't bump
				// for errors.
				return
			}

			// Push to front of LRU list, USING
			// THE RESULT'S LRU ENTRY, NOT THE
			// INDEX KEY ENTRY. VERY IMPORTANT!!
			list_move_front(&c.lruList, &res.elem)
		})
	}

	// Done with lock.
	c.mutex.Unlock()

	// Done with h.
	hash_pool.Put(h)

	return values
}

// Put will insert the given values into cache,
// calling any invalidate hook on each value.
func (c *Cache[T]) Put(values ...T) {
	var z Hash

	// Acquire lock.
	c.mutex.Lock()

	// Get func ptrs.
	invalid := c.invalid

	// Check cache init.
	if c.copy == nil {
		c.mutex.Unlock()
		panic("not initialized")
	}

	// Store all the passed values.
	for _, value := range values {
		c.store_value(nil, z, nil, value)
	}

	// Done with lock.
	c.mutex.Unlock()

	if invalid != nil {
		// Pass all invalidated values
		// to given user hook (if set).
		for _, value := range values {
			invalid(value)
		}
	}
}

// LoadOne fetches one result from the cache stored under index, using key generated from key parts.
// In the case that no result is found, the provided load callback will be used to hydrate the cache.
// Note that given number of key parts MUST match expected number and types of the given index name.
func (c *Cache[T]) LoadOne(index string, load func() (T, error), key ...any) (T, error) {
	return c.LoadOneBy(c.Index(index), load, key...)
}

// LoadOneBy fetches one result from the cache stored under index, using precalculated index key.
// In the case that no result is found, provided load callback will be used to hydrate the cache.
func (c *Cache[T]) LoadOneBy(index *Index[T], load func() (T, error), key ...any) (T, error) {
	if index == nil {
		panic("no index given")
	} else if !is_unique(index.flags) {
		panic("cannot get one by non-unique index")
	}

	var (
		// whether a result was found
		// (and so val / err are set).
		ok bool

		// separate value / error ptrs
		// as the result is liable to
		// change outside of lock.
		val T
		err error
	)

	// Acquire hasher.
	h := get_hasher()

	// Generate sum from provided key.
	sum, _ := index_hash(index, h, key)

	// Done with h.
	hash_pool.Put(h)

	// Acquire lock.
	c.mutex.Lock()

	// Get func ptrs.
	ignore := c.ignore

	// Check init'd.
	if c.copy == nil ||
		ignore == nil {
		c.mutex.Unlock()
		panic("not initialized")
	}

	// Get indexed list at hash key.
	list := index_get(index, sum, key)

	if ok = (list != nil); ok {
		entry := (*index_entry)(list.head.data)
		res := entry.result

		switch data := res.data.(type) {
		case T:
			// Return value COPY.
			val = c.copy(data)
		case error:
			// Return error.
			err = data
		}

		// Push to front of LRU list, USING
		// THE RESULT'S LRU ENTRY, NOT THE
		// INDEX KEY ENTRY. VERY IMPORTANT!!
		list_move_front(&c.lruList, &res.elem)
	}

	// Done with lock.
	c.mutex.Unlock()

	if ok {
		// result found!
		return val, err
	}

	// Load new result.
	val, err = load()

	// Check for ignored
	// (transient) errors.
	if ignore(err) {
		return val, err
	}

	// Acquire lock.
	c.mutex.Lock()

	// Index this new loaded result.
	// Note this handles copying of
	// the provided value, so it is
	// safe for us to return as-is.
	if err != nil {
		c.store_error(index, sum, key, err)
	} else {
		c.store_value(index, sum, key, val)
	}

	// Done with lock.
	c.mutex.Unlock()

	return val, err
}

// Load fetches values from the cache stored under index, using keys generated from given key parts. The provided get callback is used
// to load groups of values from the cache by the key generated from the key parts provided to the inner callback func, where the returned
// boolean indicates whether any values are currently stored. After the get callback has returned, the cache will then call provided load
// callback to hydrate the cache with any other values. Example usage here is that you may see which values are cached using 'get', and load
// the remaining uncached values using 'load', to minimize database queries. Cached error results are not included or returned by this func.
// Note that given number of key parts MUST match expected number and types of the given index name, in those provided to the get callback.
func (c *Cache[T]) Load(index string, get func(load func(key ...any) bool), load func() ([]T, error)) (values []T, err error) {
	return c.LoadBy(c.Index(index), get, load)
}

// LoadBy fetches values from the cache stored under index, using precalculated index key. The provided get callback is used to load
// groups of values from the cache by the key generated from the key parts provided to the inner callback func, where the returned boolea
// indicates whether any values are currently stored. After the get callback has returned, the cache will then call provided load callback
// to hydrate the cache with any other values. Example usage here is that you may see which values are cached using 'get', and load the
// remaining uncached values using 'load', to minimize database queries. Cached error results are not included or returned by this func.
// Note that given number of key parts MUST match expected number and types of the given index name, in those provided to the get callback.
func (c *Cache[T]) LoadBy(index *Index[T], get func(load func(key ...any) bool), load func() ([]T, error)) (values []T, err error) {
	if index == nil {
		panic("no index given")
	}

	// Acquire hasher.
	h := get_hasher()

	// Acquire lock.
	c.mutex.Lock()

	// Check init'd.
	if c.copy == nil {
		c.mutex.Unlock()
		panic("not initialized")
	}

	var unlocked bool
	defer func() {
		// Deferred unlock to catch
		// any user function panics.
		if !unlocked {
			c.mutex.Unlock()
		}
	}()

	// Pass loader to user func.
	get(func(key ...any) bool {

		// Generate sum from provided key.
		sum, ok := index_hash(index, h, key)
		if !ok {
			return false
		}

		// Get indexed results at hash key.
		list := index_get(index, sum, key)
		if list == nil {
			return false
		}

		// Value length before
		// any below appends.
		before := len(values)

		// Concatenate all *values* from non-err cached results.
		list_rangefn(list, func(e *list_elem) {
			entry := (*index_entry)(e.data)
			res := entry.result

			switch value := res.data.(type) {
			case T:
				// Append value COPY.
				value = c.copy(value)
				values = append(values, value)

			case error:
				// Don't bump
				// for errors.
				return
			}

			// Push to front of LRU list, USING
			// THE RESULT'S LRU ENTRY, NOT THE
			// INDEX KEY ENTRY. VERY IMPORTANT!!
			list_move_front(&c.lruList, &res.elem)
		})

		// Only if values changed did
		// we actually find anything.
		return len(values) != before
	})

	// Done with lock.
	c.mutex.Unlock()
	unlocked = true

	// Done with h.
	hash_pool.Put(h)

	// Load uncached values.
	uncached, err := load()
	if err != nil {
		return nil, err
	}

	// Insert uncached.
	c.Put(uncached...)

	// Append uncached to return values.
	values = append(values, uncached...)

	return
}

// Store will call the given store callback, on non-error then
// passing the provided value to the Put() function. On error
// return the value is still passed to stored invalidate hook.
func (c *Cache[T]) Store(value T, store func() error) error {
	// Store value.
	err := store()

	if err != nil {
		// Get func ptrs.
		c.mutex.Lock()
		invalid := c.invalid
		c.mutex.Unlock()

		// On error don't store
		// value, but still pass
		// to invalidate hook.
		if invalid != nil {
			invalid(value)
		}

		return err
	}

	// Store value.
	c.Put(value)

	return nil
}

// Invalidate generates index key from parts and invalidates all stored under it.
func (c *Cache[T]) Invalidate(index string, key ...any) {
	c.InvalidateBy(c.Index(index), key...)
}

// InvalidateBy invalidates all results stored under index key.
func (c *Cache[T]) InvalidateBy(index *Index[T], key ...any) {
	if index == nil {
		panic("no index given")
	}

	// Acquire hasher.
	h := get_hasher()

	// Generate sum from provided key.
	sum, ok := index_hash(index, h, key)

	// Done with h.
	hash_pool.Put(h)

	if !ok {
		return
	}

	var values []T

	// Acquire lock.
	c.mutex.Lock()

	// Get func ptrs.
	invalid := c.invalid

	// Delete all results under key from index, collecting
	// value results and dropping them from all their indices.
	index_delete(c, index, sum, key, func(del *result) {
		switch value := del.data.(type) {
		case T:
			// Append value COPY.
			value = c.copy(value)
			values = append(values, value)
		case error:
		}
		c.delete(del)
	})

	// Done with lock.
	c.mutex.Unlock()

	if invalid != nil {
		// Pass all invalidated values
		// to given user hook (if set).
		for _, value := range values {
			invalid(value)
		}
	}
}

// Trim will truncate the cache to ensure it
// stays within given percentage of MaxSize.
func (c *Cache[T]) Trim(perc float64) {
	// Acquire lock.
	c.mutex.Lock()

	// Calculate number of cache items to drop.
	max := (perc / 100) * float64(c.maxSize)
	diff := c.lruList.len - int(max)

	if diff <= 0 {
		// Trim not needed.
		c.mutex.Unlock()
		return
	}

	// Iterate over 'diff' results
	// from back (oldest) of cache.
	for i := 0; i < diff; i++ {

		// Get oldest LRU element.
		oldest := c.lruList.tail

		if oldest == nil {
			// reached end.
			break
		}

		// Drop oldest from cache.
		res := (*result)(oldest.data)
		c.delete(res)
	}

	// Done with lock.
	c.mutex.Unlock()
}

// Clear empties the cache by calling .Trim(0).
func (c *Cache[T]) Clear() { c.Trim(0) }

// Len returns the current length of cache.
func (c *Cache[T]) Len() int {
	c.mutex.Lock()
	l := c.lruList.len
	c.mutex.Unlock()
	return l
}

// Cap returns the maximum capacity (size) of cache.
func (c *Cache[T]) Cap() int {
	c.mutex.Lock()
	m := c.maxSize
	c.mutex.Unlock()
	return m
}

func (c *Cache[T]) store_value(index *Index[T], hash Hash, key []any, value T) {
	// Acquire new result.
	res := result_acquire(c)

	if index != nil {
		// Append result to the provided index
		// with precalculated key / its hash.
		index_append(c, index, hash, key, res)
	}

	// Create COPY of value.
	value = c.copy(value)
	res.data = value

	// Acquire hasher.
	h := get_hasher()

	for i := range c.indices {
		// Get current index ptr.
		idx := &(c.indices[i])

		if idx == index {
			// Already stored under
			// this index, ignore.
			continue
		}

		// Get key and hash sum for this index.
		key, sum, ok := index_key(idx, h, value)
		if !ok {
			continue
		}

		// Append result to index at key.
		index_append(c, idx, sum, key, res)
	}

	// Done with h.
	hash_pool.Put(h)

	if c.lruList.len > c.maxSize {
		// Cache has hit max size!
		// Drop the oldest element.
		ptr := c.lruList.tail.data
		res := (*result)(ptr)
		c.delete(res)
	}
}

func (c *Cache[T]) store_error(index *Index[T], hash Hash, key []any, err error) {
	if index == nil {
		// nothing we
		// can do here.
		return
	}

	// Acquire new result.
	res := result_acquire(c)
	res.data = err

	// Append result to the provided index
	// with precalculated key / its hash.
	index_append(c, index, hash, key, res)

	if c.lruList.len > c.maxSize {
		// Cache has hit max size!
		// Drop the oldest element.
		ptr := c.lruList.tail.data
		res := (*result)(ptr)
		c.delete(res)
	}
}

// delete will delete the given result from the cache, deleting
// it from all indices it is stored under, and main LRU list.
func (c *Cache[T]) delete(res *result) {
	for len(res.indexed) != 0 {

		// Pop last indexed entry from list.
		entry := res.indexed[len(res.indexed)-1]
		res.indexed = res.indexed[:len(res.indexed)-1]

		// Drop entry from index.
		index_delete_entry(c, entry)

		// Release to memory pool.
		index_entry_release(entry)
	}

	// Release res to pool.
	result_release(c, res)
}
