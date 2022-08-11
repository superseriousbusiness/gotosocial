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

// DomainCache is a cache wrapper to provide URL and URI lookups for gtsmodel.Status
type DomainBlockCache struct {
	cache cache.LookupCache[string, string, *gtsmodel.DomainBlock]
}

// NewStatusCache returns a new instantiated statusCache object
func NewDomainBlockCache() *DomainBlockCache {
	c := &DomainBlockCache{}
	c.cache = cache.NewLookup(cache.LookupCfg[string, string, *gtsmodel.DomainBlock]{
		RegisterLookups: func(lm *cache.LookupMap[string, string]) {
			lm.RegisterLookup("id")
		},

		AddLookups: func(lm *cache.LookupMap[string, string], block *gtsmodel.DomainBlock) {
			if block.ID != "" {
				lm.Set("id", block.ID, block.Domain)
			}
		},

		DeleteLookups: func(lm *cache.LookupMap[string, string], block *gtsmodel.DomainBlock) {
			if block.ID != "" {
				lm.Delete("id", block.ID)
			}
		},
	})
	c.cache.SetTTL(time.Minute*5, false)
	c.cache.Start(time.Second * 10)
	return c
}

// GetByID attempts to fetch a status from the cache by its ID, you will receive a copy for thread-safety
func (c *DomainBlockCache) GetByID(id string) (*gtsmodel.DomainBlock, bool) {
	return c.cache.GetBy("id", id)
}

// GetByURL attempts to fetch a status from the cache by its URL, you will receive a copy for thread-safety
func (c *DomainBlockCache) GetByDomain(domain string) (*gtsmodel.DomainBlock, bool) {
	return c.cache.Get(domain)
}

// Put places a status in the cache, ensuring that the object place is a copy for thread-safety
func (c *DomainBlockCache) Put(block *gtsmodel.DomainBlock) {
	if block == nil || block.Domain == "" {
		panic("invalid domain")
	}
	c.cache.Set(block.Domain, copyDomainBlock(block))
}

// InvalidateByDomain will invalidate a domain block from the cache by domain name.
func (c *DomainBlockCache) InvalidateByDomain(domain string) {
	c.cache.Invalidate(domain)
}

// copyStatus performs a surface-level copy of status, only keeping attached IDs intact, not the objects.
// due to all the data being copied being 99% primitive types or strings (which are immutable and passed by ptr)
// this should be a relatively cheap process
func copyDomainBlock(block *gtsmodel.DomainBlock) *gtsmodel.DomainBlock {
	return &gtsmodel.DomainBlock{
		ID:                 block.ID,
		CreatedAt:          block.CreatedAt,
		UpdatedAt:          block.UpdatedAt,
		Domain:             block.Domain,
		CreatedByAccountID: block.CreatedByAccountID,
		CreatedByAccount:   nil,
		PrivateComment:     block.PrivateComment,
		PublicComment:      block.PublicComment,
		Obfuscate:          block.Obfuscate,
		SubscriptionID:     block.SubscriptionID,
	}
}
