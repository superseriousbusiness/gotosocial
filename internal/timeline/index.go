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

	// Start by mapping out the list so we know what
	// we have to do. Depending on the current state
	// of the list we might not have to do *anything*.
	var (
		position         int
		listLen          = t.items.data.Len()
		behindIDPosition int
		beforeIDPosition int
	)

	for e := t.items.data.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*indexedItemsEntry) //nolint:forcetypeassert

		position++

		if entry.itemID > behindID {
			l.Trace("item is too new, continuing")
			continue
		}

		if behindIDPosition == 0 {
			// Gone far enough through the list
			// and found our behindID mark.
			// We only need to set this once.
			l.Tracef("found behindID mark %s at position %d", entry.itemID, position)
			behindIDPosition = position
		}

		if entry.itemID >= beforeID {
			// Push the beforeID mark back
			// one place every iteration.
			l.Tracef("setting beforeID mark %s at position %d", entry.itemID, position)
			beforeIDPosition = position
		}

		if entry.itemID <= beforeID {
			// We've gone beyond the bounds of
			// items we're interested in; stop.
			l.Trace("reached older items, breaking")
			break
		}
	}

	// We can now figure out if we need to make db calls.
	var grabMore bool
	switch {
	case listLen < amount:
		// The whole list is shorter than the
		// amount we're being asked to return,
		// make up the difference.
		grabMore = true
		amount -= listLen
	case beforeIDPosition-behindIDPosition < amount:
		// Not enough items between behindID and
		// beforeID to return amount required,
		// try to get more.
		grabMore = true
	}

	if !grabMore {
		// We're good!
		return nil
	}

	// Fetch additional items.
	items, err := t.grab(ctx, amount, behindID, beforeID, frontToBack)
	if err != nil {
		return err
	}

	// Index all the items we got. We already have
	// a lock on the timeline, so don't call IndexOne
	// here, since that will also try to get a lock!
	for _, item := range items {
		entry := &indexedItemsEntry{
			itemID:           item.GetID(),
			boostOfID:        item.GetBoostOfID(),
			accountID:        item.GetAccountID(),
			boostOfAccountID: item.GetBoostOfAccountID(),
		}

		if _, err := t.items.insertIndexed(ctx, entry); err != nil {
			return fmt.Errorf("error inserting entry with itemID %s into index: %w", entry.itemID, err)
		}
	}

	return nil
}

// grab wraps the timeline's grabFunction in paging + filtering logic.
func (t *timeline) grab(ctx context.Context, amount int, behindID string, beforeID string, frontToBack bool) ([]Timelineable, error) {
	var (
		sinceID  string
		minID    string
		grabbed  int
		maxID    = behindID
		filtered = make([]Timelineable, 0, amount)
	)

	if frontToBack {
		sinceID = beforeID
	} else {
		minID = beforeID
	}

	for attempts := 0; attempts < 5; attempts++ {
		if grabbed >= amount {
			// We got everything we needed.
			break
		}

		items, stop, err := t.grabFunction(
			ctx,
			t.accountID,
			maxID,
			sinceID,
			minID,
			// Don't grab more than we need to.
			amount-grabbed,
		)

		if err != nil {
			// Grab function already checks for
			// db.ErrNoEntries, so if an error
			// is returned then it's a real one.
			return nil, err
		}

		if stop || len(items) == 0 {
			// No items left.
			break
		}

		// Set next query parameters.
		if frontToBack {
			// Page down.
			maxID = items[len(items)-1].GetID()
			if maxID <= beforeID {
				// Can't go any further.
				break
			}
		} else {
			// Page up.
			minID = items[0].GetID()
			if minID >= behindID {
				// Can't go any further.
				break
			}
		}

		for _, item := range items {
			ok, err := t.filterFunction(ctx, t.accountID, item)
			if err != nil {
				if !errors.Is(err, db.ErrNoEntries) {
					// Real error here.
					return nil, err
				}
				log.Warnf(ctx, "errNoEntries while filtering item %s: %s", item.GetID(), err)
				continue
			}

			if ok {
				filtered = append(filtered, item)
				grabbed++ // count this as grabbed
			}
		}
	}

	return filtered, nil
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

func (t *timeline) Len() int {
	t.Lock()
	defer t.Unlock()

	if t.items == nil || t.items.data == nil {
		// indexedItems hasnt been initialized yet.
		return 0
	}

	return t.items.data.Len()
}

func (t *timeline) OldestIndexedItemID() string {
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

	return e.Value.(*indexedItemsEntry).itemID //nolint:forcetypeassert
}
