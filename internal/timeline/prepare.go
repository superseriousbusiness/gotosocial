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
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (t *timeline) prepareNextQuery(ctx context.Context, amount int, maxID string, sinceID string, minID string) error {
	l := logrus.WithFields(logrus.Fields{
		"func":    "prepareNextQuery",
		"amount":  amount,
		"maxID":   maxID,
		"sinceID": sinceID,
		"minID":   minID,
	})

	var err error

	// maxID is defined but sinceID isn't so take from behind
	if maxID != "" && sinceID == "" {
		l.Debug("preparing behind maxID")
		err = t.PrepareBehind(ctx, maxID, amount)
	}

	// maxID isn't defined, but sinceID || minID are, so take x before
	if maxID == "" && sinceID != "" {
		l.Debug("preparing before sinceID")
		err = t.PrepareBefore(ctx, sinceID, false, amount)
	}
	if maxID == "" && minID != "" {
		l.Debug("preparing before minID")
		err = t.PrepareBefore(ctx, minID, false, amount)
	}

	return err
}

func (t *timeline) PrepareBehind(ctx context.Context, statusID string, amount int) error {
	// lazily initialize prepared posts if it hasn't been done already
	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
		t.preparedPosts.data.Init()
	}

	if err := t.IndexBehind(ctx, statusID, true, amount); err != nil {
		return fmt.Errorf("PrepareBehind: error indexing behind id %s: %s", statusID, err)
	}

	// if the postindex is nil, nothing has been indexed yet so there's nothing to prepare
	if t.postIndex.data == nil {
		return nil
	}

	var prepared int
	var preparing bool
	t.Lock()
	defer t.Unlock()
prepareloop:
	for e := t.postIndex.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*postIndexEntry)
		if !ok {
			return errors.New("PrepareBehind: could not parse e as a postIndexEntry")
		}

		if !preparing {
			// we haven't hit the position we need to prepare from yet
			if entry.statusID == statusID {
				preparing = true
			}
		}

		if preparing {
			if err := t.prepare(ctx, entry.statusID); err != nil {
				// there's been an error
				if err != db.ErrNoEntries {
					// it's a real error
					return fmt.Errorf("PrepareBehind: error preparing status with id %s: %s", entry.statusID, err)
				}
				// the status just doesn't exist (anymore) so continue to the next one
				continue
			}
			if prepared == amount {
				// we're done
				break prepareloop
			}
			prepared = prepared + 1
		}
	}

	return nil
}

func (t *timeline) PrepareBefore(ctx context.Context, statusID string, include bool, amount int) error {
	t.Lock()
	defer t.Unlock()

	// lazily initialize prepared posts if it hasn't been done already
	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
		t.preparedPosts.data.Init()
	}

	// if the postindex is nil, nothing has been indexed yet so there's nothing to prepare
	if t.postIndex.data == nil {
		return nil
	}

	var prepared int
	var preparing bool
prepareloop:
	for e := t.postIndex.data.Back(); e != nil; e = e.Prev() {
		entry, ok := e.Value.(*postIndexEntry)
		if !ok {
			return errors.New("PrepareBefore: could not parse e as a postIndexEntry")
		}

		if !preparing {
			// we haven't hit the position we need to prepare from yet
			if entry.statusID == statusID {
				preparing = true
				if !include {
					continue
				}
			}
		}

		if preparing {
			if err := t.prepare(ctx, entry.statusID); err != nil {
				// there's been an error
				if err != db.ErrNoEntries {
					// it's a real error
					return fmt.Errorf("PrepareBefore: error preparing status with id %s: %s", entry.statusID, err)
				}
				// the status just doesn't exist (anymore) so continue to the next one
				continue
			}
			if prepared == amount {
				// we're done
				break prepareloop
			}
			prepared = prepared + 1
		}
	}

	return nil
}

func (t *timeline) PrepareFromTop(ctx context.Context, amount int) error {
	l := logrus.WithFields(logrus.Fields{
		"func":   "PrepareFromTop",
		"amount": amount,
	})

	// lazily initialize prepared posts if it hasn't been done already
	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
		t.preparedPosts.data.Init()
	}

	// if the postindex is nil, nothing has been indexed yet so index from the highest ID possible
	if t.postIndex.data == nil {
		l.Debug("postindex.data was nil, indexing behind highest possible ID")
		if err := t.IndexBehind(ctx, "ZZZZZZZZZZZZZZZZZZZZZZZZZZ", false, amount); err != nil {
			return fmt.Errorf("PrepareFromTop: error indexing behind id %s: %s", "ZZZZZZZZZZZZZZZZZZZZZZZZZZ", err)
		}
	}

	l.Trace("entering prepareloop")
	t.Lock()
	defer t.Unlock()
	var prepared int
prepareloop:
	for e := t.postIndex.data.Front(); e != nil; e = e.Next() {
		if e == nil {
			continue
		}

		entry, ok := e.Value.(*postIndexEntry)
		if !ok {
			return errors.New("PrepareFromTop: could not parse e as a postIndexEntry")
		}

		if err := t.prepare(ctx, entry.statusID); err != nil {
			// there's been an error
			if err != db.ErrNoEntries {
				// it's a real error
				return fmt.Errorf("PrepareFromTop: error preparing status with id %s: %s", entry.statusID, err)
			}
			// the status just doesn't exist (anymore) so continue to the next one
			continue
		}

		prepared = prepared + 1
		if prepared == amount {
			// we're done
			l.Trace("leaving prepareloop")
			break prepareloop
		}
	}

	l.Trace("leaving function")
	return nil
}

func (t *timeline) prepare(ctx context.Context, statusID string) error {

	// start by getting the status out of the database according to its indexed ID
	gtsStatus := &gtsmodel.Status{}
	if err := t.db.GetByID(ctx, statusID, gtsStatus); err != nil {
		return err
	}

	// if the account pointer hasn't been set on this timeline already, set it lazily here
	if t.account == nil {
		timelineOwnerAccount := &gtsmodel.Account{}
		if err := t.db.GetByID(ctx, t.accountID, timelineOwnerAccount); err != nil {
			return err
		}
		t.account = timelineOwnerAccount
	}

	// serialize the status (or, at least, convert it to a form that's ready to be serialized)
	apiModelStatus, err := t.tc.StatusToAPIStatus(ctx, gtsStatus, t.account)
	if err != nil {
		return err
	}

	// shove it in prepared posts as a prepared posts entry
	preparedPostsEntry := &preparedPostsEntry{
		statusID:         gtsStatus.ID,
		boostOfID:        gtsStatus.BoostOfID,
		accountID:        gtsStatus.AccountID,
		boostOfAccountID: gtsStatus.BoostOfAccountID,
		prepared:         apiModelStatus,
	}

	return t.preparedPosts.insertPrepared(preparedPostsEntry)
}

func (t *timeline) OldestPreparedPostID(ctx context.Context) (string, error) {
	var id string
	if t.preparedPosts == nil || t.preparedPosts.data == nil {
		// return an empty string if prepared posts hasn't been initialized yet
		return id, nil
	}

	e := t.preparedPosts.data.Back()
	if e == nil {
		// return an empty string if there's no back entry (ie., the index list hasn't been initialized yet)
		return id, nil
	}

	entry, ok := e.Value.(*preparedPostsEntry)
	if !ok {
		return id, errors.New("OldestPreparedPostID: could not parse e as a preparedPostsEntry")
	}
	return entry.statusID, nil
}
