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

package paging

import "fmt"

// MinID returns an ID boundary with given min ID value,
// using either the `since_id`,"DESC" name,ordering or
// `min_id`,"ASC" name,ordering depending on which is set.
func MinID(minID, sinceID string) Boundary[string] {
	switch {
	case minID != "":
		return Boundary[string]{
			Name:  "min_id",
			Value: minID,

			// with "min_id", return
			// as ascending order items,
			// i.e. oldest at lowest idx
			Order: OrderAscending,
		}
	default:
		// default min is `since_id`
		return Boundary[string]{
			Name:  "since_id",
			Value: sinceID,

			// with "since_id", return
			// as descending order items,
			// i.e. newest at lowest idx
			Order: OrderDescending,
		}
	}
}

// MaxID returns an ID boundary with given max
// ID value, and the "max_id" query key set.
func MaxID(maxID string) Boundary[string] {
	return Boundary[string]{
		Name:  "max_id",
		Value: maxID,

		// by default uses descending,
		// but min boundary order always
		// overrides max boundary order
		Order: OrderDescending,
	}
}

// MinShortcodeDomain returns a boundary with the given minimum emoji
// shortcode@domain, and the "min_shortcode_domain" query key set.
func MinShortcodeDomain(min string) Boundary[string] {
	return Boundary[string]{
		Name:  "min_shortcode_domain",
		Value: min,

		// with "min_shortcode_domain",
		// return as ascending order items,
		// i.e. oldest at lowest idx
		Order: OrderAscending,
	}
}

// MaxShortcodeDomain returns a boundary with the given maximum emoji
// shortcode@domain, and the "max_shortcode_domain" query key set.
func MaxShortcodeDomain(max string) Boundary[string] {
	return Boundary[string]{
		Name:  "max_shortcode_domain",
		Value: max,

		// by default uses descending,
		// but min boundary order always
		// overrides max boundary order
		Order: OrderDescending,
	}
}

// Boundary represents the upper or lower limit in a page slice.
type Boundary[T comparable] struct {
	Name  string // i.e. query key
	Value T
	Order Order
}

// new creates a new Boundary with the same ordering and name
// as the original (receiving), but with the new provided value.
func (b Boundary[T]) new(value T) Boundary[T] {
	return Boundary[T]{
		Name:  b.Name,
		Value: value,
		Order: b.Order,
	}
}

// Find finds the boundary's set value in input slice, or returns -1.
func (b Boundary[T]) Find(in []T) int {
	if zero(b.Value) {
		return -1
	}
	for i := range in {
		if in[i] == b.Value {
			return i
		}
	}
	return -1
}

// Query returns this boundary as assembled query key=value pair.
func (b Boundary[T]) Query() string {
	switch {
	case zero(b.Value):
		return ""
	case b.Name == "":
		panic("value without boundary name")
	default:
		return fmt.Sprintf("%s=%v", b.Name, b.Value)
	}
}
