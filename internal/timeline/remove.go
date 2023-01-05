/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (t *timeline) Remove(ctx context.Context, statusID string) (int, error) {
	l := log.WithFields(kv.Fields{
		{"accountTimeline", t.accountID},
		{"statusID", statusID},
	}...)

	t.Lock()
	defer t.Unlock()
	var removed int

	// remove entr(ies) from the post index
	removeIndexes := []*list.Element{}
	if t.indexedItems != nil && t.indexedItems.data != nil {
		for e := t.indexedItems.data.Front(); e != nil; e = e.Next() {
			entry, ok := e.Value.(*indexedItemsEntry)
			if !ok {
				return removed, errors.New("Remove: could not parse e as a postIndexEntry")
			}
			if entry.itemID == statusID {
				l.Debug("found status in postIndex")
				removeIndexes = append(removeIndexes, e)
			}
		}
	}
	for _, e := range removeIndexes {
		t.indexedItems.data.Remove(e)
		removed++
	}

	// remove entr(ies) from prepared posts
	removePrepared := []*list.Element{}
	if t.preparedItems != nil && t.preparedItems.data != nil {
		for e := t.preparedItems.data.Front(); e != nil; e = e.Next() {
			entry, ok := e.Value.(*preparedItemsEntry)
			if !ok {
				return removed, errors.New("Remove: could not parse e as a preparedPostsEntry")
			}
			if entry.itemID == statusID {
				l.Debug("found status in preparedPosts")
				removePrepared = append(removePrepared, e)
			}
		}
	}
	for _, e := range removePrepared {
		t.preparedItems.data.Remove(e)
		removed++
	}

	l.Debugf("removed %d entries", removed)
	return removed, nil
}

func (t *timeline) RemoveAllBy(ctx context.Context, accountID string) (int, error) {
	l := log.WithFields(kv.Fields{
		{"accountTimeline", t.accountID},
		{"accountID", accountID},
	}...)

	t.Lock()
	defer t.Unlock()
	var removed int

	// remove entr(ies) from the post index
	removeIndexes := []*list.Element{}
	if t.indexedItems != nil && t.indexedItems.data != nil {
		for e := t.indexedItems.data.Front(); e != nil; e = e.Next() {
			entry, ok := e.Value.(*indexedItemsEntry)
			if !ok {
				return removed, errors.New("Remove: could not parse e as a postIndexEntry")
			}
			if entry.accountID == accountID || entry.boostOfAccountID == accountID {
				l.Debug("found status in postIndex")
				removeIndexes = append(removeIndexes, e)
			}
		}
	}
	for _, e := range removeIndexes {
		t.indexedItems.data.Remove(e)
		removed++
	}

	// remove entr(ies) from prepared posts
	removePrepared := []*list.Element{}
	if t.preparedItems != nil && t.preparedItems.data != nil {
		for e := t.preparedItems.data.Front(); e != nil; e = e.Next() {
			entry, ok := e.Value.(*preparedItemsEntry)
			if !ok {
				return removed, errors.New("Remove: could not parse e as a preparedPostsEntry")
			}
			if entry.accountID == accountID || entry.boostOfAccountID == accountID {
				l.Debug("found status in preparedPosts")
				removePrepared = append(removePrepared, e)
			}
		}
	}
	for _, e := range removePrepared {
		t.preparedItems.data.Remove(e)
		removed++
	}

	l.Debugf("removed %d entries", removed)
	return removed, nil
}
