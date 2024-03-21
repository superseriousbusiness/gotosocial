// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cache

import (
	"slices"

	"codeberg.org/gruf/go-cache/v3/simple"
	"codeberg.org/gruf/go-structr"
)

// SliceCache wraps a simple.Cache to provide simple loader-callback
// functions for fetching + caching slices of objects (e.g. IDs).
type SliceCache[T any] struct {
	cache simple.Cache[string, []T]
}

// Init ...
func (c *SliceCache[T]) Init(len, cap int) {
	c.cache = simple.Cache[string, []T]{}
	c.cache.Init(len, cap)
}

// Load will attempt to load an existing slice from the cache for the given key, else calling the provided load function and caching the result.
func (c *SliceCache[T]) Load(key string, load func() ([]T, error)) ([]T, error) {
	// Look for cached values.
	data, ok := c.cache.Get(key)

	if !ok {
		var err error

		// Not cached, load!
		data, err = load()
		if err != nil {
			return nil, err
		}

		// Store the data.
		c.cache.Set(key, data)
	}

	// Return data clone for safety.
	return slices.Clone(data), nil
}

// Invalidate ...
func (c *SliceCache[T]) Invalidate(keys ...string) {
	_ = c.cache.InvalidateAll(keys...)
}

// Trim ...
func (c *SliceCache[T]) Trim(perc float64) {
	c.cache.Trim(perc)
}

// Clear ...
func (c *SliceCache[T]) Clear() {
	c.cache.Clear()
}

// Len ...
func (c *SliceCache[T]) Len() int {
	return c.cache.Len()
}

// Cap ...
func (c *SliceCache[T]) Cap() int {
	return c.cache.Cap()
}

// StructCache ...
type StructCache[StructType any] struct {
	cache structr.Cache[StructType]
	index map[string]*structr.Index
}

// Init ...
func (c *StructCache[T]) Init(config structr.CacheConfig[T]) {
	c.index = make(map[string]*structr.Index, len(config.Indices))
	c.cache = structr.Cache[T]{}
	c.cache.Init(config)
	for _, cfg := range config.Indices {
		c.index[cfg.Fields] = c.cache.Index(cfg.Fields)
	}
}

// GetOne ...
func (c *StructCache[T]) GetOne(index string, key ...any) (T, bool) {
	i := c.index[index]
	return c.cache.GetOne(i, i.Key(key...))
}

// Get ...
func (c *StructCache[T]) Get(index string, keys ...[]any) []T {
	i := c.index[index]
	return c.cache.Get(i, i.Keys(keys...)...)
}

// Put ...
func (c *StructCache[T]) Put(values ...T) {
	c.cache.Put(values...)
}

// LoadOne ...
func (c *StructCache[T]) LoadOne(index string, load func() (T, error), key ...any) (T, error) {
	i := c.index[index]
	return c.cache.LoadOne(i, i.Key(key...), load)
}

// LoadIDs ...
func (c *StructCache[T]) LoadIDs(index string, ids []string, load func([]string) ([]T, error)) ([]T, error) {
	i := c.index[index]
	if i == nil {
		// we only perform this check here as
		// we're going to use the index before
		// passing it to cache in main .Load().
		panic("missing index for cache type")
	}

	// Generate cache keys for ID types.
	keys := make([]structr.Key, len(ids))
	for x, id := range ids {
		keys[x] = i.Key(id)
	}

	// Pass loader callback with wrapper onto main cache load function.
	return c.cache.Load(i, keys, func(uncached []structr.Key) ([]T, error) {
		uncachedIDs := make([]string, len(uncached))
		for i := range uncached {
			uncachedIDs[i] = uncached[i].Values()[0].(string)
		}
		return load(uncachedIDs)
	})
}

// Store ...
func (c *StructCache[T]) Store(value T, store func() error) error {
	return c.cache.Store(value, store)
}

// InvalidateOne ...
func (c *StructCache[T]) Invalidate(index string, key ...any) {
	i := c.index[index]
	c.cache.Invalidate(i, i.Key(key...))
}

// InvalidateIDs ...
func (c *StructCache[T]) InvalidateIDs(index string, ids []string) {
	i := c.index[index]
	if i == nil {
		// we only perform this check here as
		// we're going to use the index before
		// passing it to cache in main .Load().
		panic("missing index for cache type")
	}

	// Generate cache keys for ID types.
	keys := make([]structr.Key, len(ids))
	for x, id := range ids {
		keys[x] = i.Key(id)
	}

	// Pass to main invalidate func.
	c.cache.Invalidate(i, keys...)
}

// Trim ...
func (c *StructCache[T]) Trim(perc float64) {
	c.cache.Trim(perc)
}

// Clear ...
func (c *StructCache[T]) Clear() {
	c.cache.Clear()
}

// Len ...
func (c *StructCache[T]) Len() int {
	return c.cache.Len()
}

// Cap ...
func (c *StructCache[T]) Cap() int {
	return c.cache.Cap()
}
