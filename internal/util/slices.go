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

// Deduplicate deduplicates entries in the given slice.
func Deduplicate[T comparable](in []T) []T {
	var (
		inL     = len(in)
		unique  = make(map[T]struct{}, inL)
		deduped = make([]T, 0, inL)
	)

	for _, v := range in {
		if _, ok := unique[v]; ok {
			// Already have this.
			continue
		}

		unique[v] = struct{}{}
		deduped = append(deduped, v)
	}

	return deduped
}

// DeduplicateFunc deduplicates entries in the given
// slice, using the result of key() to gauge uniqueness.
func DeduplicateFunc[T any, C comparable](in []T, key func(v T) C) []T {
	var (
		inL     = len(in)
		unique  = make(map[C]struct{}, inL)
		deduped = make([]T, 0, inL)
	)

	for _, v := range in {
		k := key(v)

		if _, ok := unique[k]; ok {
			// Already have this.
			continue
		}

		unique[k] = struct{}{}
		deduped = append(deduped, v)
	}

	return deduped
}

// Collate will collect the values of type T from a provided slice,
// passing each slice item to 'get' and deduplicating the end result.
func Collate[T any, K comparable](in []T, get func(T) K) []K {
	ks := make([]K, 0, len(in))
	km := make(map[K]struct{}, len(in))

	for i := 0; i < len(in); i++ {
		// Get next k.
		k := get(in[i])

		if _, ok := km[k]; !ok {
			// New value, add
			// to map + slice.
			ks = append(ks, k)
			km[k] = struct{}{}
		}
	}

	return ks
}

// OrderBy orders a slice of given type by the provided alternative slice of comparable type.
func OrderBy[T any, K comparable](in []T, keys []K, key func(T) K) {
	var (
		start  int
		offset int
	)

	for i := 0; i < len(keys); i++ {
		var (
			// key at index.
			k = keys[i]

			// sentinel
			// idx value.
			idx = -1
		)

		// Look for model with key in slice.
		for j := start; j < len(in); j++ {
			if key(in[j]) == k {
				idx = j
				break
			}
		}

		if idx == -1 {
			// model with key
			// was not found.
			offset++
			continue
		}

		// Update
		// start
		start++

		// Expected ID index.
		exp := i - offset

		if idx == exp {
			// Model is in expected
			// location, keep going.
			continue
		}

		// Swap models at current and expected.
		in[idx], in[exp] = in[exp], in[idx]
	}
}

// DeleteIf iterates through each element in a slice, passing each
// to delete to determine whether to remove. Returns resulting slice.
func DeleteIf[T any](in []T, delete func(T) bool) []T {
	for i := 0; i < len(in); {

		// Check if item
		// needs deleting.
		if delete(in[i]) {

			// Remove from slice by
			// shifting all down 1.
			copy(in[i:], in[i+1:])
			in = in[:len(in)-1]
			continue
		}

		// Iter.
		i++
	}
	return in
}
