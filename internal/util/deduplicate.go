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
