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
	"errors"
	"fmt"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (t *timeline) ItemIndexLength(ctx context.Context) int {
	if t.indexedItems == nil || t.indexedItems.data == nil {
		return 0
	}
	return t.indexedItems.data.Len()
}

func (t *timeline) indexBehind(ctx context.Context, itemID string, amount int) error {
	l := log.WithContext(ctx).WithFields(kv.Fields{
		{"amount", amount},
		{"itemID", itemID},
	}...)

	// Lazily init indexed items.
	if t.indexedItems.data == nil {
		t.indexedItems.data = &list.List{}
		t.indexedItems.data.Init()
	}

	// If we're already indexedBehind given itemID
	// by the required amount, we can return nil.
	// First find position of requested itemID.
	var position int
	for e := t.indexedItems.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*indexedItemsEntry)
		if !ok {
			l.Panic(ctx, "could not parse e as indexedItemsEntry")
		}

		if entry.itemID <= itemID {
			// We've found it.
			break
		}

		position++
	}

	// Check if the length of items already indexed
	// satisfies the amount of items required.
	var (
		requestedPosition       = position + amount
		additionalItemsRequired = requestedPosition - t.indexedItems.data.Len()
	)

	if additionalItemsRequired <= 0 {
		// We already have enough indexed behind the mark to
		// satisfy amount, so no need to make more db calls.
		l.Trace("returning nil since we already have enough items indexed")
		return nil
	}

	l.Tracef("%d additional items must be indexed to reach requested position %d", additionalItemsRequired, requestedPosition)

	var (
		additionalItems = make([]Timelineable, 0, additionalItemsRequired)
		offsetID        string
	)

	// Derive offsetID from the last entry in the list to
	// avoid making duplicate calls for entries we already
	// have indexed.
	if e := t.indexedItems.data.Back(); e != nil {
		entry, ok := e.Value.(*indexedItemsEntry)
		if !ok {
			l.Panic(ctx, "could not parse e as indexedItemsEntry")
		}

		offsetID = entry.itemID
	} else {
		// List was empty, so just use itemID.
		offsetID = itemID
	}

	// It's possible that we'll grab items that should be
	// filtered out, so we can't just grab additionalItemsLen
	// once and assume we'll end up with enough items.
	//
	// Instead, we'll try 5 times to grab the items we need.
	for attempts := 0; ; attempts++ {
		innerL := l.WithField("attempts", attempts)

		if attempts > 5 {
			innerL.Trace("max attempts reached while grabbing, breaking")
			break
		}

		innerL.Tracef("trying grab with offsetID %s", offsetID)

		// Check how many items we still need to get.
		remainingRequiredItems := additionalItemsRequired - len(additionalItems)
		if remainingRequiredItems <= 0 {
			innerL.Trace("got all required items while grabbing, breaking")
			break
		}

		items, stop, err := t.grabFunction(ctx, t.accountID, offsetID, "", "", remainingRequiredItems)
		if err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				innerL.Trace("no items left to index while grabbing, breaking")
				break
			}

			// Real error.
			return fmt.Errorf("indexBehind: error calling grabFunction: %w", err)
		}

		if stop {
			innerL.Trace("grab function returned stop, breaking")
			break
		}

		if itemsLen := len(items); itemsLen != 0 {
			// For the next offset, use the last item (we're paging down).
			offsetID = items[itemsLen-1].GetID()
		} else {
			innerL.Trace("grab function returned 0 items, breaking")
			break
		}

		// Filter each item using the caller-provided filter function.
		for _, item := range items {
			shouldIndex, err := t.filterFunction(ctx, t.accountID, item)
			if err != nil {
				innerL.Warnf("error calling filterFunction for item with id %s: %q", item.GetID(), err)
				continue
			}

			if !shouldIndex {
				continue
			}

			additionalItems = append(additionalItems, item)
		}
	}
	l.Trace("left grabloop")

	// Index all the items we got.
	for _, s := range additionalItems {
		if _, err := t.IndexOne(ctx, s.GetID(), s.GetBoostOfID(), s.GetAccountID(), s.GetBoostOfAccountID()); err != nil {
			return fmt.Errorf("indexBehind: error indexing item with id %s: %w", s.GetID(), err)
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

	preparable, err := t.prepareFunction(ctx, t.accountID, statusID)
	if err != nil {
		return false, fmt.Errorf("IndexAndPrepareOne: error preparing: %w", err)
	}

	postIndexEntry := &indexedItemsEntry{
		itemID:           statusID,
		boostOfID:        boostOfID,
		accountID:        accountID,
		boostOfAccountID: boostOfAccountID,
		preparable:       preparable,
	}

	inserted, err := t.indexedItems.insertIndexed(ctx, postIndexEntry)
	if err != nil {
		return false, fmt.Errorf("IndexAndPrepareOne: error inserting indexed: %w", err)
	}

	return inserted, nil
}

func (t *timeline) OldestIndexedItemID(ctx context.Context) string {
	if t.indexedItems == nil || t.indexedItems.data == nil {
		// indexedItems hasnt been initialized yet.
		// Return an empty string.
		return ""
	}

	e := t.indexedItems.data.Back()
	if e == nil {
		// List was empty, return empty string.
		return ""
	}

	entry, ok := e.Value.(*indexedItemsEntry)
	if !ok {
		log.Panic(ctx, "could not parse e as indexedItemsEntry")
	}

	return entry.itemID
}

func (t *timeline) NewestIndexedItemID(ctx context.Context) string {
	if t.indexedItems == nil || t.indexedItems.data == nil {
		// indexedItems hasnt been initialized yet.
		// Return an empty string.
		return ""
	}

	e := t.indexedItems.data.Front()
	if e == nil {
		// List was empty, return empty string.
		return ""
	}

	entry, ok := e.Value.(*indexedItemsEntry)
	if !ok {
		log.Panic(ctx, "could not parse e as indexedItemsEntry")
	}

	return entry.itemID
}
