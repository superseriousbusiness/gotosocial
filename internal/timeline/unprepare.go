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
	"context"
)

func (t *timeline) Unprepare(ctx context.Context, itemID string) error {
	t.Lock()
	defer t.Unlock()

	if t.items == nil || t.items.data == nil {
		// Nothing to do.
		return nil
	}

	for e := t.items.data.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*indexedItemsEntry) // nolint:forcetypeassert

		if entry.itemID != itemID && entry.boostOfID != itemID {
			// Not relevant.
			continue
		}

		if entry.prepared == nil {
			// It's already unprepared (mood).
			continue
		}

		entry.prepared = nil // <- eat this up please garbage collector nom nom nom
	}

	return nil
}
