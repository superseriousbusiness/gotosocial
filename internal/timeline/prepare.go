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

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (t *timeline) PrepareXFromTop(ctx context.Context, amount int) error {
	// Lazily init prepared items.
	if t.preparedItems.data == nil {
		t.preparedItems.data = &list.List{}
		t.preparedItems.data.Init()
	}

	// Ensure we have enough items indexed from the top.
	if err := t.indexBehind(ctx, id.Highest, amount); err != nil {
		return fmt.Errorf("PrepareFromTop: error indexing behind highest possible id: %w", err)
	}

	t.Lock()
	defer t.Unlock()

	// Iterate through indexedItems from the front,
	// and prepare each entry until we have enough,
	// or until we just hit the end of the list.
	var prepared int
	for e := t.indexedItems.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*indexedItemsEntry)
		if !ok {
			log.Panic(ctx, "could not parse e as indexedItemsEntry")
		}

		if err := t.prepareOne(ctx, entry.itemID); err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				// Likely the status just doesn't
				// exist anymore, so continue.
				continue
			}
			// Real error has occurred.
			return fmt.Errorf("PrepareFromTop: error preparing status with id %s: %w", entry.itemID, err)
		}

		prepared++
		if prepared == amount {
			break
		}
	}

	return nil
}

func (t *timeline) prepareNextQuery(ctx context.Context, amount int, maxID string, sinceID string, minID string) error {
	switch {
	case maxID != "" && sinceID == "":
		return t.prepareXBehindID(ctx, maxID, amount)
	case maxID == "" && sinceID != "":
		return t.prepareBefore(ctx, sinceID, false, amount)
	case maxID == "" && minID != "":
		return t.prepareBefore(ctx, minID, false, amount)
	default:
		return errors.New("prepareNextQuery: reached end of switch statement")
	}
}

// prepareXBehindID instructs the timeline to prepare x amount
// of entries for serialization, from itemID onwards (inclusive).
func (t *timeline) prepareXBehindID(ctx context.Context, itemID string, amount int) error {
	// Lazily init prepared items.
	if t.preparedItems.data == nil {
		t.preparedItems.data = &list.List{}
		t.preparedItems.data.Init()
	}

	// Ensure we have enough items indexed.
	if err := t.indexBehind(ctx, itemID, amount); err != nil {
		return fmt.Errorf("prepareXBehindID: error indexing behind id %s: %w", itemID, err)
	}

	t.Lock()
	defer t.Unlock()

	// Iterate through indexedItems from the front,
	// until we get to itemID, then prepare each entry
	// until we have enough, or until we just hit the
	// end of the list.
	var prepared int
	for e := t.indexedItems.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*indexedItemsEntry)
		if !ok {
			log.Panic(ctx, "could not parse e as indexedItemsEntry")
		}

		if entry.itemID > itemID {
			// ID of this item is too high,
			// just keep iterating.
			continue
		}

		if err := t.prepareOne(ctx, entry.itemID); err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				// Likely the status just doesn't
				// exist anymore, so continue.
				continue
			}
			// Real error has occurred.
			return fmt.Errorf("prepareXBehindID: error preparing status with id %s: %w", entry.itemID, err)
		}

		prepared++
		if prepared == amount {
			break
		}
	}

	return nil
}

func (t *timeline) prepareBefore(ctx context.Context, statusID string, include bool, amount int) error {
	t.Lock()
	defer t.Unlock()

	// lazily initialize prepared posts if it hasn't been done already
	if t.preparedItems.data == nil {
		t.preparedItems.data = &list.List{}
		t.preparedItems.data.Init()
	}

	// if the postindex is nil, nothing has been indexed yet so there's nothing to prepare
	if t.indexedItems.data == nil {
		return nil
	}

	var prepared int
	var preparing bool
prepareloop:
	for e := t.indexedItems.data.Back(); e != nil; e = e.Prev() {
		entry, ok := e.Value.(*indexedItemsEntry)
		if !ok {
			return errors.New("prepareBefore: could not parse e as a postIndexEntry")
		}

		if !preparing {
			// we haven't hit the position we need to prepare from yet
			if entry.itemID == statusID {
				preparing = true
				if !include {
					continue
				}
			}
		}

		if preparing {
			if err := t.prepareOne(ctx, entry.itemID); err != nil {
				// there's been an error
				if !errors.Is(err, db.ErrNoEntries) {
					// it's a real error
					return fmt.Errorf("prepareBefore: error preparing status with id %s: %s", entry.itemID, err)
				}
				// the status just doesn't exist (anymore) so continue to the next one
				continue
			}
			if prepared == amount {
				// we're done
				break prepareloop
			}
			prepared++
		}
	}

	return nil
}

func (t *timeline) prepareOne(ctx context.Context, itemID string) error {
	prepared, err := t.prepareFunction(ctx, t.accountID, itemID)
	if err != nil {
		return err
	}

	preparedItemsEntry := &preparedItemsEntry{
		itemID:           prepared.GetID(),
		boostOfID:        prepared.GetBoostOfID(),
		accountID:        prepared.GetAccountID(),
		boostOfAccountID: prepared.GetBoostOfAccountID(),
		prepared:         prepared,
	}

	return t.preparedItems.insertPrepared(ctx, preparedItemsEntry)
}

func (t *timeline) oldestPreparedItemID(ctx context.Context) string {
	if t.preparedItems == nil || t.preparedItems.data == nil {
		return ""
	}

	e := t.preparedItems.data.Back()
	if e == nil {
		return ""
	}

	entry, ok := e.Value.(*preparedItemsEntry)
	if !ok {
		log.Panic(ctx, "could not parse e as preparedItemsEntry")
	}

	return entry.itemID
}
