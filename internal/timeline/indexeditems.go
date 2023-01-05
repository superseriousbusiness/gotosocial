/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package timeline

import (
	"container/list"
	"context"
	"errors"
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
}

func (i *indexedItems) insertIndexed(ctx context.Context, newEntry *indexedItemsEntry) (bool, error) {
	if i.data == nil {
		i.data = &list.List{}
	}

	// if we have no entries yet, this is both the newest and oldest entry, so just put it in the front
	if i.data.Len() == 0 {
		i.data.PushFront(newEntry)
		return true, nil
	}

	var insertMark *list.Element
	var position int
	// We need to iterate through the index to make sure we put this item in the appropriate place according to when it was created.
	// We also need to make sure we're not inserting a duplicate item -- this can happen sometimes and it's not nice UX (*shudder*).
	for e := i.data.Front(); e != nil; e = e.Next() {
		position++

		entry, ok := e.Value.(*indexedItemsEntry)
		if !ok {
			return false, errors.New("insertIndexed: could not parse e as an indexedItemsEntry")
		}

		skip, err := i.skipInsert(ctx, newEntry.itemID, newEntry.accountID, newEntry.boostOfID, newEntry.boostOfAccountID, entry.itemID, entry.accountID, entry.boostOfID, entry.boostOfAccountID, position)
		if err != nil {
			return false, err
		}
		if skip {
			return false, nil
		}

		// if the item to index is newer than e, insert it before e in the list
		if insertMark == nil {
			if newEntry.itemID > entry.itemID {
				insertMark = e
			}
		}
	}

	if insertMark != nil {
		i.data.InsertBefore(newEntry, insertMark)
		return true, nil
	}

	// if we reach this point it's the oldest item we've seen so put it at the back
	i.data.PushBack(newEntry)
	return true, nil
}
