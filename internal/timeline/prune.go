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

	l := t.items.data
	if l == nil {
		// Nothing to prune.
		return 0
	}

	var (
		position    int
		totalPruned int
		toRemove    *[]*list.Element
	)

	// Only initialize toRemove if we know we're
	// going to need it, otherwise skiperino.
	if toRemoveLen := t.items.data.Len() - desiredIndexedItemsLength; toRemoveLen > 0 {
		toRemove = func() *[]*list.Element { tr := make([]*list.Element, 0, toRemoveLen); return &tr }()
	}

	// Work from the front of the list until we get
	// to the point where we need to start pruning.
	for e := l.Front(); e != nil; e = e.Next() {
		position++

		if position <= desiredPreparedItemsLength {
			// We're still within our allotted
			// prepped length, nothing to do yet.
			continue
		}

		// We need to *at least* unprepare this entry.
		// If we're beyond our indexed length already,
		// we can just remove the item completely.
		if position > desiredIndexedItemsLength {
			*toRemove = append(*toRemove, e)
			totalPruned++
			continue
		}

		entry := e.Value.(*indexedItemsEntry) //nolint:forcetypeassert
		if entry.prepared == nil {
			// It's already unprepared (mood).
			continue
		}

		entry.prepared = nil // <- eat this up please garbage collector nom nom nom
		totalPruned++
	}

	if toRemove != nil {
		for _, e := range *toRemove {
			l.Remove(e)
		}
	}

	return totalPruned
}
