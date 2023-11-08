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

// EqualPtrs returns whether the values contained within two comparable ptr types are equal.
func EqualPtrs[T comparable](t1, t2 *T) bool {
	switch {
	case t1 == nil:
		return (t2 == nil)
	case t2 == nil:
		return false
	default:
		return (*t1 == *t2)
	}
}

// Ptr returns a pointer to the passed in type
func Ptr[T any](t T) *T {
	return &t
}
