/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

const retries = 5

func (t *timeline) Get(ctx context.Context, amount int, maxID string, sinceID string, minID string, prepareNext bool) ([]*apimodel.Status, error) {
	l := logrus.WithFields(logrus.Fields{
		"func":      "Get",
		"accountID": t.accountID,
		"amount":    amount,
		"maxID":     maxID,
		"sinceID":   sinceID,
		"minID":     minID,
	})
	l.Debug("entering get")

	var statuses []*apimodel.Status
	var err error

	// no params are defined to just fetch from the top
	// this is equivalent to a user asking for the top x posts from their timeline
	if maxID == "" && sinceID == "" && minID == "" {
		statuses, err = t.GetXFromTop(ctx, amount)
		// aysnchronously prepare the next predicted query so it's ready when the user asks for it
		if len(statuses) != 0 {
			nextMaxID := statuses[len(statuses)-1].ID
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
	// this is equivalent to a user asking for the next x posts from their timeline, starting from maxID
	if maxID != "" && sinceID == "" {
		attempts := 0
		statuses, err = t.GetXBehindID(ctx, amount, maxID, &attempts)
		// aysnchronously prepare the next predicted query so it's ready when the user asks for it
		if len(statuses) != 0 {
			nextMaxID := statuses[len(statuses)-1].ID
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
	// this is equivalent to a user asking for posts older than x but newer than y
	if maxID != "" && sinceID != "" {
		statuses, err = t.GetXBetweenID(ctx, amount, maxID, minID)
	}
	if maxID != "" && minID != "" {
		statuses, err = t.GetXBetweenID(ctx, amount, maxID, minID)
	}

	// maxID isn't defined, but sinceID || minID are, so take x before
	// this is equivalent to a user asking for posts newer than x (eg., refreshing the top of their timeline)
	if maxID == "" && sinceID != "" {
		statuses, err = t.GetXBeforeID(ctx, amount, sinceID, true)
	}
	if maxID == "" && minID != "" {
		statuses, err = t.GetXBeforeID(ctx, amount, minID, true)
	}

	return statuses, err
}

func (t *timeline) GetXFromTop(ctx context.Context, amount int) ([]*apimodel.Status, error) {
	// make a slice of statuses with the length we need to return
	statuses := make([]*apimodel.Status, 0, amount)

	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
	}

	// make sure we have enough posts prepared to return
	if t.preparedPosts.data.Len() < amount {
		if err := t.PrepareFromTop(ctx, amount); err != nil {
			return nil, err
		}
	}

	// work through the prepared posts from the top and return
	var served int
	for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXFromTop: could not parse e as a preparedPostsEntry")
		}
		statuses = append(statuses, entry.prepared)
		served = served + 1
		if served >= amount {
			break
		}
	}

	return statuses, nil
}

func (t *timeline) GetXBehindID(ctx context.Context, amount int, behindID string, attempts *int) ([]*apimodel.Status, error) {
	l := logrus.WithFields(logrus.Fields{
		"func":     "GetXBehindID",
		"amount":   amount,
		"behindID": behindID,
		"attempts": *attempts,
	})

	newAttempts := *attempts
	newAttempts = newAttempts + 1
	attempts = &newAttempts

	// make a slice of statuses with the length we need to return
	statuses := make([]*apimodel.Status, 0, amount)

	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
	}

	// iterate through the modified list until we hit the mark we're looking for
	var position int
	var behindIDMark *list.Element

findMarkLoop:
	for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
		position = position + 1
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXBehindID: could not parse e as a preparedPostsEntry")
		}

		if entry.statusID <= behindID {
			l.Trace("found behindID mark")
			behindIDMark = e
			break findMarkLoop
		}
	}

	// we didn't find it, so we need to make sure it's indexed and prepared and then try again
	// this can happen when a user asks for really old posts
	if behindIDMark == nil {
		if err := t.PrepareBehind(ctx, behindID, amount); err != nil {
			return nil, fmt.Errorf("GetXBehindID: error preparing behind and including ID %s", behindID)
		}
		oldestID, err := t.OldestPreparedPostID(ctx)
		if err != nil {
			return nil, err
		}
		if oldestID == "" {
			l.Tracef("oldestID is empty so we can't return behindID %s", behindID)
			return statuses, nil
		}
		if oldestID == behindID {
			l.Tracef("given behindID %s is the same as oldestID %s so there's nothing to return behind it", behindID, oldestID)
			return statuses, nil
		}
		if *attempts > retries {
			l.Tracef("exceeded retries looking for behindID %s", behindID)
			return statuses, nil
		}
		l.Trace("trying GetXBehindID again")
		return t.GetXBehindID(ctx, amount, behindID, attempts)
	}

	// make sure we have enough posts prepared behind it to return what we're being asked for
	if t.preparedPosts.data.Len() < amount+position {
		if err := t.PrepareBehind(ctx, behindID, amount); err != nil {
			return nil, err
		}
	}

	// start serving from the entry right after the mark
	var served int
serveloop:
	for e := behindIDMark.Next(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXBehindID: could not parse e as a preparedPostsEntry")
		}

		// serve up to the amount requested
		statuses = append(statuses, entry.prepared)
		served = served + 1
		if served >= amount {
			break serveloop
		}
	}

	return statuses, nil
}

func (t *timeline) GetXBeforeID(ctx context.Context, amount int, beforeID string, startFromTop bool) ([]*apimodel.Status, error) {
	// make a slice of statuses with the length we need to return
	statuses := make([]*apimodel.Status, 0, amount)

	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
	}

	// iterate through the modified list until we hit the mark we're looking for, or as close as possible to it
	var beforeIDMark *list.Element
findMarkLoop:
	for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXBeforeID: could not parse e as a preparedPostsEntry")
		}

		if entry.statusID >= beforeID {
			beforeIDMark = e
		} else {
			break findMarkLoop
		}
	}

	if beforeIDMark == nil {
		return statuses, nil
	}

	var served int

	if startFromTop {
		// start serving from the front/top and keep going until we hit mark or get x amount statuses
	serveloopFromTop:
		for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
			entry, ok := e.Value.(*preparedPostsEntry)
			if !ok {
				return nil, errors.New("GetXBeforeID: could not parse e as a preparedPostsEntry")
			}

			if entry.statusID == beforeID {
				break serveloopFromTop
			}

			// serve up to the amount requested
			statuses = append(statuses, entry.prepared)
			served = served + 1
			if served >= amount {
				break serveloopFromTop
			}
		}
	} else if !startFromTop {
		// start serving from the entry right before the mark
	serveloopFromBottom:
		for e := beforeIDMark.Prev(); e != nil; e = e.Prev() {
			entry, ok := e.Value.(*preparedPostsEntry)
			if !ok {
				return nil, errors.New("GetXBeforeID: could not parse e as a preparedPostsEntry")
			}

			// serve up to the amount requested
			statuses = append(statuses, entry.prepared)
			served = served + 1
			if served >= amount {
				break serveloopFromBottom
			}
		}
	}

	return statuses, nil
}

func (t *timeline) GetXBetweenID(ctx context.Context, amount int, behindID string, beforeID string) ([]*apimodel.Status, error) {
	// make a slice of statuses with the length we need to return
	statuses := make([]*apimodel.Status, 0, amount)

	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
	}

	// iterate through the modified list until we hit the mark we're looking for
	var position int
	var behindIDMark *list.Element
findMarkLoop:
	for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
		position = position + 1
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXBetweenID: could not parse e as a preparedPostsEntry")
		}

		if entry.statusID == behindID {
			behindIDMark = e
			break findMarkLoop
		}
	}

	// we didn't find it
	if behindIDMark == nil {
		return nil, fmt.Errorf("GetXBetweenID: couldn't find status with ID %s", behindID)
	}

	// make sure we have enough posts prepared behind it to return what we're being asked for
	if t.preparedPosts.data.Len() < amount+position {
		if err := t.PrepareBehind(ctx, behindID, amount); err != nil {
			return nil, err
		}
	}

	// start serving from the entry right after the mark
	var served int
serveloop:
	for e := behindIDMark.Next(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXBetweenID: could not parse e as a preparedPostsEntry")
		}

		if entry.statusID == beforeID {
			break serveloop
		}

		// serve up to the amount requested
		statuses = append(statuses, entry.prepared)
		served = served + 1
		if served >= amount {
			break serveloop
		}
	}

	return statuses, nil
}
