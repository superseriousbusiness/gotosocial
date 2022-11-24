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
	"fmt"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (t *timeline) ItemIndexLength(ctx context.Context) int {
	if t.indexedItems == nil || t.indexedItems.data == nil {
		return 0
	}
	return t.indexedItems.data.Len()
}

// func (t *timeline) indexBefore(ctx context.Context, itemID string, amount int) error {
// 	l := log.WithFields(kv.Fields{{"amount", amount}}...)

// 	// lazily initialize index if it hasn't been done already
// 	if t.indexedItems.data == nil {
// 		t.indexedItems.data = &list.List{}
// 		t.indexedItems.data.Init()
// 	}

// 	toIndex := []Timelineable{}
// 	offsetID := itemID

// 	l.Trace("entering grabloop")
// grabloop:
// 	for i := 0; len(toIndex) < amount && i < 5; i++ { // try the grabloop 5 times only
// 		// first grab items using the caller-provided grab function
// 		l.Trace("grabbing...")
// 		items, stop, err := t.grabFunction(ctx, t.accountID, "", "", offsetID, amount)
// 		if err != nil {
// 			return err
// 		}
// 		if stop {
// 			break grabloop
// 		}

// 		l.Trace("filtering...")
// 		// now filter each item using the caller-provided filter function
// 		for _, item := range items {
// 			shouldIndex, err := t.filterFunction(ctx, t.accountID, item)
// 			if err != nil {
// 				return err
// 			}
// 			if shouldIndex {
// 				toIndex = append(toIndex, item)
// 			}
// 			offsetID = item.GetID()
// 		}
// 	}
// 	l.Trace("left grabloop")

// 	// index the items we got
// 	for _, s := range toIndex {
// 		if _, err := t.IndexOne(ctx, s.GetID(), s.GetBoostOfID(), s.GetAccountID(), s.GetBoostOfAccountID()); err != nil {
// 			return fmt.Errorf("indexBefore: error indexing item with id %s: %s", s.GetID(), err)
// 		}
// 	}

// 	return nil
// }

func (t *timeline) indexBehind(ctx context.Context, itemID string, amount int) error {
	l := log.WithFields(kv.Fields{{"amount", amount}}...)

	// lazily initialize index if it hasn't been done already
	if t.indexedItems.data == nil {
		t.indexedItems.data = &list.List{}
		t.indexedItems.data.Init()
	}

	// If we're already indexedBehind given itemID by the required amount, we can return nil.
	// First find position of itemID (or as near as possible).
	var position int
positionLoop:
	for e := t.indexedItems.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*indexedItemsEntry)
		if !ok {
			return errors.New("indexBehind: could not parse e as an itemIndexEntry")
		}

		if entry.itemID <= itemID {
			// we've found it
			break positionLoop
		}
		position++
	}

	// now check if the length of indexed items exceeds the amount of items required (position of itemID, plus amount of posts requested after that)
	if t.indexedItems.data.Len() > position+amount {
		// we have enough indexed behind already to satisfy amount, so don't need to make db calls
		l.Trace("returning nil since we already have enough items indexed")
		return nil
	}

	toIndex := []Timelineable{}
	offsetID := itemID

	l.Trace("entering grabloop")
grabloop:
	for i := 0; len(toIndex) < amount && i < 5; i++ { // try the grabloop 5 times only
		// first grab items using the caller-provided grab function
		l.Trace("grabbing...")
		items, stop, err := t.grabFunction(ctx, t.accountID, offsetID, "", "", amount)
		if err != nil {
			return err
		}
		if stop {
			break grabloop
		}

		l.Trace("filtering...")
		// now filter each item using the caller-provided filter function
		for _, item := range items {
			shouldIndex, err := t.filterFunction(ctx, t.accountID, item)
			if err != nil {
				return err
			}
			if shouldIndex {
				toIndex = append(toIndex, item)
			}
			offsetID = item.GetID()
		}
	}
	l.Trace("left grabloop")

	// index the items we got
	for _, s := range toIndex {
		if _, err := t.IndexOne(ctx, s.GetID(), s.GetBoostOfID(), s.GetAccountID(), s.GetBoostOfAccountID()); err != nil {
			return fmt.Errorf("indexBehind: error indexing item with id %s: %s", s.GetID(), err)
		}
	}

	return nil
}

func (t *timeline) IndexOne(ctx context.Context, itemID string, boostOfID string, accountID string, boostOfAccountID string) (bool, error) {
	t.Lock()
	defer t.Unlock()

	postIndexEntry := &indexedItemsEntry{
		itemID:           itemID,
		boostOfID:        boostOfID,
		accountID:        accountID,
		boostOfAccountID: boostOfAccountID,
	}

	return t.indexedItems.insertIndexed(ctx, postIndexEntry)
}

func (t *timeline) IndexAndPrepareOne(ctx context.Context, statusID string, boostOfID string, accountID string, boostOfAccountID string) (bool, error) {
	t.Lock()
	defer t.Unlock()

	postIndexEntry := &indexedItemsEntry{
		itemID:           statusID,
		boostOfID:        boostOfID,
		accountID:        accountID,
		boostOfAccountID: boostOfAccountID,
	}

	inserted, err := t.indexedItems.insertIndexed(ctx, postIndexEntry)
	if err != nil {
		return inserted, fmt.Errorf("IndexAndPrepareOne: error inserting indexed: %s", err)
	}

	if inserted {
		if err := t.prepare(ctx, statusID); err != nil {
			return inserted, fmt.Errorf("IndexAndPrepareOne: error preparing: %s", err)
		}
	}

	return inserted, nil
}

func (t *timeline) OldestIndexedItemID(ctx context.Context) (string, error) {
	var id string
	if t.indexedItems == nil || t.indexedItems.data == nil || t.indexedItems.data.Back() == nil {
		// return an empty string if postindex hasn't been initialized yet
		return id, nil
	}

	e := t.indexedItems.data.Back()
	entry, ok := e.Value.(*indexedItemsEntry)
	if !ok {
		return id, errors.New("OldestIndexedItemID: could not parse e as itemIndexEntry")
	}
	return entry.itemID, nil
}

func (t *timeline) NewestIndexedItemID(ctx context.Context) (string, error) {
	var id string
	if t.indexedItems == nil || t.indexedItems.data == nil || t.indexedItems.data.Front() == nil {
		// return an empty string if postindex hasn't been initialized yet
		return id, nil
	}

	e := t.indexedItems.data.Front()
	entry, ok := e.Value.(*indexedItemsEntry)
	if !ok {
		return id, errors.New("NewestIndexedItemID: could not parse e as itemIndexEntry")
	}
	return entry.itemID, nil
}
