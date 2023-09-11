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

// Reverse will reverse the given input slice.
func Reverse(in []string) []string {
	var (
		// Start at front.
		i = 0

		// Start at back.
		j = len(in) - 1
	)

	for i < j {
		// Swap i,j index values in slice.
		in[i], in[j] = in[j], in[i]

		// incr + decr,
		// looping until
		// they meet in
		// the middle.
		i++
		j--
	}

	return in
}
