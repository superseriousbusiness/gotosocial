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

	"github.com/sirupsen/logrus"
)

const retries = 5

func (t *timeline) Get(ctx context.Context, amount int, maxID string, sinceID string, minID string, prepareNext bool) ([]Preparable, error) {
	l := logrus.WithFields(logrus.Fields{
		"func":      "Get",
		"accountID": t.accountID,
		"amount":    amount,
		"maxID":     maxID,
		"sinceID":   sinceID,
		"minID":     minID,
	})
	l.Debug("entering get")

	var items []Preparable
	var err error

	// no params are defined to just fetch from the top
	// this is equivalent to a user asking for the top x items from their timeline
	if maxID == "" && sinceID == "" && minID == "" {
		items, err = t.GetXFromTop(ctx, amount)
		// aysnchronously prepare the next predicted query so it's ready when the user asks for it
		if len(items) != 0 {
			nextMaxID := items[len(items)-1].GetID()
			if prepareNext {
				// already cache the next query to speed up scrolling
				go func() {
					// use context.Background() because we don't want the query to abort when the request finishes
					if err := t.prepareNextQuery(context.Background(), amount, nextMaxID, "", ""); err != nil {
						l.Errorf("error preparing next query: %s", err)
					}
				}()
			}
		}
	}

	// maxID is defined but sinceID isn't so take from behind
	// this is equivalent to a user asking for the next x items from their timeline, starting from maxID
	if maxID != "" && sinceID == "" {
		attempts := 0
		items, err = t.GetXBehindID(ctx, amount, maxID, &attempts)
		// aysnchronously prepare the next predicted query so it's ready when the user asks for it
		if len(items) != 0 {
			nextMaxID := items[len(items)-1].GetID()
			if prepareNext {
				// already cache the next query to speed up scrolling
				go func() {
					// use context.Background() because we don't want the query to abort when the request finishes
					if err := t.prepareNextQuery(context.Background(), amount, nextMaxID, "", ""); err != nil {
						l.Errorf("error preparing next query: %s", err)
					}
				}()
			}
		}
	}

	// maxID is defined and sinceID || minID are as well, so take a slice between them
	// this is equivalent to a user asking for items older than x but newer than y
	if maxID != "" && sinceID != "" {
		items, err = t.GetXBetweenID(ctx, amount, maxID, minID)
	}
	if maxID != "" && minID != "" {
		items, err = t.GetXBetweenID(ctx, amount, maxID, minID)
	}

	// maxID isn't defined, but sinceID || minID are, so take x before
	// this is equivalent to a user asking for items newer than x (eg., refreshing the top of their timeline)
	if maxID == "" && sinceID != "" {
		items, err = t.GetXBeforeID(ctx, amount, sinceID, true)
	}
	if maxID == "" && minID != "" {
		items, err = t.GetXBeforeID(ctx, amount, minID, true)
	}

	return items, err
}

func (t *timeline) GetXFromTop(ctx context.Context, amount int) ([]Preparable, error) {
	// make a slice of preparedItems with the length we need to return
	preparedItems := make([]Preparable, 0, amount)

	if t.preparedItems.data == nil {
		t.preparedItems.data = &list.List{}
	}

	// make sure we have enough items prepared to return
	if t.preparedItems.data.Len() < amount {
		if err := t.PrepareFromTop(ctx, amount); err != nil {
			return nil, err
		}
	}

	// work through the prepared items from the top and return
	var served int
	for e := t.preparedItems.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedItemsEntry)
		if !ok {
			return nil, errors.New("GetXFromTop: could not parse e as a preparedItemsEntry")
		}
		preparedItems = append(preparedItems, entry.prepared)
		served++
		if served >= amount {
			break
		}
	}

	return preparedItems, nil
}

func (t *timeline) GetXBehindID(ctx context.Context, amount int, behindID string, attempts *int) ([]Preparable, error) {
	l := logrus.WithFields(logrus.Fields{
		"func":     "GetXBehindID",
		"amount":   amount,
		"behindID": behindID,
		"attempts": *attempts,
	})

	newAttempts := *attempts
	newAttempts++
	attempts = &newAttempts

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
			return nil, errors.New("GetXBehindID: could not parse e as a preparedPostsEntry")
		}

		if entry.itemID <= behindID {
			l.Trace("found behindID mark")
			behindIDMark = e
			break findMarkLoop
		}
	}

	// we didn't find it, so we need to make sure it's indexed and prepared and then try again
	// this can happen when a user asks for really old items
	if behindIDMark == nil {
		if err := t.PrepareBehind(ctx, behindID, amount); err != nil {
			return nil, fmt.Errorf("GetXBehindID: error preparing behind and including ID %s", behindID)
		}
		oldestID, err := t.OldestPreparedItemID(ctx)
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
		l.Trace("trying GetXBehindID again")
		return t.GetXBehindID(ctx, amount, behindID, attempts)
	}

	// make sure we have enough items prepared behind it to return what we're being asked for
	if t.preparedItems.data.Len() < amount+position {
		if err := t.PrepareBehind(ctx, behindID, amount); err != nil {
			return nil, err
		}
	}

	// start serving from the entry right after the mark
	var served int
serveloop:
	for e := behindIDMark.Next(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedItemsEntry)
		if !ok {
			return nil, errors.New("GetXBehindID: could not parse e as a preparedPostsEntry")
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

func (t *timeline) GetXBeforeID(ctx context.Context, amount int, beforeID string, startFromTop bool) ([]Preparable, error) {
	// make a slice of items with the length we need to return
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
			return nil, errors.New("GetXBeforeID: could not parse e as a preparedPostsEntry")
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
				return nil, errors.New("GetXBeforeID: could not parse e as a preparedPostsEntry")
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
				return nil, errors.New("GetXBeforeID: could not parse e as a preparedPostsEntry")
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

func (t *timeline) GetXBetweenID(ctx context.Context, amount int, behindID string, beforeID string) ([]Preparable, error) {
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
			return nil, errors.New("GetXBetweenID: could not parse e as a preparedPostsEntry")
		}

		if entry.itemID == behindID {
			behindIDMark = e
			break findMarkLoop
		}
	}

	// we didn't find it
	if behindIDMark == nil {
		return nil, fmt.Errorf("GetXBetweenID: couldn't find item with ID %s", behindID)
	}

	// make sure we have enough items prepared behind it to return what we're being asked for
	if t.preparedItems.data.Len() < amount+position {
		if err := t.PrepareBehind(ctx, behindID, amount); err != nil {
			return nil, err
		}
	}

	// start serving from the entry right after the mark
	var served int
serveloop:
	for e := behindIDMark.Next(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedItemsEntry)
		if !ok {
			return nil, errors.New("GetXBetweenID: could not parse e as a preparedPostsEntry")
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
