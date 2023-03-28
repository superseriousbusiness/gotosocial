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
	"github.com/superseriousbusiness/gotosocial/internal/id"
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
	// trying to "fill in the blanks" between two points,
	// paging either up or down.
	case maxID != "" && sinceID != "":
		items, err = t.getXBetweenIDs(ctx, amount, maxID, sinceID, true)
	case maxID != "" && minID != "":
		items, err = t.getXBetweenIDs(ctx, amount, maxID, minID, false)

	// In the final cases, maxID is not defined, but
	// either sinceID or minID are. This is equivalent to
	// a user either "pulling up" at the top of their timeline
	// to refresh it and check if newer posts have come in.
	//
	// In these calls, we use the highest possible ulid as
	// behindID because we don't have a cap for newest that
	// we're interested in.
	case maxID == "" && sinceID != "":
		items, err = t.getXBetweenIDs(ctx, amount, id.Highest, sinceID, true)
	case maxID == "" && minID != "":
		items, err = t.getXBetweenIDs(ctx, amount, id.Highest, minID, false)
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
		t.preparedItems.data.Init()
	}

	// Make sure we have enough items prepared to return.
	if t.preparedItems.data.Len() < amount {
		if err := t.PrepareXFromTop(ctx, amount); err != nil {
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
			{"attempts", *attempts},
		}...)

	newAttempts := *attempts
	newAttempts++
	attempts = &newAttempts

	// Assume length we need to return.
	items := make([]Preparable, 0, amount)

	// Lazily init prepared items.
	if t.preparedItems.data == nil {
		t.preparedItems.data = &list.List{}
		t.preparedItems.data.Init()
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
		if err := t.prepareXBehindID(ctx, behindID, amount); err != nil {
			return nil, fmt.Errorf("getXBehindID: error preparing behind and including ID %s", behindID)
		}

		oldestID := t.oldestPreparedItemID(ctx)

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
		if err := t.prepareXBehindID(ctx, behindID, amount); err != nil {
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

// getXBetweenIDs returns x amount of items somewhere between (not including) the given IDs.
//
// If frontToBack is true, items will be served paging down from behindID.
// This corresponds to an api call to /timelines/home?max_id=WHATEVER&since_id=WHATEVER
//
// If frontToBack is false, items will be served paging up from beforeID.
// This corresponds to an api call to /timelines/home?max_id=WHATEVER&min_id=WHATEVER
func (t *timeline) getXBetweenIDs(ctx context.Context, amount int, behindID string, beforeID string, frontToBack bool) ([]Preparable, error) {
	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"amount", amount},
			{"behindID", behindID},
			{"beforeID", beforeID},
			{"frontToBack", frontToBack},
		}...)
	l.Trace("entering getXBetweenID")

	// Assume length we need to return.
	items := make([]Preparable, 0, amount)

	// Lazily init prepared items.
	if t.preparedItems.data == nil {
		t.preparedItems.data = &list.List{}
		t.preparedItems.data.Init()
	}

	if beforeID >= behindID {
		// This is an impossible situation, we
		// can't serve anything between these.
		return items, nil
	}

	var (
		beforeIDMark *list.Element
		served       int
		// Our behavior while ranging through the
		// list changes depending on if we're
		// going front-to-back or back-to-front.
		//
		// To avoid checking which one we're doing
		// in each loop iteration, define our range
		// function here outside the loop.
		//
		// The bool indicates to the caller whether
		// iteration should continue (true) or stop
		// (false).
		rangeF func(e *list.Element) bool
	)

	if frontToBack {
		// We're going front-to-back, which means we
		// don't need to look for a mark per se, we
		// just keep serving items until we've reached
		// a point where the items are out of the range
		// we're interested in.
		rangeF = func(e *list.Element) bool {
			entry, ok := e.Value.(*preparedItemsEntry)
			if !ok {
				l.Panic(ctx, "could not parse e as a preparedItemsEntry")
			}

			if entry.itemID >= behindID {
				// ID of this item is too high,
				// just keep iterating.
				l.Trace("item is too new, continuing")
				return true
			}

			if entry.itemID <= beforeID {
				// We've gone as far as we can through
				// the list and reached entries that are
				// now too old for us, stop here.
				l.Trace("reached older items, breaking")
				return false
			}

			items = append(items, entry.prepared)

			served++
			return served < amount
		}
	} else {
		// Iterate through the list from the top, until
		// we reach an item with id smaller than beforeID;
		// ie., an item OLDER than beforeID. At that point,
		// we can stop looking because we're not interested
		// in older entries.
		rangeF = func(e *list.Element) bool {
			entry, ok := e.Value.(*preparedItemsEntry)
			if !ok {
				l.Panic(ctx, "could not parse e as a preparedItemsEntry")
			}

			// Move the mark back one place each loop.
			beforeIDMark = e

			if entry.itemID <= beforeID {
				// We've gone as far as we can through
				// the list and reached entries that are
				// now too old for us, stop here.
				l.Trace("reached older items, breaking")
				return false
			}

			return true
		}
	}

	// Iterate through the list until the function
	// we defined above instructs us to stop.
	for e := t.preparedItems.data.Front(); e != nil; e = e.Next() {
		if !rangeF(e) {
			break
		}
	}

	if frontToBack || beforeIDMark == nil {
		// If we're serving front to back, then
		// items should be populated by now. If
		// we're serving back to front but didn't
		// find any items newer than beforeID,
		// we can just return empty items.
		return items, nil
	}

	// We're serving back to front, so iterate upwards
	// towards the front of the list from the mark we found,
	// until we either get to the front, serve enough
	// items, or reach behindID.
	//
	// To preserve ordering, we need to reverse the slice
	// when we're finished.
	for e := beforeIDMark; e != nil; e = e.Prev() {
		entry, ok := e.Value.(*preparedItemsEntry)
		if !ok {
			l.Panic(ctx, "could not parse e as a preparedItemsEntry")
		}

		if entry.itemID == beforeID {
			// Don't include the beforeID
			// entry itself, just continue.
			l.Trace("entry item ID is equal to beforeID, skipping")
			continue
		}

		if entry.itemID >= behindID {
			// We've reached items that are
			// newer than what we're looking
			// for, just stop here.
			l.Trace("reached newer items, breaking")
			break
		}

		items = append(items, entry.prepared)

		served++
		if served >= amount {
			break
		}
	}

	// Reverse order of items.
	// https://zchee.github.io/golang-wiki/SliceTricks/#reversing
	for l, r := 0, len(items)-1; l < r; l, r = l+1, r-1 {
		items[l], items[r] = items[r], items[l]
	}

	return items, nil
}
