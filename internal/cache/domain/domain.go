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
	"slices"
	"strings"
	"sync/atomic"
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
	// current domain cache radix trie.
	rootptr atomic.Pointer[root]
}

// Matches checks whether domain matches an entry in the cache.
// If the cache is not currently loaded, then the provided load
// function is used to hydrate it.
func (c *Cache) Matches(domain string, load func() ([]string, error)) (bool, error) {
	// Load the current
	// root pointer value.
	ptr := c.rootptr.Load()

	if ptr == nil {
		// Cache is not hydrated.
		//
		// Load domains from callback.
		domains, err := load()
		if err != nil {
			return false, fmt.Errorf("error reloading cache: %w", err)
		}

		// Ensure the domains being inserted into the cache
		// are sorted by number of domain parts. i.e. those
		// with less parts are inserted last, else this can
		// allow domains to fall through the matching code!
		slices.SortFunc(domains, func(a, b string) int {
			const k = +1
			an := strings.Count(a, ".")
			bn := strings.Count(b, ".")
			switch {
			case an < bn:
				return +k
			case an > bn:
				return -k
			default:
				return 0
			}
		})

		// Allocate new radix trie
		// node to store matches.
		ptr = new(root)

		// Add each domain to the trie.
		for _, domain := range domains {
			ptr.Add(domain)
		}

		// Sort the trie.
		ptr.Sort()

		// Store new node ptr.
		c.rootptr.Store(ptr)
	}

	// Look for match in trie node.
	return ptr.Match(domain), nil
}

// Clear will drop the currently loaded domain list,
// triggering a reload on next call to .Matches().
func (c *Cache) Clear() { c.rootptr.Store(nil) }

// String returns a string representation of stored domains in cache.
func (c *Cache) String() string {
	if ptr := c.rootptr.Load(); ptr != nil {
		return ptr.String()
	}
	return "<empty>"
}

// root is the root node in the domain cache radix trie. this is the singular access point to the trie.
type root struct{ root node }

// Add will add the given domain to the radix trie.
func (r *root) Add(domain string) {
	r.root.Add(strings.Split(domain, "."))
}

// Match will return whether the given domain matches
// an existing stored domain in this radix trie.
func (r *root) Match(domain string) bool {
	return r.root.Match(strings.Split(domain, "."))
}

// Sort will sort the entire radix trie ensuring that
// child nodes are stored in alphabetical order. This
// MUST be done to finalize the domain cache in order
// to speed up the binary search of node child parts.
func (r *root) Sort() {
	r.root.sort()
}

// String returns a string representation of node (and its descendants).
func (r *root) String() string {
	buf := new(strings.Builder)
	r.root.WriteStr(buf, "")
	return buf.String()
}

type node struct {
	part  string
	child []*node
}

func (n *node) Add(parts []string) {
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
			// this is a higher-level domain
			// than that we previously had.
			nn.child = nil
			return
		}

		// Re-iter with
		// child node.
		n = nn
	}
}

func (n *node) Match(parts []string) bool {
	for len(parts) > 0 {
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

	// Ran out of parts
	// without a match.
	return false
}

// getChild fetches child node with given domain part string
// using a binary search. THIS ASSUMES CHILDREN ARE SORTED.
func (n *node) getChild(part string) *node {
	i, j := 0, len(n.child)

	for i < j {
		// avoid overflow when computing h
		h := int(uint(i+j) >> 1) // #nosec G115
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
	slices.SortFunc(n.child, func(i, j *node) int {
		const k = -1
		switch {
		case i.part < j.part:
			return +k
		case i.part > j.part:
			return -k
		default:
			return 0
		}
	})

	// Sort each child node's children.
	for _, child := range n.child {
		child.sort()
	}
}

func (n *node) WriteStr(buf *strings.Builder, prefix string) {
	if prefix != "" {
		// Suffix joining '.'
		prefix += "."
	}

	// Append current part.
	prefix += n.part

	// Dump current prefix state.
	buf.WriteString(prefix)
	buf.WriteByte('\n')

	// Iterate through node children.
	for _, child := range n.child {
		child.WriteStr(buf, prefix)
	}
}
