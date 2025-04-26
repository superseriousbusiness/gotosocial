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

package xslices

import (
	"slices"
)

// ToAny converts a slice of any input type
// to the abstrace empty interface slice type.
func ToAny[T any](in []T) []any {
	out := make([]any, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

// GrowJust increases slice capacity to guarantee
// extra room 'size', where in the case that it does
// need to allocate more it ONLY allocates 'size' extra.
// This is different to typical slices.Grow behaviour,
// which simply guarantees extra through append() which
// may allocate more than necessary extra size.
func GrowJust[T any](in []T, size int) []T {

	if cap(in)-len(in) < size {
		// Reallocate enough for in + size.
		in2 := make([]T, len(in), len(in)+size)
		_ = copy(in2, in)
		in = in2
	}

	return in
}

// AppendJust appends extra elements to slice,
// ONLY allocating at most len(extra) elements. This
// is different to the typical append behaviour which
// will append extra, in a manner to reduce the need
// for new allocations on every call to append.
func AppendJust[T any](in []T, extra ...T) []T {
	l := len(in)

	if cap(in)-l < len(extra) {
		// Reallocate enough for + extra.
		in2 := make([]T, l+len(extra))
		_ = copy(in2, in)
		in = in2
	} else {
		// Reslice for + extra.
		in = in[:l+len(extra)]
	}

	// Copy extra into slice.
	_ = copy(in[l:], extra)
	return in
}

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

	if key == nil {
		panic("nil func")
	}

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

// Gather will collect the values of type V from input type []T,
// passing each item to 'get' and appending V to the return slice.
func Gather[T, V any](out []V, in []T, get func(T) V) []V {
	if get == nil {
		panic("nil func")
	}

	// Starting write index
	// in the resliced / re
	// alloc'd output slice.
	start := len(out)

	// Total required slice len.
	total := start + len(in)

	if total > cap(out) {
		// Reallocate output with
		// capacity for total len.
		out2 := make([]V, len(out), total)
		copy(out2, out)
		out = out2
	}

	// Reslice with capacity
	// up to total required.
	out = out[:total]

	// Gather vs from 'in'.
	for i, v := range in {
		j := start + i
		out[j] = get(v)
	}

	return out
}

// GatherIf is functionally similar to Gather(), but only when return bool is true.
// If you don't need to check the boolean, Gather() will be very slightly faster.
func GatherIf[T, V any](out []V, in []T, get func(T) (V, bool)) []V {
	if get == nil {
		panic("nil func")
	}

	if cap(out)-len(out) < len(in) {
		// Reallocate output with capacity for 'in'.
		out2 := make([]V, len(out), cap(out)+len(in))
		copy(out2, out)
		out = out2
	}

	// Gather vs from 'in'.
	for _, v := range in {
		if v, ok := get(v); ok {
			out = append(out, v)
		}
	}

	return out
}

// Collate will collect the values of type K from input type []T,
// passing each item to 'get' and deduplicating the end result.
// This is equivalent to calling Gather() followed by Deduplicate().
func Collate[T any, K comparable](in []T, get func(T) K) []K {
	if get == nil {
		panic("nil func")
	}

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
	if key == nil {
		panic("nil func")
	}

	// Create lookup of keys->idx.
	m := make(map[K]int, len(in))
	for i, k := range keys {
		m[k] = i
	}

	// Sort according to the reverse lookup.
	slices.SortFunc(in, func(a, b T) int {
		ai := m[key(a)]
		bi := m[key(b)]
		if ai < bi {
			return -1
		} else if bi < ai {
			return +1
		}
		return 0
	})
}
