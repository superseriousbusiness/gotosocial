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
	"container/list"
)

func (t *timeline) Prune(desiredPreparedItemsLength int, desiredIndexedItemsLength int) int {
	t.Lock()
	defer t.Unlock()

	pruneList := func(pruneTo int, listToPrune *list.List) int {
		if listToPrune == nil {
			// no need to prune
			return 0
		}

		unprunedLength := listToPrune.Len()
		if unprunedLength <= pruneTo {
			// no need to prune
			return 0
		}

		// work from the back + assemble a slice of entries that we will prune
		amountStillToPrune := unprunedLength - pruneTo
		itemsToPrune := make([]*list.Element, 0, amountStillToPrune)
		for e := listToPrune.Back(); amountStillToPrune > 0; e = e.Prev() {
			itemsToPrune = append(itemsToPrune, e)
			amountStillToPrune--
		}

		// remove the entries we found
		var totalPruned int
		for _, e := range itemsToPrune {
			listToPrune.Remove(e)
			totalPruned++
		}

		return totalPruned
	}

	prunedPrepared := pruneList(desiredPreparedItemsLength, t.preparedItems.data)
	prunedIndexed := pruneList(desiredIndexedItemsLength, t.indexedItems.data)

	return prunedPrepared + prunedIndexed
}
