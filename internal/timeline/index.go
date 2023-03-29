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

func (t *timeline) indexXBetweenIDs(ctx context.Context, amount int, behindID string, beforeID string, frontToBack bool) error {
	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"amount", amount},
			{"behindID", behindID},
			{"beforeID", beforeID},
			{"frontToBack", frontToBack},
		}...)
	l.Trace("entering indexXBetweenIDs")
	
	if beforeID >= behindID {
		// This is an impossible situation, we
		// can't index anything between these.
		return nil
	}

	t.Lock()
	defer t.Unlock()
	
	// Lazily init indexed items.
	if t.items.data == nil {
		t.items.data = &list.List{}
		t.items.data.Init()
	}

	// If we're going front to back, and we already
	// indexed behind behindID by the required amount,
	// we can just return nil.
	var position int
	for e := t.items.data.Front(); e != nil; e = e.Next() {
		if entry := e.Value.(*indexedItemsEntry); entry.itemID <= itemID {
			// We've found it.
			break
		}

		position++
	}

	// Check if the length of items already indexed
	// satisfies the amount of items required.
	var (
		requestedPosition       = position + amount
		additionalItemsRequired = requestedPosition - t.items.data.Len()
	)

	if additionalItemsRequired <= 0 {
		// We already have enough indexed behind the mark to
		// satisfy amount, so no need to make more db calls.
		l.Trace("returning nil since we already have enough items indexed")
		return nil
	}

	l.Tracef("%d additional items must be indexed to reach requested position %d", additionalItemsRequired, requestedPosition)

	var (
		itemsToIndex = make([]Timelineable, 0, additionalItemsRequired)
		offsetID        string
	)

	// Derive offsetID from the last entry in the list to
	// avoid making duplicate calls for entries we already
	// have indexed.
	if e := t.items.data.Back(); e != nil {
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
		remainingRequiredItems := additionalItemsRequired - len(itemsToIndex)
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

			itemsToIndex = append(itemsToIndex, item)
		}
	}
	l.Trace("left grabloop")

	// Index all the items we got.
	// We already have a lock on the timeline,
	// so don't call IndexOne here, since that
	// will also try to get a lock!
	for _, item := range itemsToIndex {
		entry := &indexedItemsEntry{
			itemID:           item.GetID(),
			boostOfID:        item.GetBoostOfID(),
			accountID:        item.GetAccountID(),
			boostOfAccountID: item.GetBoostOfAccountID(),
		}

		if _, err := t.items.insertIndexed(ctx, entry); err != nil {

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

	return t.items.insertIndexed(ctx, postIndexEntry)
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

	if inserted, err := t.items.insertIndexed(ctx, postIndexEntry); err != nil {
		return false, fmt.Errorf("IndexAndPrepareOne: error inserting indexed: %w", err)
	} else if !inserted {
		// Entry wasn't inserted, so
		// don't bother preparing it.
		return false, nil
	}

	preparable, err := t.prepareFunction(ctx, t.accountID, statusID)
	if err != nil {
		return true, fmt.Errorf("IndexAndPrepareOne: error preparing: %w", err)
	}
	postIndexEntry.prepared = preparable

	return true, nil
}

func (t *timeline) ItemIndexLength(ctx context.Context) int {
	t.Lock()
	defer t.Unlock()

	if t.items == nil || t.items.data == nil {
		// indexedItems hasnt been initialized yet.
		return 0
	}

	return t.items.data.Len()
}

func (t *timeline) OldestIndexedItemID(ctx context.Context) string {
	t.Lock()
	defer t.Unlock()

	if t.items == nil || t.items.data == nil {
		// indexedItems hasnt been initialized yet.
		return ""
	}

	e := t.items.data.Back()
	if e == nil {
		// List was empty.
		return ""
	}

	return e.Value.(*indexedItemsEntry).itemID
}

func (t *timeline) NewestIndexedItemID(ctx context.Context) string {
	t.Lock()
	defer t.Unlock()

	if t.items == nil || t.items.data == nil {
		// indexedItems hasnt been initialized yet.
		return ""
	}

	e := t.items.data.Front()
	if e == nil {
		// List was empty.
		return ""
	}

	return e.Value.(*indexedItemsEntry).itemID
}
