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
	"codeberg.org/gruf/go-cache/v3/result"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

type VisibilityCache struct {
	*result.Cache[*CachedVisibility]
}

// Init will initialize the visibility cache in this collection.
// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
func (c *VisibilityCache) Init() {
	c.Cache = result.New([]result.Lookup{
		{Name: "ItemID"},
		{Name: "RequesterID"},
		{Name: "Type.RequesterID.ItemID"},
	}, func(v1 *CachedVisibility) *CachedVisibility {
		v2 := new(CachedVisibility)
		*v2 = *v1
		return v2
	}, config.GetCacheVisibilityMaxSize())
	c.Cache.SetTTL(config.GetCacheVisibilityTTL(), true)
	c.Cache.IgnoreErrors(ignoreErrors)
}

// Start will attempt to start the visibility cache, or panic.
func (c *VisibilityCache) Start() {
	tryStart(c.Cache, config.GetCacheVisibilitySweepFreq())
}

// Stop will attempt to stop the visibility cache, or panic.
func (c *VisibilityCache) Stop() {
	tryStop(c.Cache, config.GetCacheVisibilitySweepFreq())
}

// VisibilityType represents a visibility lookup type.
// We use a byte type here to improve performance in the
// result cache when generating the key.
type VisibilityType byte

const (
	// Possible cache visibility lookup types.
	VisibilityTypeAccount = VisibilityType('a')
	VisibilityTypeStatus  = VisibilityType('s')
	VisibilityTypeHome    = VisibilityType('h')
	VisibilityTypePublic  = VisibilityType('p')
)

// CachedVisibility represents a cached visibility lookup value.
type CachedVisibility struct {
	// ItemID is the ID of the item in question (status / account).
	ItemID string

	// RequesterID is the ID of the requesting account for this visibility lookup.
	RequesterID string

	// Type is the visibility lookup type.
	Type VisibilityType

	// Value is the actual visibility value.
	Value bool
}
