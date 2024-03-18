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

// GetOne fetches value from cache stored under index, using precalculated index key.
func (c *Cache[T]) GetOne(index *Index[T], key Key) (T, bool) {
	if index == nil {
		panic("no index given")
	} else if !is_unique(index.flags) {
		panic("cannot get one by non-unique index")
	}
	values := c.Get(index, key)
	if len(values) == 0 {
		var zero T
		return zero, false
	}
	return values[0], true
}

// Get fetches values from the cache stored under index, using precalculated index keys.
func (c *Cache[T]) Get(index *Index[T], keys ...Key) []T {
	if index == nil {
		panic("no index given")
	}

	// Preallocate expected ret slice.
	values := make([]T, 0, len(keys))

	// Acquire lock.
	c.mutex.Lock()

	// Check cache init.
	if c.copy == nil {
		c.mutex.Unlock()
		panic("not initialized")
	}

	for i := range keys {

		// Get indexed results list at key.
		list := index_get(index, keys[i])
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

	return values
}

// Put will insert the given values into cache,
// calling any invalidate hook on each value.
func (c *Cache[T]) Put(values ...T) {
	// Acquire lock.
	c.mutex.Lock()

	// Get func ptrs.
	invalid := c.invalid

	// Check cache init.
	if c.copy == nil {
		c.mutex.Unlock()
		panic("not initialized")
	}

	// Store all passed values.
	for i := range values {
		c.store_value(nil, Key{}, values[i])
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

// LoadOneBy fetches one result from the cache stored under index, using precalculated index key.
// In the case that no result is found, provided load callback will be used to hydrate the cache.
func (c *Cache[T]) LoadOne(index *Index[T], key Key, load func() (T, error)) (T, error) {
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

	// Get indexed result list at key.
	list := index_get(index, key)

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
		c.store_error(index, key, err)
	} else {
		c.store_value(index, key, val)
	}

	// Done with lock.
	c.mutex.Unlock()

	return val, err
}

// Load fetches values from the cache stored under index, using precalculated index keys. The cache will attempt to
// results with values stored under keys, passing keys with uncached results to the provider load callback to further
// hydrate the cache with missing results. Cached error results not included or returned by this function.
func (c *Cache[T]) Load(index *Index[T], keys []Key, load func([]Key) ([]T, error)) ([]T, error) {
	if index == nil {
		panic("no index given")
	}

	// Preallocate expected ret slice.
	values := make([]T, 0, len(keys))

	// Acquire lock.
	c.mutex.Lock()

	// Check init'd.
	if c.copy == nil {
		c.mutex.Unlock()
		panic("not initialized")
	}

	for i := 0; i < len(keys); i++ {

		// Get indexed results at key.
		list := index_get(index, keys[i])
		if list == nil {
			continue
		}

		// Value length before
		// any below appends.
		before := len(values)

		// Concat all *values* from cached results.
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
		if len(values) != before {

			// We found values at key,
			// drop key from the slice.
			copy(keys[i:], keys[i+1:])
			keys = keys[:len(keys)-1]
		}
	}

	// Done with lock.
	c.mutex.Unlock()

	// Load uncached values.
	uncached, err := load(keys)
	if err != nil {
		return nil, err
	}

	// Insert uncached.
	c.Put(uncached...)

	// Append uncached to return values.
	values = append(values, uncached...)

	return values, nil
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

// Invalidate invalidates all results stored under index keys.
func (c *Cache[T]) Invalidate(index *Index[T], keys ...Key) {
	if index == nil {
		panic("no index given")
	}

	// Acquire lock.
	c.mutex.Lock()

	// Preallocate expected ret slice.
	values := make([]T, 0, len(keys))

	for i := range keys {
		// Delete all results under key from index, collecting
		// value results and dropping them from all their indices.
		index_delete(index, keys[i], func(del *result) {
			switch value := del.data.(type) {
			case T:
				// Append value COPY.
				value = c.copy(value)
				values = append(values, value)
			case error:
			}
			c.delete(del)
		})
	}

	// Get func ptrs.
	invalid := c.invalid

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

func (c *Cache[T]) store_value(index *Index[T], key Key, value T) {
	// Acquire new result.
	res := result_acquire(c)

	if index != nil {
		// Append result to the provided index
		// with precalculated key / its hash.
		index_append(c, index, key, res)
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

		// Extract struct fields comprising
		// key parts configured for this index.
		parts := extract_fields(value, idx.fields)

		// Calculate key for this index.
		key := index_key(idx, h, parts)
		if key.Zero() {
			continue
		}

		// Append result to index at key.
		index_append(c, idx, key, res)
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

func (c *Cache[T]) store_error(index *Index[T], key Key, err error) {
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
	index_append(c, index, key, res)

	if c.lruList.len > c.maxSize {
		// Cache has hit max size!
		// Drop the oldest element.
		ptr := c.lruList.tail.data
		res := (*result)(ptr)
		c.delete(res)
	}
}

func (c *Cache[T]) delete(res *result) {
	for len(res.indexed) != 0 {

		// Pop last indexed entry from list.
		entry := res.indexed[len(res.indexed)-1]
		res.indexed = res.indexed[:len(res.indexed)-1]

		// Drop entry from index.
		index_delete_entry[T](entry)

		// Release to memory pool.
		index_entry_release(entry)
	}

	// Release res to pool.
	result_release(c, res)
}
