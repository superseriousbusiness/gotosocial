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

package headerfilter

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/headerfilter"
)

// Cache provides a means of caching domains in memory to reduce
// load on an underlying storage mechanism, e.g. a database.
//
// The in-memory domain list is kept up-to-date by means of a passed
// loader function during every call to .Matches(). In the case of
// a nil internal domain list, the loader function is called to hydrate
// the cache with the latest list of domains.
//
// The .Clear() function can be used to invalidate the cache,
// e.g. when an entry is added / deleted from the database.
type Cache struct {
	ptr atomic.Pointer[headerfilter.Filters]
}

// Allow checks whether header matches positively against filter in the cache. If cache
// is not currently loaded then the provided load function is called to hydrate it first.
func (c *Cache) Allow(h http.Header, load func() ([]*gtsmodel.HeaderFilter, error)) (bool, error) {
	// Load ptr value.
	ptr := c.ptr.Load()

	if ptr == nil {
		// Cache is not hydrated.
		// Load filters from callback.
		filters, err := loadFilters(load)
		if err != nil {
			return false, err
		}

		// Store the new
		// header filters.
		ptr = &filters
		c.ptr.Store(ptr)
	}

	// Deref and perform match.
	return ptr.Allow(h), nil
}

// Block checks whether header matches negatively against filter in the cache. If cache
// is not currently loaded then the provided load function is called to hydrate it first.
func (c *Cache) Block(h http.Header, load func() ([]*gtsmodel.HeaderFilter, error)) (bool, error) {
	// Load ptr value.
	ptr := c.ptr.Load()

	if ptr == nil {
		// Cache is not hydrated.
		// Load filters from callback.
		filters, err := loadFilters(load)
		if err != nil {
			return false, err
		}

		// Store the new
		// header filters.
		ptr = &filters
		c.ptr.Store(ptr)
	}

	// Deref and perform match.
	return ptr.Block(h), nil
}

// Stats returns match statistics associated with currently cached header filters.
func (c *Cache) Stats() map[string]map[string]uint64 {
	if ptr := c.ptr.Load(); ptr != nil {
		return ptr.Stats()
	}
	return nil
}

// loadFilters will load filters from given load callback, creating and parsing raw filters.
func loadFilters(load func() ([]*gtsmodel.HeaderFilter, error)) (headerfilter.Filters, error) {
	// Load filters from callback.
	hdrFilters, err := load()
	if err != nil {
		return nil, fmt.Errorf("error reloading cache: %w", err)
	}

	// Allocate new header filter slice to store expressions.
	filters := make(headerfilter.Filters, 0, len(hdrFilters))

	// Add all raw expression to filter slice.
	for _, filter := range hdrFilters {
		if err := filters.Append(filter.Header, filter.Regex); err != nil {
			return nil, fmt.Errorf("error appending exprs: %w", err)
		}
	}

	return filters, nil
}

// Clear will drop the currently loaded filters,
// triggering a reload on next call to .Matches().
func (c *Cache) Clear() { c.ptr.Store(nil) }
