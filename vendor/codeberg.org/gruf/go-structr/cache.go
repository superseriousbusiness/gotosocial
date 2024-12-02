package structr

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"unsafe"
)

// DefaultIgnoreErr is the default function used to
// ignore (i.e. not cache) incoming error results during
// Load() calls. By default ignores context pkg errors.
func DefaultIgnoreErr(err error) bool {
	return errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded)
}

// CacheConfig defines config vars
// for initializing a struct cache.
type CacheConfig[StructType any] struct {

	// IgnoreErr defines which errors to
	// ignore (i.e. not cache) returned
	// from load function callback calls.
	// This may be left as nil, on which
	// DefaultIgnoreErr will be used.
	IgnoreErr func(error) bool

	// Copy provides a means of copying
	// cached values, to ensure returned values
	// do not share memory with those in cache.
	Copy func(StructType) StructType

	// Invalidate is called when cache values
	// (NOT errors) are invalidated, either
	// as the values passed to Put() / Store(),
	// or by the keys by calls to Invalidate().
	Invalidate func(StructType)

	// Indices defines indices to create
	// in the Cache for the receiving
	// generic struct type parameter.
	Indices []IndexConfig

	// MaxSize defines the maximum number
	// of items allowed in the Cache at
	// one time, before old items start
	// getting evicted.
	MaxSize int
}

// Cache provides a structure cache with automated
// indexing and lookups by any initialization-defined
// combination of fields. This also supports caching
// of negative results (errors!) returned by LoadOne().
type Cache[StructType any] struct {

	// hook functions.
	ignore  func(error) bool
	copy    func(StructType) StructType
	invalid func(StructType)

	// keeps track of all indexed items,
	// in order of last recently used (LRU).
	lru list

	// indices used in storing passed struct
	// types by user defined sets of fields.
	indices []Index

	// max cache size, imposes size
	// limit on the lruList in order
	// to evict old entries.
	maxSize int

	// protective mutex, guards:
	// - Cache{}.lruList
	// - Index{}.data
	// - Cache{} hook fns
	mutex sync.Mutex
}

// Init initializes the cache with given configuration
// including struct fields to index, and necessary fns.
func (c *Cache[T]) Init(config CacheConfig[T]) {
	t := reflect.TypeOf((*T)(nil)).Elem()

	if len(config.Indices) == 0 {
		panic("no indices provided")
	}

	if config.IgnoreErr == nil {
		config.IgnoreErr = DefaultIgnoreErr
	}

	if config.Copy == nil {
		panic("copy function must be provided")
	}

	if config.MaxSize < 2 {
		panic("minimum cache size is 2 for LRU to work")
	}

	// Safely copy over
	// provided config.
	c.mutex.Lock()
	c.indices = make([]Index, len(config.Indices))
	for i, cfg := range config.Indices {
		c.indices[i].ptr = unsafe.Pointer(c)
		c.indices[i].init(t, cfg, config.MaxSize)
	}
	c.ignore = config.IgnoreErr
	c.copy = config.Copy
	c.invalid = config.Invalidate
	c.maxSize = config.MaxSize
	c.mutex.Unlock()
}

// Index selects index with given name from cache, else panics.
func (c *Cache[T]) Index(name string) *Index {
	for i, idx := range c.indices {
		if idx.name == name {
			return &(c.indices[i])
		}
	}
	panic("unknown index: " + name)
}

// GetOne fetches value from cache stored under index, using precalculated index key.
func (c *Cache[T]) GetOne(index *Index, key Key) (T, bool) {
	values := c.Get(index, key)
	if len(values) == 0 {
		var zero T
		return zero, false
	}
	return values[0], true
}

// Get fetches values from the cache stored under index, using precalculated index keys.
func (c *Cache[T]) Get(index *Index, keys ...Key) []T {
	if index == nil {
		panic("no index given")
	} else if index.ptr != unsafe.Pointer(c) {
		panic("invalid index for cache")
	}

	// Preallocate expected ret slice.
	values := make([]T, 0, len(keys))

	// Acquire lock.
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check cache init.
	if c.copy == nil {
		panic("not initialized")
	}

	for i := range keys {
		// Concatenate all *values* from cached items.
		index.get(keys[i].key, func(item *indexed_item) {
			if value, ok := item.data.(T); ok {
				// Append value COPY.
				value = c.copy(value)
				values = append(values, value)

				// Push to front of LRU list, USING
				// THE ITEM'S LRU ENTRY, NOT THE
				// INDEX KEY ENTRY. VERY IMPORTANT!!
				c.lru.move_front(&item.elem)
			}
		})
	}

	return values
}

// Put will insert the given values into cache,
// calling any invalidate hook on each value.
func (c *Cache[T]) Put(values ...T) {
	// Acquire lock.
	c.mutex.Lock()

	// Wrap unlock to only do once.
	unlock := once(c.mutex.Unlock)
	defer unlock()

	// Check cache init.
	if c.copy == nil {
		panic("not initialized")
	}

	// Store all passed values.
	for i := range values {
		c.store_value(
			nil, "",
			values[i],
		)
	}

	// Get func ptrs.
	invalid := c.invalid

	// Done with
	// the lock.
	unlock()

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
func (c *Cache[T]) LoadOne(index *Index, key Key, load func() (T, error)) (T, error) {
	if index == nil {
		panic("no index given")
	} else if index.ptr != unsafe.Pointer(c) {
		panic("invalid index for cache")
	} else if !is_unique(index.flags) {
		panic("cannot get one by non-unique index")
	}

	var (
		// whether an item was found
		// (and so val / err are set).
		ok bool

		// separate value / error ptrs
		// as the item is liable to
		// change outside of lock.
		val T
		err error
	)

	// Acquire lock.
	c.mutex.Lock()

	// Wrap unlock to only do once.
	unlock := once(c.mutex.Unlock)
	defer unlock()

	// Check init'd.
	if c.copy == nil ||
		c.ignore == nil {
		panic("not initialized")
	}

	// Get item indexed at key.
	item := index.get_one(key)

	if ok = (item != nil); ok {
		var is bool

		if val, is = item.data.(T); is {
			// Set value COPY.
			val = c.copy(val)

			// Push to front of LRU list, USING
			// THE ITEM'S LRU ENTRY, NOT THE
			// INDEX KEY ENTRY. VERY IMPORTANT!!
			c.lru.move_front(&item.elem)

		} else {

			// Attempt to return error.
			err, _ = item.data.(error)
		}
	}

	// Get func ptrs.
	ignore := c.ignore

	// Done with
	// the lock.
	unlock()

	if ok {
		// item found!
		return val, err
	}

	// Load new result.
	val, err = load()

	// Check for ignored error types.
	if err != nil && ignore(err) {
		return val, err
	}

	// Acquire lock.
	c.mutex.Lock()

	// Index this new loaded item.
	// Note this handles copying of
	// the provided value, so it is
	// safe for us to return as-is.
	if err != nil {
		c.store_error(index, key.key, err)
	} else {
		c.store_value(index, key.key, val)
	}

	// Done with lock.
	c.mutex.Unlock()

	return val, err
}

// Load fetches values from the cache stored under index, using precalculated index keys. The cache will attempt to
// results with values stored under keys, passing keys with uncached results to the provider load callback to further
// hydrate the cache with missing results. Cached error results not included or returned by this function.
func (c *Cache[T]) Load(index *Index, keys []Key, load func([]Key) ([]T, error)) ([]T, error) {
	if index == nil {
		panic("no index given")
	} else if index.ptr != unsafe.Pointer(c) {
		panic("invalid index for cache")
	}

	// Preallocate expected ret slice.
	values := make([]T, 0, len(keys))

	// Acquire lock.
	c.mutex.Lock()

	// Wrap unlock to only do once.
	unlock := once(c.mutex.Unlock)
	defer unlock()

	// Check init'd.
	if c.copy == nil {
		panic("not initialized")
	}

	// Iterate keys and catch uncached.
	toLoad := make([]Key, 0, len(keys))
	for _, key := range keys {

		// Value length before
		// any below appends.
		before := len(values)

		// Concatenate all *values* from cached items.
		index.get(key.key, func(item *indexed_item) {
			if value, ok := item.data.(T); ok {
				// Append value COPY.
				value = c.copy(value)
				values = append(values, value)

				// Push to front of LRU list, USING
				// THE ITEM'S LRU ENTRY, NOT THE
				// INDEX KEY ENTRY. VERY IMPORTANT!!
				c.lru.move_front(&item.elem)
			}
		})

		// Only if values changed did
		// we actually find anything.
		if len(values) == before {
			toLoad = append(toLoad, key)
		}
	}

	// Done with
	// the lock.
	unlock()

	if len(toLoad) == 0 {
		// We loaded everything!
		return values, nil
	}

	// Load uncached key values.
	uncached, err := load(toLoad)
	if err != nil {
		return nil, err
	}

	// Acquire lock.
	c.mutex.Lock()

	// Store all uncached values.
	for i := range uncached {
		c.store_value(
			nil, "",
			uncached[i],
		)
	}

	// Done with lock.
	c.mutex.Unlock()

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
func (c *Cache[T]) Invalidate(index *Index, keys ...Key) {
	if index == nil {
		panic("no index given")
	} else if index.ptr != unsafe.Pointer(c) {
		panic("invalid index for cache")
	}

	// Acquire lock.
	c.mutex.Lock()

	// Preallocate expected ret slice.
	values := make([]T, 0, len(keys))

	for i := range keys {
		// Delete all items under key from index, collecting
		// value items and dropping them from all their indices.
		index.delete(keys[i].key, func(item *indexed_item) {

			if value, ok := item.data.(T); ok {
				// No need to copy, as item
				// being deleted from cache.
				values = append(values, value)
			}

			// Delete cached.
			c.delete(item)
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
	diff := c.lru.len - int(max)
	if diff <= 0 {

		// Trim not needed.
		c.mutex.Unlock()
		return
	}

	// Iterate over 'diff' items
	// from back (oldest) of cache.
	for i := 0; i < diff; i++ {

		// Get oldest LRU elem.
		oldest := c.lru.tail
		if oldest == nil {

			// reached
			// end.
			break
		}

		// Drop oldest item from cache.
		item := (*indexed_item)(oldest.data)
		c.delete(item)
	}

	// Compact index data stores.
	for _, idx := range c.indices {
		(&idx).data.Compact()
	}

	// Done with lock.
	c.mutex.Unlock()
}

// Clear empties the cache by calling .Trim(0).
func (c *Cache[T]) Clear() { c.Trim(0) }

// Len returns the current length of cache.
func (c *Cache[T]) Len() int {
	c.mutex.Lock()
	l := c.lru.len
	c.mutex.Unlock()
	return l
}

// Debug returns debug stats about cache.
func (c *Cache[T]) Debug() map[string]any {
	m := make(map[string]any, 2)
	c.mutex.Lock()
	m["lru"] = c.lru.len
	indices := make(map[string]any, len(c.indices))
	m["indices"] = indices
	for _, idx := range c.indices {
		var n uint64
		for _, l := range idx.data.m {
			n += uint64(l.len)
		}
		indices[idx.name] = n
	}
	c.mutex.Unlock()
	return m
}

// Cap returns the maximum capacity (size) of cache.
func (c *Cache[T]) Cap() int {
	c.mutex.Lock()
	m := c.maxSize
	c.mutex.Unlock()
	return m
}

func (c *Cache[T]) store_value(index *Index, key string, value T) {
	// Alloc new index item.
	item := new_indexed_item()
	if cap(item.indexed) < len(c.indices) {

		// Preallocate item indices slice to prevent Go auto
		// allocating overlying large slices we don't need.
		item.indexed = make([]*index_entry, 0, len(c.indices))
	}

	// Create COPY of value.
	value = c.copy(value)
	item.data = value

	if index != nil {
		// Append item to index a key
		// was already generated for.
		index.append(&c.lru, key, item)
	}

	// Get ptr to value data.
	ptr := unsafe.Pointer(&value)

	// Acquire key buf.
	buf := new_buffer()

	for i := range c.indices {
		// Get current index ptr.
		idx := (&c.indices[i])
		if idx == index {

			// Already stored under
			// this index, ignore.
			continue
		}

		// Extract fields comprising index key.
		parts := extract_fields(ptr, idx.fields)
		if parts == nil {
			continue
		}

		// Calculate index key.
		key := idx.key(buf, parts)
		if key == "" {
			continue
		}

		// Append item to this index.
		idx.append(&c.lru, key, item)
	}

	// Add item to main lru list.
	c.lru.push_front(&item.elem)

	// Done with buf.
	free_buffer(buf)

	if c.lru.len > c.maxSize {
		// Cache has hit max size!
		// Drop the oldest element.
		ptr := c.lru.tail.data
		item := (*indexed_item)(ptr)
		c.delete(item)
	}
}

func (c *Cache[T]) store_error(index *Index, key string, err error) {
	if index == nil {
		// nothing we
		// can do here.
		return
	}

	// Alloc new index item.
	item := new_indexed_item()
	if cap(item.indexed) < len(c.indices) {

		// Preallocate item indices slice to prevent Go auto
		// allocating overlying large slices we don't need.
		item.indexed = make([]*index_entry, 0, len(c.indices))
	}

	// Set error val.
	item.data = err

	// Append item to index a key
	// was already generated for.
	index.append(&c.lru, key, item)

	// Add item to main lru list.
	c.lru.push_front(&item.elem)

	if c.lru.len > c.maxSize {
		// Cache has hit max size!
		// Drop the oldest element.
		ptr := c.lru.tail.data
		item := (*indexed_item)(ptr)
		c.delete(item)
	}
}

func (c *Cache[T]) delete(item *indexed_item) {
	for len(item.indexed) != 0 {
		// Pop last indexed entry from list.
		entry := item.indexed[len(item.indexed)-1]
		item.indexed = item.indexed[:len(item.indexed)-1]

		// Drop index_entry from index.
		entry.index.delete_entry(entry)
	}

	// Drop entry from lru list.
	c.lru.remove(&item.elem)

	// Free now-unused item.
	free_indexed_item(item)
}
