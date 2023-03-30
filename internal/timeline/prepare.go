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

func (t *timeline) prepareXBetweenIDs(ctx context.Context, amount int, behindID string, beforeID string, frontToBack bool) error {
	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"amount", amount},
			{"behindID", behindID},
			{"beforeID", beforeID},
			{"frontToBack", frontToBack},
		}...)
	l.Trace("entering prepareXBetweenIDs")

	if beforeID >= behindID {
		// This is an impossible situation, we
		// can't prepare anything between these.
		return nil
	}

	if err := t.indexXBetweenIDs(ctx, amount, behindID, beforeID, frontToBack); err != nil {
		// An error here doesn't necessarily mean we
		// can't prepare anything, so log + keep going.
		l.Debugf("error calling prepareXBetweenIDs: %s", err)
	}

	t.Lock()
	defer t.Unlock()

	// Try to prepare everything between (and including) the two points.
	var (
		toPrepare      = make(map[*list.Element]*indexedItemsEntry)
		foundToPrepare int
	)

	if frontToBack {
		// Paging forwards / down.
		for e := t.items.data.Front(); e != nil; e = e.Next() {
			entry := e.Value.(*indexedItemsEntry) //nolint:forcetypeassert

			if entry.itemID > behindID {
				l.Trace("item is too new, continuing")
				continue
			}

			if entry.itemID < beforeID {
				// We've gone beyond the bounds of
				// items we're interested in; stop.
				l.Trace("reached older items, breaking")
				break
			}

			// Only prepare entry if it's not
			// already prepared, save db calls.
			if entry.prepared == nil {
				toPrepare[e] = entry
			}

			foundToPrepare++
			if foundToPrepare >= amount {
				break
			}
		}
	} else {
		// Paging backwards / up.
		for e := t.items.data.Back(); e != nil; e = e.Prev() {
			entry := e.Value.(*indexedItemsEntry) //nolint:forcetypeassert

			if entry.itemID < beforeID {
				l.Trace("item is too old, continuing")
				continue
			}

			if entry.itemID > behindID {
				// We've gone beyond the bounds of
				// items we're interested in; stop.
				l.Trace("reached newer items, breaking")
				break
			}

			if entry.prepared == nil {
				toPrepare[e] = entry
			}

			// Only prepare entry if it's not
			// already prepared, save db calls.
			foundToPrepare++
			if foundToPrepare >= amount {
				break
			}
		}
	}

	for e, entry := range toPrepare {
		prepared, err := t.prepareFunction(ctx, t.accountID, entry.itemID)
		if err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				// ErrNoEntries means something has been deleted,
				// so we'll likely not be able to ever prepare this.
				// This means we can remove it and skip past it.
				l.Debugf("db.ErrNoEntries while trying to prepare %s; will remove from timeline", entry.itemID)
				t.items.data.Remove(e)
			}
			// We've got a proper db error.
			return fmt.Errorf("prepareXBetweenIDs: db error while trying to prepare %s: %w", entry.itemID, err)
		}
		entry.prepared = prepared
	}

	return nil
}
