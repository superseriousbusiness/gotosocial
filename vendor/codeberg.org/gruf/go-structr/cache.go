package structr

import (
	"context"
	"errors"
	"reflect"
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
	lruList list[*result[StructType]]

	// memory pools of common types.
	llsPool []*list[*result[StructType]]
	resPool []*result[StructType]
	keyPool []*indexkey[StructType]

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
	// - Cache{} pools
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
	for i, config := range config.Indices {
		c.indices[i].init(config)
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
func (c *Cache[T]) GetOne(index string, keyParts ...any) (T, bool) {
	// Get index with name.
	idx := c.Index(index)

	// Generate index key from provided parts.
	key, ok := idx.keygen.FromParts(keyParts...)
	if !ok {
		var zero T
		return zero, false
	}

	// Fetch one value for key.
	return c.GetOneBy(idx, key)
}

// GetOneBy fetches value from cache stored under index, using precalculated index key.
func (c *Cache[T]) GetOneBy(index *Index[T], key string) (T, bool) {
	if index == nil {
		panic("no index given")
	} else if !index.unique {
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
func (c *Cache[T]) Get(index string, keysParts ...[]any) []T {
	// Get index with name.
	idx := c.Index(index)

	// Preallocate expected keys slice length.
	keys := make([]string, 0, len(keysParts))

	// Acquire buf.
	buf := getBuf()

	for _, parts := range keysParts {
		// Reset buf.
		buf.Reset()

		// Generate key from provided parts into buffer.
		if !idx.keygen.AppendFromParts(buf, parts...) {
			continue
		}

		// Get string copy of
		// genarated idx key.
		key := string(buf.B)

		// Append key to keys.
		keys = append(keys, key)
	}

	// Done with buf.
	putBuf(buf)

	// Continue fetching values.
	return c.GetBy(idx, keys...)
}

// GetBy fetches values from the cache stored under index, using precalculated index keys.
func (c *Cache[T]) GetBy(index *Index[T], keys ...string) []T {
	if index == nil {
		panic("no index given")
	}

	// Preallocate a slice of est. len.
	values := make([]T, 0, len(keys))

	// Acquire lock.
	c.mutex.Lock()

	// Check cache init.
	if c.copy == nil {
		c.mutex.Unlock()
		panic("not initialized")
	}

	// Check index for all keys.
	for _, key := range keys {

		// Get indexed results.
		list := index.data[key]

		if list != nil {
			// Concatenate all results with values.
			list.rangefn(func(e *elem[*result[T]]) {
				if e.Value.err != nil {
					return
				}

				// Append a copy of value.
				value := c.copy(e.Value.value)
				values = append(values, value)

				// Push to front of LRU list, USING
				// THE RESULT'S LRU ENTRY, NOT THE
				// INDEX KEY ENTRY. VERY IMPORTANT!!
				c.lruList.moveFront(&e.Value.entry)
			})
		}
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

	// Store all the passed values.
	for _, value := range values {
		c.store(nil, "", value, nil)
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
func (c *Cache[T]) LoadOne(index string, load func() (T, error), keyParts ...any) (T, error) {
	// Get index with name.
	idx := c.Index(index)

	// Generate cache from from provided parts.
	key, _ := idx.keygen.FromParts(keyParts...)

	// Continue loading this result.
	return c.LoadOneBy(idx, load, key)
}

// LoadOneBy fetches one result from the cache stored under index, using precalculated index key.
// In the case that no result is found, provided load callback will be used to hydrate the cache.
func (c *Cache[T]) LoadOneBy(index *Index[T], load func() (T, error), key string) (T, error) {
	if index == nil {
		panic("no index given")
	} else if !index.unique {
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

	// Get indexed results.
	list := index.data[key]

	if ok = (list != nil && list.head != nil); ok {
		e := list.head

		// Extract val / err.
		val = e.Value.value
		err = e.Value.err

		if err == nil {
			// We only ever ret
			// a COPY of value.
			val = c.copy(val)
		}

		// Push to front of LRU list, USING
		// THE RESULT'S LRU ENTRY, NOT THE
		// INDEX KEY ENTRY. VERY IMPORTANT!!
		c.lruList.moveFront(&e.Value.entry)
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
	c.store(index, key, val, err)

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
func (c *Cache[T]) Load(index string, get func(load func(keyParts ...any) bool), load func() ([]T, error)) (values []T, err error) {
	return c.LoadBy(c.Index(index), get, load)
}

// LoadBy fetches values from the cache stored under index, using precalculated index key. The provided get callback is used to load
// groups of values from the cache by the key generated from the key parts provided to the inner callback func, where the returned boolea
// indicates whether any values are currently stored. After the get callback has returned, the cache will then call provided load callback
// to hydrate the cache with any other values. Example usage here is that you may see which values are cached using 'get', and load the
// remaining uncached values using 'load', to minimize database queries. Cached error results are not included or returned by this func.
// Note that given number of key parts MUST match expected number and types of the given index name, in those provided to the get callback.
func (c *Cache[T]) LoadBy(index *Index[T], get func(load func(keyParts ...any) bool), load func() ([]T, error)) (values []T, err error) {
	if index == nil {
		panic("no index given")
	}

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

	// Acquire buf.
	buf := getBuf()

	// Pass cache check to user func.
	get(func(keyParts ...any) bool {

		// Reset buf.
		buf.Reset()

		// Generate index key from provided key parts.
		if !index.keygen.AppendFromParts(buf, keyParts...) {
			return false
		}

		// Get temp generated key str,
		// (not needed after return).
		keyStr := buf.String()

		// Get all indexed results.
		list := index.data[keyStr]

		if list != nil && list.len > 0 {
			// Value length before
			// any below appends.
			before := len(values)

			// Concatenate all results with values.
			list.rangefn(func(e *elem[*result[T]]) {
				if e.Value.err != nil {
					return
				}

				// Append a copy of value.
				value := c.copy(e.Value.value)
				values = append(values, value)

				// Push to front of LRU list, USING
				// THE RESULT'S LRU ENTRY, NOT THE
				// INDEX KEY ENTRY. VERY IMPORTANT!!
				c.lruList.moveFront(&e.Value.entry)
			})

			// Only if values changed did
			// we actually find anything.
			return len(values) != before
		}

		return false
	})

	// Done with buf.
	putBuf(buf)

	// Done with lock.
	c.mutex.Unlock()
	unlocked = true

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
func (c *Cache[T]) Invalidate(index string, keyParts ...any) {
	// Get index with name.
	idx := c.Index(index)

	// Generate cache from from provided parts.
	key, ok := idx.keygen.FromParts(keyParts...)
	if !ok {
		return
	}

	// Continue invalidation.
	c.InvalidateBy(idx, key)
}

// InvalidateBy invalidates all results stored under index key.
func (c *Cache[T]) InvalidateBy(index *Index[T], key string) {
	if index == nil {
		panic("no index given")
	}

	var values []T

	// Acquire lock.
	c.mutex.Lock()

	// Get func ptrs.
	invalid := c.invalid

	// Delete all results under key from index, collecting
	// value results and dropping them from all their indices.
	index_delete(c, index, key, func(del *result[T]) {
		if del.err == nil {
			values = append(values, del.value)
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
		c.delete(oldest.Value)
	}

	// Done with lock.
	c.mutex.Unlock()
}

// Clear empties the cache by calling .Trim(0).
func (c *Cache[T]) Clear() { c.Trim(0) }

// Clean drops unused items from its memory pools.
// Useful to free memory if cache has downsized.
func (c *Cache[T]) Clean() {
	c.mutex.Lock()
	c.llsPool = nil
	c.resPool = nil
	c.keyPool = nil
	c.mutex.Unlock()
}

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

// store will store the given value / error result in the cache, storing it under the
// already provided index + key if provided, else generating keys from provided value.
func (c *Cache[T]) store(index *Index[T], key string, value T, err error) {
	// Acquire new result.
	res := result_acquire(c)

	if index != nil {
		// Append result to the provided
		// precalculated key and its index.
		index_append(c, index, key, res)

	} else if err != nil {

		// This is an error result without
		// an index provided, nothing we
		// can do here so release result.
		result_release(c, res)
		return
	}

	// Set and check the result error.
	if res.err = err; res.err == nil {

		// This is value result, we need to
		// store it under all other indices
		// other than the provided.
		//
		// Create COPY of value.
		res.value = c.copy(value)

		// Get reflected value of incoming
		// value, used during cache key gen.
		rvalue := reflect.ValueOf(value)

		// Acquire buf.
		buf := getBuf()

		for i := range c.indices {
			// Get current index ptr.
			idx := &(c.indices[i])

			if idx == index {
				// Already stored under
				// this index, ignore.
				continue
			}

			// Generate key from reflect value,
			// (this ignores zero value keys).
			buf.Reset() // reset buf first
			if !idx.keygen.appendFromRValue(buf, rvalue) {
				continue
			}

			// Alloc key copy.
			key := string(buf.B)

			// Append result to index at key.
			index_append(c, idx, key, res)
		}

		// Done with buf.
		putBuf(buf)
	}

	if c.lruList.len > c.maxSize {
		// Cache has hit max size!
		// Drop the oldest element.
		res := c.lruList.tail.Value
		c.delete(res)
	}
}

// delete will delete the given result from the cache, deleting
// it from all indices it is stored under, and main LRU list.
func (c *Cache[T]) delete(res *result[T]) {
	for len(res.keys) != 0 {

		// Pop indexkey at end of list.
		ikey := res.keys[len(res.keys)-1]
		res.keys = res.keys[:len(res.keys)-1]

		// Drop this result from list at key.
		index_deleteOne(c, ikey.index, ikey)

		// Release ikey to pool.
		indexkey_release(c, ikey)
	}

	// Release res to pool.
	result_release(c, res)
}
