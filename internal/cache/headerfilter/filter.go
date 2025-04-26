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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/headerfilter"
)

// Cache provides a means of caching headerfilter.Filters in
// memory to reduce load on an underlying storage mechanism.
type Cache struct {
	// current cached header filters slice.
	ptr atomic.Pointer[headerfilter.Filters]
}

// RegularMatch performs .RegularMatch() on cached headerfilter.Filters, loading using callback if necessary.
func (c *Cache) RegularMatch(h http.Header, load func() ([]*gtsmodel.HeaderFilter, error)) (string, string, error) {
	// Load ptr value.
	ptr := c.ptr.Load()

	if ptr == nil {
		// Cache is not hydrated.
		// Load filters from callback.
		filters, err := loadFilters(load)
		if err != nil {
			return "", "", err
		}

		// Store the new
		// header filters.
		ptr = &filters
		c.ptr.Store(ptr)
	}

	// Deref and perform match.
	return ptr.RegularMatch(h)
}

// InverseMatch performs .InverseMatch() on cached headerfilter.Filters, loading using callback if necessary.
func (c *Cache) InverseMatch(h http.Header, load func() ([]*gtsmodel.HeaderFilter, error)) (string, string, error) {
	// Load ptr value.
	ptr := c.ptr.Load()

	if ptr == nil {
		// Cache is not hydrated.
		// Load filters from callback.
		filters, err := loadFilters(load)
		if err != nil {
			return "", "", err
		}

		// Store the new
		// header filters.
		ptr = &filters
		c.ptr.Store(ptr)
	}

	// Deref and perform match.
	return ptr.InverseMatch(h)
}

// Clear will drop the currently loaded filters,
// triggering a reload on next call to ._Match().
func (c *Cache) Clear() { c.ptr.Store(nil) }

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
