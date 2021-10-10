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
	"time"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (t *timeline) IndexBefore(ctx context.Context, statusID string, include bool, amount int) error {
	// lazily initialize index if it hasn't been done already
	if t.postIndex.data == nil {
		t.postIndex.data = &list.List{}
		t.postIndex.data.Init()
	}

	filtered := []*gtsmodel.Status{}
	offsetStatus := statusID

	if include {
		// if we have the status with given statusID in the database, include it in the results set as well
		s := &gtsmodel.Status{}
		if err := t.db.GetByID(ctx, statusID, s); err == nil {
			filtered = append(filtered, s)
		}
	}

	i := 0
grabloop:
	for ; len(filtered) < amount && i < 5; i = i + 1 { // try the grabloop 5 times only
		statuses, err := t.db.GetHomeTimeline(ctx, t.accountID, "", "", offsetStatus, amount, false)
		if err != nil {
			if err == db.ErrNoEntries {
				break grabloop // we just don't have enough statuses left in the db so index what we've got and then bail
			}
			return fmt.Errorf("IndexBefore: error getting statuses from db: %s", err)
		}

		for _, s := range statuses {
			timelineable, err := t.filter.StatusHometimelineable(ctx, s, t.account)
			if err != nil {
				continue
			}
			if timelineable {
				filtered = append(filtered, s)
			}
			offsetStatus = s.ID
		}
	}

	for _, s := range filtered {
		if _, err := t.IndexOne(ctx, s.CreatedAt, s.ID, s.BoostOfID, s.AccountID, s.BoostOfAccountID); err != nil {
			return fmt.Errorf("IndexBefore: error indexing status with id %s: %s", s.ID, err)
		}
	}

	return nil
}

func (t *timeline) IndexBehind(ctx context.Context, statusID string, include bool, amount int) error {
	l := logrus.WithFields(logrus.Fields{
		"func":    "IndexBehind",
		"include": include,
		"amount":  amount,
	})

	// lazily initialize index if it hasn't been done already
	if t.postIndex.data == nil {
		t.postIndex.data = &list.List{}
		t.postIndex.data.Init()
	}

	// If we're already indexedBehind given statusID by the required amount, we can return nil.
	// First find position of statusID (or as near as possible).
	var position int
positionLoop:
	for e := t.postIndex.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*postIndexEntry)
		if !ok {
			return errors.New("IndexBehind: could not parse e as a postIndexEntry")
		}

		if entry.statusID <= statusID {
			// we've found it
			break positionLoop
		}
		position++
	}
	// now check if the length of indexed posts exceeds the amount of posts required (position of statusID, plus amount of posts requested after that)
	if t.postIndex.data.Len() > position+amount {
		// we have enough indexed behind already to satisfy amount, so don't need to make db calls
		l.Trace("returning nil since we already have enough posts indexed")
		return nil
	}

	filtered := []*gtsmodel.Status{}
	offsetStatus := statusID

	if include {
		// if we have the status with given statusID in the database, include it in the results set as well
		s := &gtsmodel.Status{}
		if err := t.db.GetByID(ctx, statusID, s); err == nil {
			filtered = append(filtered, s)
		}
	}

	i := 0
grabloop:
	for ; len(filtered) < amount && i < 5; i = i + 1 { // try the grabloop 5 times only
		l.Tracef("entering grabloop; i is %d; len(filtered) is %d", i, len(filtered))
		statuses, err := t.db.GetHomeTimeline(ctx, t.accountID, offsetStatus, "", "", amount, false)
		if err != nil {
			if err == db.ErrNoEntries {
				break grabloop // we just don't have enough statuses left in the db so index what we've got and then bail
			}
			return fmt.Errorf("IndexBehind: error getting statuses from db: %s", err)
		}
		l.Tracef("got %d statuses", len(statuses))

		for _, s := range statuses {
			timelineable, err := t.filter.StatusHometimelineable(ctx, s, t.account)
			if err != nil {
				l.Tracef("status was not hometimelineable: %s", err)
				continue
			}
			if timelineable {
				filtered = append(filtered, s)
			}
			offsetStatus = s.ID
		}
	}
	l.Trace("left grabloop")

	for _, s := range filtered {
		if _, err := t.IndexOne(ctx, s.CreatedAt, s.ID, s.BoostOfID, s.AccountID, s.BoostOfAccountID); err != nil {
			return fmt.Errorf("IndexBehind: error indexing status with id %s: %s", s.ID, err)
		}
	}

	l.Trace("exiting function")
	return nil
}

func (t *timeline) IndexOne(ctx context.Context, statusCreatedAt time.Time, statusID string, boostOfID string, accountID string, boostOfAccountID string) (bool, error) {
	t.Lock()
	defer t.Unlock()

	postIndexEntry := &postIndexEntry{
		statusID:         statusID,
		boostOfID:        boostOfID,
		accountID:        accountID,
		boostOfAccountID: boostOfAccountID,
	}

	return t.postIndex.insertIndexed(postIndexEntry)
}

func (t *timeline) IndexAndPrepareOne(ctx context.Context, statusCreatedAt time.Time, statusID string, boostOfID string, accountID string, boostOfAccountID string) (bool, error) {
	t.Lock()
	defer t.Unlock()

	postIndexEntry := &postIndexEntry{
		statusID:         statusID,
		boostOfID:        boostOfID,
		accountID:        accountID,
		boostOfAccountID: boostOfAccountID,
	}

	inserted, err := t.postIndex.insertIndexed(postIndexEntry)
	if err != nil {
		return inserted, fmt.Errorf("IndexAndPrepareOne: error inserting indexed: %s", err)
	}

	if inserted {
		if err := t.prepare(ctx, statusID); err != nil {
			return inserted, fmt.Errorf("IndexAndPrepareOne: error preparing: %s", err)
		}
	}

	return inserted, nil
}

func (t *timeline) OldestIndexedPostID(ctx context.Context) (string, error) {
	var id string
	if t.postIndex == nil || t.postIndex.data == nil || t.postIndex.data.Back() == nil {
		// return an empty string if postindex hasn't been initialized yet
		return id, nil
	}

	e := t.postIndex.data.Back()
	entry, ok := e.Value.(*postIndexEntry)
	if !ok {
		return id, errors.New("OldestIndexedPostID: could not parse e as a postIndexEntry")
	}
	return entry.statusID, nil
}

func (t *timeline) NewestIndexedPostID(ctx context.Context) (string, error) {
	var id string
	if t.postIndex == nil || t.postIndex.data == nil || t.postIndex.data.Front() == nil {
		// return an empty string if postindex hasn't been initialized yet
		return id, nil
	}

	e := t.postIndex.data.Front()
	entry, ok := e.Value.(*postIndexEntry)
	if !ok {
		return id, errors.New("NewestIndexedPostID: could not parse e as a postIndexEntry")
	}
	return entry.statusID, nil
}
