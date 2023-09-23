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
	"codeberg.org/gruf/go-cache/v3/simple"
	"golang.org/x/exp/slices"
)

// SliceCache wraps a simple.Cache to provide simple loader-callback
// functions for fetching + caching slices of objects (e.g. IDs).
type SliceCache[T any] struct {
	*simple.Cache[string, []T]
}

// Load will attempt to load an existing slice from the cache for the given key, else calling the provided load function and caching the result.
func (c *SliceCache[T]) Load(key string, load func() ([]T, error)) ([]T, error) {
	// Look for follow IDs list in cache under this key.
	data, ok := c.Get(key)

	if !ok {
		var err error

		// Not cached, load!
		data, err = load()
		if err != nil {
			return nil, err
		}

		// Store the data.
		c.Set(key, data)
	}

	// Return data clone for safety.
	return slices.Clone(data), nil
}
