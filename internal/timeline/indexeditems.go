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
	"context"
	"fmt"
)

type indexedItems struct {
	data       *list.List
	skipInsert SkipInsertFunction
}

type indexedItemsEntry struct {
	itemID           string
	boostOfID        string
	accountID        string
	boostOfAccountID string
	prepared         Preparable
}

// WARNING: ONLY CALL THIS FUNCTION IF YOU ALREADY HAVE
// A LOCK ON THE TIMELINE CONTAINING THIS INDEXEDITEMS!
func (i *indexedItems) insertIndexed(ctx context.Context, newEntry *indexedItemsEntry) (bool, error) {
	// Lazily init indexed items.
	if i.data == nil {
		i.data = &list.List{}
		i.data.Init()
	}

	if i.data.Len() == 0 {
		// We have no entries yet, meaning this is both the
		// newest + oldest entry, so just put it in the front.
		i.data.PushFront(newEntry)
		return true, nil
	}

	var (
		insertMark      *list.Element
		currentPosition int
	)

	// We need to iterate through the index to make sure we put
	// this item in the appropriate place according to its id.
	// We also need to make sure we're not inserting a duplicate
	// item -- this can happen sometimes and it's sucky UX.
	for e := i.data.Front(); e != nil; e = e.Next() {
		currentPosition++

		currentEntry := e.Value.(*indexedItemsEntry) //nolint:forcetypeassert

		// Check if we need to skip inserting this item based on
		// the current item.
		//
		// For example, if the new item is a boost, and the current
		// item is the original, we may not want to insert the boost
		// if it would appear very shortly after the original.
		if skip, err := i.skipInsert(
			ctx,
			newEntry.itemID,
			newEntry.accountID,
			newEntry.boostOfID,
			newEntry.boostOfAccountID,
			currentEntry.itemID,
			currentEntry.accountID,
			currentEntry.boostOfID,
			currentEntry.boostOfAccountID,
			currentPosition,
		); err != nil {
			return false, fmt.Errorf("insertIndexed: error calling skipInsert: %w", err)
		} else if skip {
			// We don't need to insert this at all,
			// so we can safely bail.
			return false, nil
		}

		if insertMark != nil {
			// We already found our mark.
			continue
		}

		if currentEntry.itemID > newEntry.itemID {
			// We're still in items newer than
			// the one we're trying to insert.
			continue
		}

		// We found our spot!
		insertMark = e
	}

	if insertMark == nil {
		// We looked through the whole timeline and didn't find
		// a mark, so the new item is the oldest item we've seen;
		// insert it at the back.
		i.data.PushBack(newEntry)
		return true, nil
	}

	i.data.InsertBefore(newEntry, insertMark)
	return true, nil
}
