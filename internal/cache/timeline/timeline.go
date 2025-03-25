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

package timeline

import (
	"codeberg.org/gruf/go-structr"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// nextPageParams gets the next set of paging
// parameters to use based on the current set,
// and the next set of lo / hi values. This will
// correctly handle making sure that, depending
// on the paging order, the cursor value gets
// updated while maintaining the boundary value.
func nextPageParams(
	curLo, curHi string,
	nextLo, nextHi string,
	order paging.Order,
) (lo string, hi string) {
	if order.Ascending() {
		return nextLo, curHi
	} else /* i.e. descending */ {
		return curLo, nextHi
	}
}

// toDirection converts page order to timeline direction.
func toDirection(order paging.Order) structr.Direction {
	if order.Ascending() {
		return structr.Asc
	} else /* i.e. descending */ {
		return structr.Desc
	}
}
