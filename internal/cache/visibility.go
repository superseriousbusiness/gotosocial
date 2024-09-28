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
	"codeberg.org/gruf/go-structr"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type VisibilityCache struct {
	StructCache[*CachedVisibility]
}

func (c *Caches) initVisibility() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofVisibility(), // model in-mem size.
		config.GetCacheVisibilityMemRatio(),
	)

	log.Infof(nil, "Visibility cache size = %d", cap)

	copyF := func(v1 *CachedVisibility) *CachedVisibility {
		v2 := new(CachedVisibility)
		*v2 = *v1
		return v2
	}

	c.Visibility.Init(structr.CacheConfig[*CachedVisibility]{
		Indices: []structr.IndexConfig{
			{Fields: "ItemID", Multiple: true},
			{Fields: "RequesterID", Multiple: true},
			{Fields: "Type,RequesterID,ItemID"},
		},
		MaxSize: cap,
		IgnoreErr: func(err error) bool {
			// don't cache any errors,
			// it gets a little too tricky
			// otherwise with ensuring
			// errors are cleared out
			return true
		},
		Copy: copyF,
	})
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
