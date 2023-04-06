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

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (t *timeline) Remove(ctx context.Context, statusID string) (int, error) {
	l := log.WithContext(ctx).
		WithFields(kv.Fields{
			{"accountTimeline", t.accountID},
			{"statusID", statusID},
		}...)

	t.Lock()
	defer t.Unlock()

	if t.items == nil || t.items.data == nil {
		// Nothing to do.
		return 0, nil
	}

	var toRemove []*list.Element
	for e := t.items.data.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*indexedItemsEntry) // nolint:forcetypeassert

		if entry.itemID != statusID {
			// Not relevant.
			continue
		}

		l.Debug("removing item")
		toRemove = append(toRemove, e)
	}

	for _, e := range toRemove {
		t.items.data.Remove(e)
	}

	return len(toRemove), nil
}

func (t *timeline) RemoveAllByOrBoosting(ctx context.Context, accountID string) (int, error) {
	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"accountTimeline", t.accountID},
			{"accountID", accountID},
		}...)

	t.Lock()
	defer t.Unlock()

	if t.items == nil || t.items.data == nil {
		// Nothing to do.
		return 0, nil
	}

	var toRemove []*list.Element
	for e := t.items.data.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*indexedItemsEntry) // nolint:forcetypeassert

		if entry.accountID != accountID && entry.boostOfAccountID != accountID {
			// Not relevant.
			continue
		}

		l.Debug("removing item")
		toRemove = append(toRemove, e)
	}

	for _, e := range toRemove {
		t.items.data.Remove(e)
	}

	return len(toRemove), nil
}
