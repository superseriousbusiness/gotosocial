/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

type preparedItems struct {
	data       *list.List
	skipInsert SkipInsertFunction
}

type preparedItemsEntry struct {
	itemID           string
	boostOfID        string
	accountID        string
	boostOfAccountID string
	prepared         Preparable
}

func (p *preparedItems) insertPrepared(ctx context.Context, newEntry *preparedItemsEntry) error {
	if p.data == nil {
		p.data = &list.List{}
	}

	// if we have no entries yet, this is both the newest and oldest entry, so just put it in the front
	if p.data.Len() == 0 {
		p.data.PushFront(newEntry)
		return nil
	}

	var insertMark *list.Element
	var position int
	// We need to iterate through the index to make sure we put this entry in the appropriate place according to when it was created.
	// We also need to make sure we're not inserting a duplicate entry -- this can happen sometimes and it's not nice UX (*shudder*).
	for e := p.data.Front(); e != nil; e = e.Next() {
		position++

		entry, ok := e.Value.(*preparedItemsEntry)
		if !ok {
			return errors.New("insertPrepared: could not parse e as a preparedItemsEntry")
		}

		skip, err := p.skipInsert(ctx, newEntry.itemID, newEntry.accountID, newEntry.boostOfID, newEntry.boostOfAccountID, entry.itemID, entry.accountID, entry.boostOfID, entry.boostOfAccountID, position)
		if err != nil {
			return err
		}
		if skip {
			return nil
		}

		// if the entry to index is newer than e, insert it before e in the list
		if insertMark == nil {
			if newEntry.itemID > entry.itemID {
				insertMark = e
			}
		}

		// make sure we don't insert a duplicate
		if entry.itemID == newEntry.itemID {
			return nil
		}
	}

	if insertMark != nil {
		p.data.InsertBefore(newEntry, insertMark)
		return nil
	}

	// if we reach this point it's the oldest entry we've seen so put it at the back
	p.data.PushBack(newEntry)
	return nil
}
