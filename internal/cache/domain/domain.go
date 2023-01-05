/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package domain

import (
	"fmt"
	"time"

	"codeberg.org/gruf/go-cache/v3/ttl"
	"github.com/miekg/dns"
)

// BlockCache provides a means of caching domain blocks in memory to reduce load
// on an underlying storage mechanism, e.g. a database.
//
// It consists of a TTL primary cache that stores calculated domain string to block results,
// that on cache miss is filled by calculating block status by iterating over a list of all of
// the domain blocks stored in memory. This reduces CPU usage required by not need needing to
// iterate through a possible 100-1000s long block list, while saving memory by having a primary
// cache of limited size that evicts stale entries. The raw list of all domain blocks should in
// most cases be negligible when it comes to memory usage.
//
// The in-memory block list is kept up-to-date by means of a passed loader function during every
// call to .IsBlocked(). In the case of a nil internal block list, the loader function is called to
// hydrate the cache with the latest list of domain blocks. The .Clear() function can be used to invalidate
// the cache, e.g. when a domain block is added / deleted from the database. It will drop the current
// list of domain blocks and clear all entries from the primary cache.
type BlockCache struct {
	pcache *ttl.Cache[string, bool] // primary cache of domains -> block results
	blocks []block                  // raw list of all domain blocks, nil => not loaded.
}

// New returns a new initialized BlockCache instance with given primary cache capacity and TTL.
func New(pcap int, pttl time.Duration) *BlockCache {
	c := new(BlockCache)
	c.pcache = new(ttl.Cache[string, bool])
	c.pcache.Init(0, pcap, pttl)
	return c
}

// Start will start the cache background eviction routine with given sweep frequency. If already running or a freq <= 0 provided, this is a no-op. This will block until the eviction routine has started.
func (b *BlockCache) Start(pfreq time.Duration) bool {
	return b.pcache.Start(pfreq)
}

// Stop will stop cache background eviction routine. If not running this is a no-op. This will block until the eviction routine has stopped.
func (b *BlockCache) Stop() bool {
	return b.pcache.Stop()
}

// IsBlocked checks whether domain is blocked. If the cache is not currently loaded, then the provided load function is used to hydrate it.
// NOTE: be VERY careful using any kind of locking mechanism within the load function, as this itself is ran within the cache mutex lock.
func (b *BlockCache) IsBlocked(domain string, load func() ([]string, error)) (bool, error) {
	var blocked bool

	// Acquire cache lock
	b.pcache.Lock()
	defer b.pcache.Unlock()

	// Check primary cache for result
	entry, ok := b.pcache.Cache.Get(domain)
	if ok {
		return entry.Value, nil
	}

	if b.blocks == nil {
		// Cache is not hydrated
		//
		// Load domains from callback
		domains, err := load()
		if err != nil {
			return false, fmt.Errorf("error reloading cache: %w", err)
		}

		// Drop all domain blocks and recreate
		b.blocks = make([]block, len(domains))

		for i, domain := range domains {
			// Store pre-split labels for each domain block
			b.blocks[i].labels = dns.SplitDomainName(domain)
		}
	}

	// Split domain into it separate labels
	labels := dns.SplitDomainName(domain)

	// Compare this to our stored blocks
	for _, block := range b.blocks {
		if block.Blocks(labels) {
			blocked = true
			break
		}
	}

	// Store block result in primary cache
	b.pcache.Cache.Set(domain, &ttl.Entry[string, bool]{
		Key:    domain,
		Value:  blocked,
		Expiry: time.Now().Add(b.pcache.TTL),
	})

	return blocked, nil
}

// Clear will drop the currently loaded domain list, and clear the primary cache.
// This will trigger a reload on next call to .IsBlocked().
func (b *BlockCache) Clear() {
	// Drop all blocks.
	b.pcache.Lock()
	b.blocks = nil
	b.pcache.Unlock()

	// Clear needs to be done _outside_ of
	// lock, as also acquires a mutex lock.
	b.pcache.Clear()
}

// block represents a domain block, and stores the
// deconstructed labels of a singular domain block.
// e.g. []string{"gts", "superseriousbusiness", "org"}.
type block struct {
	labels []string
}

// Blocks checks whether the separated domain labels of an
// incoming domain matches the stored (receiving struct) block.
func (b block) Blocks(labels []string) bool {
	// Calculate length difference
	d := len(labels) - len(b.labels)
	if d < 0 {
		return false
	}

	// Iterate backwards through domain block's
	// labels, omparing against the incoming domain's.
	//
	// So for the following input:
	// labels   = []string{"mail", "google", "com"}
	// b.labels = []string{"google", "com"}
	//
	// These would be matched in reverse order along
	// the entirety of the block object's labels:
	// "com"    => match
	// "google" => match
	//
	// And so would reach the end and return true.
	for i := len(b.labels) - 1; i >= 0; i-- {
		if b.labels[i] != labels[i+d] {
			return false
		}
	}

	return true
}
