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

package domain

import (
	"fmt"
	"time"

	"codeberg.org/gruf/go-cache/v3/ttl"
	"github.com/miekg/dns"
)

type BlockCache struct {
	pcache *ttl.Cache[string, bool]
	blocks []block
	loaded bool
}

func New(pcap int, pttl time.Duration) *BlockCache {
	c := new(BlockCache)
	c.pcache = new(ttl.Cache[string, bool])
	c.pcache.Init(0, pcap, pttl)
	return c
}

func (b *BlockCache) Start(pfreq time.Duration) bool {
	return b.pcache.Start(pfreq)
}

func (b *BlockCache) Stop() bool {
	return b.pcache.Stop()
}

func (b *BlockCache) IsBlocked(domain string, load func() ([]string, error)) (bool, error) {
	var blocked bool

	// Acquire cache lock
	b.pcache.Lock()
	defer b.pcache.Unlock()

	if !b.loaded {
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

		// Mark as loaded
		b.loaded = true
	}

	// Check primary cache for result
	entry, ok := b.pcache.Cache.Get(domain)
	if ok {
		return entry.Value, nil
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

func (b *BlockCache) Clear() {
	b.pcache.Lock()
	b.loaded = false
	b.pcache.Unlock()
	b.pcache.Clear()
}

type block struct {
	labels []string
}

func (b block) Blocks(labels []string) bool {
	if len(labels) < len(b.labels) {
		return false
	}

	for i := range b.labels {
		if labels[i] != b.labels[i] {
			return false
		}
	}

	return true
}
