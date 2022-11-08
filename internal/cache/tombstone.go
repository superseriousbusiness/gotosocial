/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package cache

import (
	"time"

	"codeberg.org/gruf/go-cache/v2"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// TombstoneCache is a cache wrapper to provide lookups for *gtsmodel.Tombstones
type TombstoneCache struct {
	cache cache.LookupCache[string, string, *gtsmodel.Tombstone]
}

// NewTombstoneCache returns a new instantiated TombstoneCache object
func NewTombstoneCache() *TombstoneCache {
	c := &TombstoneCache{}
	c.cache = cache.NewLookup(cache.LookupCfg[string, string, *gtsmodel.Tombstone]{
		RegisterLookups: func(lm *cache.LookupMap[string, string]) {
			lm.RegisterLookup("uri")
		},

		AddLookups: func(lm *cache.LookupMap[string, string], tombstone *gtsmodel.Tombstone) {
			lm.Set("uri", tombstone.URI, tombstone.ID)
		},

		DeleteLookups: func(lm *cache.LookupMap[string, string], tombstone *gtsmodel.Tombstone) {
			lm.Delete("uri", tombstone.URI)
		},
	})
	c.cache.SetTTL(time.Minute*5, false)
	c.cache.Start(time.Second * 10)
	return c
}

func (c *TombstoneCache) GetByURI(uri string) (*gtsmodel.Tombstone, bool) {
	return c.cache.GetBy("uri", uri)
}

func (c *TombstoneCache) Put(tombstone *gtsmodel.Tombstone) {
	c.cache.Put(tombstone.ID, copyTombstone(tombstone))
}

func (c *TombstoneCache) Invalidate(id string) {
	c.cache.Invalidate(id)
}

func copyTombstone(tombstone *gtsmodel.Tombstone) *gtsmodel.Tombstone {
	return &gtsmodel.Tombstone{
		ID:        tombstone.ID,
		CreatedAt: tombstone.CreatedAt,
		UpdatedAt: tombstone.UpdatedAt,
		Domain:    tombstone.Domain,
		URI:       tombstone.URI,
	}
}
