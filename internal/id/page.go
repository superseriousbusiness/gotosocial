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

package id

import (
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// ValidatePage ensures that passed page has valid paging
// values for the current defined ordering. That is, it
// ensures a valid page *cursor* value, using id.Highest
// or id.Lowest where appropriate when none given.
func ValidatePage(page *paging.Page) {
	if page == nil {
		// unpaged
		return
	}

	switch page.Order() {
	// If the page order is ascending,
	// ensure that a minimum value is set.
	// This will be used as the cursor.
	case paging.OrderAscending:
		if page.Min.Value == "" {
			page.Min.Value = Lowest
		}

	// If the page order is descending,
	// ensure that a maximum value is set.
	// This will be used as the cursor.
	case paging.OrderDescending:
		if page.Max.Value == "" {
			page.Max.Value = Highest
		}
	}
}
