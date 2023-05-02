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

package domain

import (
	"fmt"
	"strings"
	"sync/atomic"
	"unsafe"

	"golang.org/x/exp/slices"
)

// BlockCache provides a means of caching domain blocks in memory to reduce load
// on an underlying storage mechanism, e.g. a database.
//
// The in-memory block list is kept up-to-date by means of a passed loader function during every
// call to .IsBlocked(). In the case of a nil internal block list, the loader function is called to
// hydrate the cache with the latest list of domain blocks. The .Clear() function can be used to
// invalidate the cache, e.g. when a domain block is added / deleted from the database.
type BlockCache struct {
	// atomically updated ptr value to the
	// current domain block cache radix trie.
	rootptr unsafe.Pointer
}

// IsBlocked checks whether domain is blocked. If the cache is not currently loaded, then the provided load function is used to hydrate it.
func (b *BlockCache) IsBlocked(domain string, load func() ([]string, error)) (bool, error) {
	// Load the current root pointer value.
	ptr := atomic.LoadPointer(&b.rootptr)

	if ptr == nil {
		// Cache is not hydrated.
		//
		// Load domains from callback.
		domains, err := load()
		if err != nil {
			return false, fmt.Errorf("error reloading cache: %w", err)
		}

		// Allocate new radix trie
		// node to store matches.
		root := new(root)

		// Add each domain to the trie.
		for _, domain := range domains {
			root.Add(domain)
		}

		// Sort the trie.
		root.Sort()

		// Store the new node ptr.
		ptr = unsafe.Pointer(root)
		atomic.StorePointer(&b.rootptr, ptr)
	}

	// Look for a match in the trie node.
	return (*root)(ptr).Match(domain), nil
}

// Clear will drop the currently loaded domain list,
// triggering a reload on next call to .IsBlocked().
func (b *BlockCache) Clear() {
	atomic.StorePointer(&b.rootptr, nil)
}

// root is the root node in the domain
// block cache radix trie. this is the
// singular access point to the trie.
type root struct{ root node }

// Add will add the given domain to the radix trie.
func (r *root) Add(domain string) {
	r.root.add(strings.Split(domain, "."))
}

// Match will return whether the given domain matches
// an existing stored domain block in this radix trie.
func (r *root) Match(domain string) bool {
	return r.root.match(strings.Split(domain, "."))
}

// Sort will sort the entire radix trie ensuring that
// child nodes are stored in alphabetical order. This
// MUST be done to finalize the block cache in order
// to speed up the binary search of node child parts.
func (r *root) Sort() {
	r.root.sort()
}

type node struct {
	part  string
	child []*node
}

func (n *node) add(parts []string) {
	if len(parts) == 0 {
		panic("invalid domain")
	}

	for {
		// Pop next domain part.
		i := len(parts) - 1
		part := parts[i]
		parts = parts[:i]

		var nn *node

		// Look for existing child node
		// that matches next domain part.
		for _, child := range n.child {
			if child.part == part {
				nn = child
				break
			}
		}

		if nn == nil {
			// Alloc new child node.
			nn = &node{part: part}
			n.child = append(n.child, nn)
		}

		if len(parts) == 0 {
			// Drop all children here as
			// this is a higher-level block
			// than that we previously had.
			nn.child = nil
			return
		}

		// Re-iter with
		// child node.
		n = nn
	}
}

func (n *node) match(parts []string) bool {
	if len(parts) == 0 {
		// Invalid domain.
		return false
	}

	for {
		// Pop next domain part.
		i := len(parts) - 1
		part := parts[i]
		parts = parts[:i]

		// Look for existing child
		// that matches next part.
		nn := n.getChild(part)

		if nn == nil {
			// No match :(
			return false
		}

		if len(nn.child) == 0 {
			// It's a match!
			return true
		}

		// Re-iter with
		// child node.
		n = nn
	}
}

// getChild fetches child node with given domain part string
// using a binary search. THIS ASSUMES CHILDREN ARE SORTED.
func (n *node) getChild(part string) *node {
	i, j := 0, len(n.child)

	for i < j {
		// avoid overflow when computing h
		h := int(uint(i+j) >> 1)
		// i ≤ h < j

		if n.child[h].part < part {
			// preserves:
			// n.child[i-1].part != part
			i = h + 1
		} else {
			// preserves:
			// n.child[h].part == part
			j = h
		}
	}

	if i >= len(n.child) || n.child[i].part != part {
		return nil // no match
	}

	return n.child[i]
}

func (n *node) sort() {
	// Sort this node's slice of child nodes.
	slices.SortFunc(n.child, func(i, j *node) bool {
		return i.part < j.part
	})

	// Sort each child node's children.
	for _, child := range n.child {
		child.sort()
	}
}
