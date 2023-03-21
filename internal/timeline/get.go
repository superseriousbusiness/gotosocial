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
	"time"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

const retries = 5

func (t *timeline) LastGot() time.Time {
	t.Lock()
	defer t.Unlock()
	return t.lastGot
}

func (t *timeline) Get(ctx context.Context, amount int, maxID string, sinceID string, minID string, prepareNext bool) ([]Preparable, error) {
	l := log.WithContext(ctx).
		WithFields(kv.Fields{
			{"accountID", t.accountID},
			{"amount", amount},
			{"maxID", maxID},
			{"sinceID", sinceID},
			{"minID", minID},
		}...)
	l.Trace("entering get and updating t.lastGot")

	// Regardless of what happens below, update the
	// last time Get was called for this timeline.
	t.Lock()
	t.lastGot = time.Now()
	t.Unlock()

	var (
		items []Preparable
		err   error
	)

	switch {
	case maxID == "" && sinceID == "" && minID == "":
		// No params are defined so just fetch from the top.
		// This is equivalent to a user starting to view
		// their timeline from newest -> older posts.
		items, err = t.getXFromTop(ctx, amount)

		// Cache expected next query to speed up scrolling.
		// We use context.Background() because we don't want
		// this to break when the current request finishes.
		if prepareNext && err == nil && len(items) != 0 {
			nextMaxID := items[len(items)-1].GetID()
			go func() {
				if err := t.prepareNextQuery(context.Background(), amount, nextMaxID, "", ""); err != nil {
					l.Errorf("error preparing next query: %s", err)
				}
			}()
		}

	case maxID != "" && sinceID == "" && minID == "":
		// Only maxID is defined, so fetch from maxID onwards.
		// This is equivalent to a user paging further down
		// their timeline from newer -> older posts.
		attempts := 0
		items, err = t.getXBehindID(ctx, amount, maxID, &attempts)

		// Cache expected next query to speed up scrolling.
		// We use context.Background() because we don't want
		// this to break when the current request finishes.
		if prepareNext && err == nil && len(items) != 0 {
			nextMaxID := items[len(items)-1].GetID()
			go func() {
				if err := t.prepareNextQuery(context.Background(), amount, nextMaxID, "", ""); err != nil {
					l.Errorf("error preparing next query: %s", err)
				}
			}()
		}

	// In the next cases, maxID is defined, and so are
	// either sinceID or minID. This is equivalent to
	// a user opening an in-progress timeline and asking
	// for a slice of posts somewhere in the middle, or
	// trying to "fill in the blanks" between two points.
	case maxID != "" && sinceID != "":
		minID = sinceID
		fallthrough
	case maxID != "" && minID != "":
		items, err = t.getXBetweenID(ctx, amount, maxID, minID)

	// In the final cases, maxID is not defined, but
	// either sinceID or minID are. This is equivalent to
	// a user "pulling up" at the top of their timeline
	// to refresh it and check if newer posts have come in.
	case maxID == "" && sinceID != "":
		minID = sinceID
		fallthrough
	case maxID == "" && minID != "":
		items, err = t.getXBeforeID(ctx, amount, minID, true)

	default:
		err = errors.New("Get: switch statement exhausted with no results")
	}

	return items, err
}

// getXFromTop returns x amount of items from the top of the timeline, from newest to oldest.
func (t *timeline) getXFromTop(ctx context.Context, amount int) ([]Preparable, error) {
	// Assume length we need to return.
	items := make([]Preparable, 0, amount)

	// Lazily init prepared items.
	if t.preparedItems.data == nil {
		t.preparedItems.data = &list.List{}
	}

	// Make sure we have enough items prepared to return.
	if t.preparedItems.data.Len() < amount {
		if err := t.PrepareFromTop(ctx, amount); err != nil {
			return nil, err
		}
	}

	// Return prepared items from the top.
	var served int
	for e := t.preparedItems.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedItemsEntry)
		if !ok {
			log.Panic(ctx, "could not parse e as a preparedItemsEntry")
		}

		items = append(items, entry.prepared)

		served++
		if served >= amount {
			break
		}
	}

	return items, nil
}

// getXBehindID returns x amount of items from the given id onwards, from newest to oldest.
// This will NOT include the item with the given ID.
//
// This corresponds to an api call to /timelines/home?max_id=WHATEVER
func (t *timeline) getXBehindID(ctx context.Context, amount int, behindID string, attempts *int) ([]Preparable, error) {
	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"amount", amount},
			{"behindID", behindID},
			{"attempts", attempts},
		}...)

	newAttempts := *attempts
	newAttempts++
	attempts = &newAttempts

	// Assume length we need to return.
	items := make([]Preparable, 0, amount)

	// Lazily init prepared items.
	if t.preparedItems.data == nil {
		t.preparedItems.data = &list.List{}
	}

	var (
		behindIDMark *list.Element
		markPosition int
	)

	// Iterate through the list from the top, until
	// we reach the mark we're looking for. It doesn't
	// have to be precisely equal to behindID, because
	// behindID might have been deleted or something;
	// just get the nearest we can.
	for e := t.preparedItems.data.Front(); e != nil; e = e.Next() {
		markPosition++

		entry, ok := e.Value.(*preparedItemsEntry)
		if !ok {
			l.Panic(ctx, "could not parse e as a preparedItemsEntry")
		}

		if entry.itemID <= behindID {
			l.Trace("found behindID mark")
			behindIDMark = e
			break
		}
	}

	if behindIDMark == nil {
		// We looked through the whole list, but we didn't
		// find the mark yet, so we need to make sure it's
		// indexed and prepared and then try again.
		//
		// This can happen when a user asks for really old
		// items that are no longer prepared because they've
		// been cleaned up.
		if err := t.prepareBehind(ctx, behindID, amount); err != nil {
			return nil, fmt.Errorf("getXBehindID: error preparing behind and including ID %s", behindID)
		}

		oldestID, err := t.oldestPreparedItemID(ctx)
		if err != nil {
			return nil, err
		}

		if oldestID == "" {
			l.Tracef("oldestID is empty so we can't return behindID %s", behindID)
			return items, nil
		}

		if oldestID == behindID {
			l.Tracef("given behindID %s is the same as oldestID %s so there's nothing to return behind it", behindID, oldestID)
			return items, nil
		}

		if *attempts > retries {
			l.Tracef("exceeded retries looking for behindID %s", behindID)
			return items, nil
		}

		l.Trace("trying getXBehindID again")
		return t.getXBehindID(ctx, amount, behindID, attempts)
	}

	// Try to make sure we have enough items prepared
	// *behind* the mark to return requested amount.
	if t.preparedItems.data.Len() < amount+markPosition {
		if err := t.prepareBehind(ctx, behindID, amount); err != nil {
			return nil, fmt.Errorf("getXBehindID: error preparing behind and including ID %s", behindID)
		}
	}

	// Return prepared items *from behindIDMark onwards*.
	var served int
	for e := behindIDMark.Next(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedItemsEntry)
		if !ok {
			log.Panic(ctx, "could not parse e as a preparedItemsEntry")
		}

		items = append(items, entry.prepared)

		served++
		if served >= amount {
			break
		}
	}

	return items, nil
}

// getXBeforeID returns x amount of items up to the given id, from newest to oldest.
// This will NOT include the item with the given ID.
//
// This corresponds to an api call to /timelines/home?since_id=WHATEVER
func (t *timeline) getXBeforeID(ctx context.Context, amount int, beforeID string, startFromTop bool) ([]Preparable, error) {
	aaaaaaaaaa// make a slice of items with the length we need to return
	items := make([]Preparable, 0, amount)

	if t.preparedItems.data == nil {
		t.preparedItems.data = &list.List{}
	}

	// iterate through the modified list until we hit the mark we're looking for, or as close as possible to it
	var beforeIDMark *list.Element
findMarkLoop:
	for e := t.preparedItems.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedItemsEntry)
		if !ok {
			return nil, errors.New("getXBeforeID: could not parse e as a preparedPostsEntry")
		}

		if entry.itemID >= beforeID {
			beforeIDMark = e
		} else {
			break findMarkLoop
		}
	}

	if beforeIDMark == nil {
		return items, nil
	}

	var served int

	if startFromTop {
		// start serving from the front/top and keep going until we hit mark or get x amount items
	serveloopFromTop:
		for e := t.preparedItems.data.Front(); e != nil; e = e.Next() {
			entry, ok := e.Value.(*preparedItemsEntry)
			if !ok {
				return nil, errors.New("getXBeforeID: could not parse e as a preparedPostsEntry")
			}

			if entry.itemID == beforeID {
				break serveloopFromTop
			}

			// serve up to the amount requested
			items = append(items, entry.prepared)
			served++
			if served >= amount {
				break serveloopFromTop
			}
		}
	} else if !startFromTop {
		// start serving from the entry right before the mark
	serveloopFromBottom:
		for e := beforeIDMark.Prev(); e != nil; e = e.Prev() {
			entry, ok := e.Value.(*preparedItemsEntry)
			if !ok {
				return nil, errors.New("getXBeforeID: could not parse e as a preparedPostsEntry")
			}

			// serve up to the amount requested
			items = append(items, entry.prepared)
			served++
			if served >= amount {
				break serveloopFromBottom
			}
		}
	}

	return items, nil
}

// getXBetweenID returns x amount of items from the given maxID, up to the given id, from newest to oldest.
// This will NOT include the item with the given IDs.
//
// This corresponds to an api call to /timelines/home?since_id=WHATEVER&max_id=WHATEVER_ELSE
func (t *timeline) getXBetweenID(ctx context.Context, amount int, behindID string, beforeID string) ([]Preparable, error) {
	// make a slice of items with the length we need to return
	items := make([]Preparable, 0, amount)

	if t.preparedItems.data == nil {
		t.preparedItems.data = &list.List{}
	}

	// iterate through the modified list until we hit the mark we're looking for
	var position int
	var behindIDMark *list.Element
findMarkLoop:
	for e := t.preparedItems.data.Front(); e != nil; e = e.Next() {
		position++
		entry, ok := e.Value.(*preparedItemsEntry)
		if !ok {
			return nil, errors.New("getXBetweenID: could not parse e as a preparedPostsEntry")
		}

		if entry.itemID == behindID {
			behindIDMark = e
			break findMarkLoop
		}
	}

	// we didn't find it
	if behindIDMark == nil {
		return nil, fmt.Errorf("getXBetweenID: couldn't find item with ID %s", behindID)
	}

	// make sure we have enough items prepared behind it to return what we're being asked for
	if t.preparedItems.data.Len() < amount+position {
		if err := t.prepareBehind(ctx, behindID, amount); err != nil {
			return nil, err
		}
	}

	// start serving from the entry right after the mark
	var served int
serveloop:
	for e := behindIDMark.Next(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedItemsEntry)
		if !ok {
			return nil, errors.New("getXBetweenID: could not parse e as a preparedPostsEntry")
		}

		if entry.itemID == beforeID {
			break serveloop
		}

		// serve up to the amount requested
		items = append(items, entry.prepared)
		served++
		if served >= amount {
			break serveloop
		}
	}

	return items, nil
}
