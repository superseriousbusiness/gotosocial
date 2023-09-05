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

// Order represents a paging ordering.
type Order int

const (
	_default Order = iota
	OrderDescending
	OrderAscending
)

// Ascending returns whether this order is ascending.
func (i Order) Ascending() bool {
	return i == OrderAscending
}

// Descending returns whether this order is descending.
func (i Order) Descending() bool {
	return i == OrderDescending
}

// String returns a string representation of Order.
func (i Order) String() string {
	switch i {
	case OrderDescending:
		return "Descending"
	case OrderAscending:
		return "Ascending"
	default:
		return "not-specified"
	}
}
