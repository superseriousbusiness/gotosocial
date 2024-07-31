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

package util

// Set represents a hashmap of only keys,
// useful for deduplication / key checking.
type Set[T comparable] map[T]struct{}

// ToSet creates a Set[T] from given values,
// noting that this does not maintain any order.
func ToSet[T comparable](in []T) Set[T] {
	set := make(Set[T], len(in))
	for _, v := range in {
		set[v] = struct{}{}
	}
	return set
}

// FromSet extracts the values from set to slice,
// noting that this does not maintain any order.
func FromSet[T comparable](in Set[T]) []T {
	out := make([]T, len(in))
	var i int
	for v := range in {
		out[i] = v
		i++
	}
	return out
}

// In returns input slice filtered to
// only contain those in receiving set.
func (s Set[T]) In(vs []T) []T {
	out := make([]T, 0, len(vs))
	for _, v := range vs {
		if _, ok := s[v]; ok {
			out = append(out, v)
		}
	}
	return out
}

// NotIn is the functional inverse of In().
func (s Set[T]) NotIn(vs []T) []T {
	out := make([]T, 0, len(vs))
	for _, v := range vs {
		if _, ok := s[v]; !ok {
			out = append(out, v)
		}
	}
	return out
}

// Has returns if value is in Set.
func (s Set[T]) Has(v T) bool {
	_, ok := s[v]
	return ok
}
