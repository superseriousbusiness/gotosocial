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
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (t *timeline) PrepareFromTop(ctx context.Context, amount int) error {
	l := log.WithFields(kv.Fields{{"amount", amount}}...)

	// lazily initialize prepared posts if it hasn't been done already
	if t.preparedItems.data == nil {
		t.preparedItems.data = &list.List{}
		t.preparedItems.data.Init()
	}

	// if the postindex is nil, nothing has been indexed yet so index from the highest ID possible
	if t.indexedItems.data == nil {
		l.Debug("postindex.data was nil, indexing behind highest possible ID")
		if err := t.indexBehind(ctx, id.Highest, amount); err != nil {
			return fmt.Errorf("PrepareFromTop: error indexing behind id %s: %s", id.Highest, err)
		}
	}

	l.Trace("entering prepareloop")
	t.Lock()
	defer t.Unlock()
	var prepared int
prepareloop:
	for e := t.indexedItems.data.Front(); e != nil; e = e.Next() {
		if e == nil {
			continue
		}

		entry, ok := e.Value.(*indexedItemsEntry)
		if !ok {
			return errors.New("PrepareFromTop: could not parse e as a postIndexEntry")
		}

		if err := t.prepare(ctx, entry.itemID); err != nil {
			// there's been an error
			if err != db.ErrNoEntries {
				// it's a real error
				return fmt.Errorf("PrepareFromTop: error preparing status with id %s: %s", entry.itemID, err)
			}
			// the status just doesn't exist (anymore) so continue to the next one
			continue
		}

		prepared++
		if prepared == amount {
			// we're done
			l.Trace("leaving prepareloop")
			break prepareloop
		}
	}

	l.Trace("leaving function")
	return nil
}

func (t *timeline) prepareNextQuery(ctx context.Context, amount int, maxID string, sinceID string, minID string) error {
	l := log.WithFields(kv.Fields{
		{"amount", amount},
		{"maxID", maxID},
		{"sinceID", sinceID},
		{"minID", minID},
	}...)

	var err error
	switch {
	case maxID != "" && sinceID == "":
		l.Debug("preparing behind maxID")
		err = t.prepareBehind(ctx, maxID, amount)
	case maxID == "" && sinceID != "":
		l.Debug("preparing before sinceID")
		err = t.prepareBefore(ctx, sinceID, false, amount)
	case maxID == "" && minID != "":
		l.Debug("preparing before minID")
		err = t.prepareBefore(ctx, minID, false, amount)
	}

	return err
}

// prepareBehind instructs the timeline to prepare the next amount of entries for serialization, from position onwards.
// If include is true, then the given item ID will also be prepared, otherwise only entries behind it will be prepared.
func (t *timeline) prepareBehind(ctx context.Context, itemID string, amount int) error {
	// lazily initialize prepared items if it hasn't been done already
	if t.preparedItems.data == nil {
		t.preparedItems.data = &list.List{}
		t.preparedItems.data.Init()
	}

	if err := t.indexBehind(ctx, itemID, amount); err != nil {
		return fmt.Errorf("prepareBehind: error indexing behind id %s: %s", itemID, err)
	}

	// if the itemindex is nil, nothing has been indexed yet so there's nothing to prepare
	if t.indexedItems.data == nil {
		return nil
	}

	var prepared int
	var preparing bool
	t.Lock()
	defer t.Unlock()
prepareloop:
	for e := t.indexedItems.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*indexedItemsEntry)
		if !ok {
			return errors.New("prepareBehind: could not parse e as itemIndexEntry")
		}

		if !preparing {
			// we haven't hit the position we need to prepare from yet
			if entry.itemID == itemID {
				preparing = true
			}
		}

		if preparing {
			if err := t.prepare(ctx, entry.itemID); err != nil {
				// there's been an error
				if err != db.ErrNoEntries {
					// it's a real error
					return fmt.Errorf("prepareBehind: error preparing item with id %s: %s", entry.itemID, err)
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
			if err := t.prepare(ctx, entry.itemID); err != nil {
				// there's been an error
				if err != db.ErrNoEntries {
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

func (t *timeline) prepare(ctx context.Context, itemID string) error {
	// trigger the caller-provided prepare function
	prepared, err := t.prepareFunction(ctx, t.accountID, itemID)
	if err != nil {
		return err
	}

	// shove it in prepared items as a prepared items entry
	preparedItemsEntry := &preparedItemsEntry{
		itemID:           prepared.GetID(),
		boostOfID:        prepared.GetBoostOfID(),
		accountID:        prepared.GetAccountID(),
		boostOfAccountID: prepared.GetBoostOfAccountID(),
		prepared:         prepared,
	}

	return t.preparedItems.insertPrepared(ctx, preparedItemsEntry)
}

// oldestPreparedItemID returns the id of the rearmost (ie., the oldest) prepared item, or an error if something goes wrong.
// If nothing goes wrong but there's no oldest item, an empty string will be returned so make sure to check for this.
func (t *timeline) oldestPreparedItemID(ctx context.Context) (string, error) {
	var id string
	if t.preparedItems == nil || t.preparedItems.data == nil {
		// return an empty string if prepared items hasn't been initialized yet
		return id, nil
	}

	e := t.preparedItems.data.Back()
	if e == nil {
		// return an empty string if there's no back entry (ie., the index list hasn't been initialized yet)
		return id, nil
	}

	entry, ok := e.Value.(*preparedItemsEntry)
	if !ok {
		return id, errors.New("oldestPreparedItemID: could not parse e as a preparedItemsEntry")
	}

	return entry.itemID, nil
}
