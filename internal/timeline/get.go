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
	"time"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (t *timeline) LastGot() time.Time {
	t.Lock()
	defer t.Unlock()
	return t.lastGot
}

func (t *timeline) Get(ctx context.Context, amount int, maxID string, sinceID string, minID string, prepareNext bool) ([]Preparable, error) {
	l := log.WithContext(ctx).
		WithFields(kv.Fields{
			{"accountID", t.timelineID},
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
		items, err = t.getXBetweenIDs(ctx, amount, id.Highest, id.Lowest, true)

		// Cache expected next query to speed up scrolling.
		// Assume the user will be scrolling downwards from
		// the final ID in items.
		if prepareNext && err == nil && len(items) != 0 {
			nextMaxID := items[len(items)-1].GetID()
			t.prepareNextQuery(amount, nextMaxID, "", "")
		}

	case maxID != "" && sinceID == "" && minID == "":
		// Only maxID is defined, so fetch from maxID onwards.
		// This is equivalent to a user paging further down
		// their timeline from newer -> older posts.
		items, err = t.getXBetweenIDs(ctx, amount, maxID, id.Lowest, true)

		// Cache expected next query to speed up scrolling.
		// Assume the user will be scrolling downwards from
		// the final ID in items.
		if prepareNext && err == nil && len(items) != 0 {
			nextMaxID := items[len(items)-1].GetID()
			t.prepareNextQuery(amount, nextMaxID, "", "")
		}

	// In the next cases, maxID is defined, and so are
	// either sinceID or minID. This is equivalent to
	// a user opening an in-progress timeline and asking
	// for a slice of posts somewhere in the middle, or
	// trying to "fill in the blanks" between two points,
	// paging either up or down.
	case maxID != "" && sinceID != "":
		items, err = t.getXBetweenIDs(ctx, amount, maxID, sinceID, true)

		// Cache expected next query to speed up scrolling.
		// We can assume the caller is scrolling downwards.
		// Guess id.Lowest as sinceID, since we don't actually
		// know what the next sinceID would be.
		if prepareNext && err == nil && len(items) != 0 {
			nextMaxID := items[len(items)-1].GetID()
			t.prepareNextQuery(amount, nextMaxID, id.Lowest, "")
		}

	case maxID != "" && minID != "":
		items, err = t.getXBetweenIDs(ctx, amount, maxID, minID, false)

		// Cache expected next query to speed up scrolling.
		// We can assume the caller is scrolling upwards.
		// Guess id.Highest as maxID, since we don't actually
		// know what the next maxID would be.
		if prepareNext && err == nil && len(items) != 0 {
			prevMinID := items[0].GetID()
			t.prepareNextQuery(amount, id.Highest, "", prevMinID)
		}

	// In the final cases, maxID is not defined, but
	// either sinceID or minID are. This is equivalent to
	// a user either "pulling up" at the top of their timeline
	// to refresh it and check if newer posts have come in, or
	// trying to scroll upwards from an old post to see what
	// they missed since then.
	//
	// In these calls, we use the highest possible ulid as
	// behindID because we don't have a cap for newest that
	// we're interested in.
	case maxID == "" && sinceID != "":
		items, err = t.getXBetweenIDs(ctx, amount, id.Highest, sinceID, true)

		// We can't cache an expected next query for this one,
		// since presumably the caller is at the top of their
		// timeline already.

	case maxID == "" && minID != "":
		items, err = t.getXBetweenIDs(ctx, amount, id.Highest, minID, false)

		// Cache expected next query to speed up scrolling.
		// We can assume the caller is scrolling upwards.
		// Guess id.Highest as maxID, since we don't actually
		// know what the next maxID would be.
		if prepareNext && err == nil && len(items) != 0 {
			prevMinID := items[0].GetID()
			t.prepareNextQuery(amount, id.Highest, "", prevMinID)
		}

	default:
		err = gtserror.New("switch statement exhausted with no results")
	}

	return items, err
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

	if beforeID >= behindID {
		// This is an impossible situation, we
		// can't serve anything between these.
		return items, nil
	}

	// Try to ensure we have enough items prepared.
	if err := t.prepareXBetweenIDs(ctx, amount, behindID, beforeID, frontToBack); err != nil {
		// An error here doesn't necessarily mean we
		// can't serve anything, so log + keep going.
		l.Debugf("error calling prepareXBetweenIDs: %s", err)
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
		rangeF func(e *list.Element) (bool, error)
		// If we get certain errors on entries as we're
		// looking through, we might want to cheekily
		// remove their elements from the timeline.
		// Everything added to this slice will be removed.
		removeElements = []*list.Element{}
	)

	defer func() {
		for _, e := range removeElements {
			t.items.data.Remove(e)
		}
	}()

	if frontToBack {
		// We're going front-to-back, which means we
		// don't need to look for a mark per se, we
		// just keep serving items until we've reached
		// a point where the items are out of the range
		// we're interested in.
		rangeF = func(e *list.Element) (bool, error) {
			entry := e.Value.(*indexedItemsEntry)

			if entry.itemID >= behindID {
				// ID of this item is too high,
				// just keep iterating.
				l.Trace("item is too new, continuing")
				return true, nil
			}

			if entry.itemID <= beforeID {
				// We've gone as far as we can through
				// the list and reached entries that are
				// now too old for us, stop here.
				l.Trace("reached older items, breaking")
				return false, nil
			}

			l.Trace("entry is just right")

			if entry.prepared == nil {
				// Whoops, this entry isn't prepared yet; some
				// race condition? That's OK, we can do it now.
				prepared, err := t.prepareFunction(ctx, t.timelineID, entry.itemID)
				if err != nil {
					if errors.Is(err, statusfilter.ErrHideStatus) {
						// This item has been filtered out by the requesting user's filters.
						// Remove it and skip past it.
						removeElements = append(removeElements, e)
						return true, nil
					}
					if errors.Is(err, db.ErrNoEntries) {
						// ErrNoEntries means something has been deleted,
						// so we'll likely not be able to ever prepare this.
						// This means we can remove it and skip past it.
						l.Debugf("db.ErrNoEntries while trying to prepare %s; will remove from timeline", entry.itemID)
						removeElements = append(removeElements, e)
						return true, nil
					}
					// We've got a proper db error.
					err = gtserror.Newf("db error while trying to prepare %s: %w", entry.itemID, err)
					return false, err
				}
				entry.prepared = prepared
			}

			items = append(items, entry.prepared)

			served++
			return served < amount, nil
		}
	} else {
		// Iterate through the list from the top, until
		// we reach an item with id smaller than beforeID;
		// ie., an item OLDER than beforeID. At that point,
		// we can stop looking because we're not interested
		// in older entries.
		rangeF = func(e *list.Element) (bool, error) {
			// Move the mark back one place each loop.
			beforeIDMark = e

			if entry := e.Value.(*indexedItemsEntry); entry.itemID <= beforeID {
				// We've gone as far as we can through
				// the list and reached entries that are
				// now too old for us, stop here.
				l.Trace("reached older items, breaking")
				return false, nil
			}

			return true, nil
		}
	}

	// Iterate through the list until the function
	// we defined above instructs us to stop.
	for e := t.items.data.Front(); e != nil; e = e.Next() {
		keepGoing, err := rangeF(e)
		if err != nil {
			return nil, err
		}

		if !keepGoing {
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
		entry := e.Value.(*indexedItemsEntry)

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

		if entry.prepared == nil {
			// Whoops, this entry isn't prepared yet; some
			// race condition? That's OK, we can do it now.
			prepared, err := t.prepareFunction(ctx, t.timelineID, entry.itemID)
			if err != nil {
				if errors.Is(err, statusfilter.ErrHideStatus) {
					// This item has been filtered out by the requesting user's filters.
					// Remove it and skip past it.
					removeElements = append(removeElements, e)
					continue
				}
				if errors.Is(err, db.ErrNoEntries) {
					// ErrNoEntries means something has been deleted,
					// so we'll likely not be able to ever prepare this.
					// This means we can remove it and skip past it.
					l.Debugf("db.ErrNoEntries while trying to prepare %s; will remove from timeline", entry.itemID)
					removeElements = append(removeElements, e)
					continue
				}
				// We've got a proper db error.
				err = gtserror.Newf("db error while trying to prepare %s: %w", entry.itemID, err)
				return nil, err
			}
			entry.prepared = prepared
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

func (t *timeline) prepareNextQuery(amount int, maxID string, sinceID string, minID string) {
	var (
		// We explicitly use context.Background() rather than
		// accepting a context param because we don't want this
		// to stop/break when the calling context finishes.
		ctx = context.Background()
		err error
	)

	// Always perform this async so caller doesn't have to wait.
	go func() {
		switch {
		case maxID == "" && sinceID == "" && minID == "":
			err = t.prepareXBetweenIDs(ctx, amount, id.Highest, id.Lowest, true)
		case maxID != "" && sinceID == "" && minID == "":
			err = t.prepareXBetweenIDs(ctx, amount, maxID, id.Lowest, true)
		case maxID != "" && sinceID != "":
			err = t.prepareXBetweenIDs(ctx, amount, maxID, sinceID, true)
		case maxID != "" && minID != "":
			err = t.prepareXBetweenIDs(ctx, amount, maxID, minID, false)
		case maxID == "" && sinceID != "":
			err = t.prepareXBetweenIDs(ctx, amount, id.Highest, sinceID, true)
		case maxID == "" && minID != "":
			err = t.prepareXBetweenIDs(ctx, amount, id.Highest, minID, false)
		default:
			err = gtserror.New("switch statement exhausted with no results")
		}

		if err != nil {
			log.
				WithContext(ctx).
				WithFields(kv.Fields{
					{"amount", amount},
					{"maxID", maxID},
					{"sinceID", sinceID},
					{"minID", minID},
				}...).
				Warnf("error preparing next query: %s", err)
		}
	}()
}
