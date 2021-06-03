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
	"errors"
	"fmt"
	"sync"
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

const (
	preparedPostsMaxLength = desiredPostIndexLength
)

type Timeline interface {
	// GetXFromTop returns x amount of posts from the top of the timeline, from newest to oldest.
	GetXFromTop(amount int) ([]*apimodel.Status, error)
	// GetXFromID returns x amount of posts from the given id onwards, from newest to oldest.
	// This will include the status with the given ID.
	GetXFromID(amount int, fromID string) ([]*apimodel.Status, error)

	// IndexOne puts a status into the timeline at the appropriate place according to its 'createdAt' property.
	IndexOne(statusCreatedAt time.Time, statusID string) error
	// Remove removes a status from the timeline.
	Remove(statusID string) error
	// OldestIndexedPostID returns the id of the rearmost (ie., the oldest) indexed post, or an error if something goes wrong.
	// If nothing goes wrong but there's no oldest post, an empty string will be returned so make sure to check for this.
	OldestIndexedPostID() (string, error)

	// PrepareXFromTop instructs the timeline to prepare x amount of posts from the top of the timeline.
	PrepareXFromTop(amount int) error
	// PrepareXFromIndex instrucst the timeline to prepare the next amount of entries for serialization, from index onwards.
	PrepareXFromIndex(amount int, index int) error

	// ActualPostIndexLength returns the actual length of the post index at this point in time.
	PostIndexLength() int

	// Reset instructs the timeline to reset to its base state -- cache only the minimum amount of posts.
	Reset() error
}

type timeline struct {
	postIndex     *postIndex
	preparedPosts *preparedPosts
	accountID     string
	account       *gtsmodel.Account
	db            db.DB
	tc            typeutils.TypeConverter
	sync.Mutex
}

func NewTimeline(accountID string, db db.DB, typeConverter typeutils.TypeConverter) Timeline {
	return &timeline{
		postIndex:     &postIndex{},
		preparedPosts: &preparedPosts{},
		accountID:     accountID,
		db:            db,
		tc:            typeConverter,
	}
}

func (t *timeline) PrepareXFromIndex(amount int, index int) error {
	t.Lock()
	defer t.Unlock()

	var indexed int
	var prepared int
	var preparing bool
	for e := t.postIndex.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*postIndexEntry)
		if !ok {
			return errors.New("PrepareXFromTop: could not parse e as a postIndexEntry")
		}

		if !preparing {
			// we haven't hit the index we need to prepare from yet
			if indexed == index {
				preparing = true
			}
			indexed = indexed + 1
			continue
		} else {
			if err := t.prepare(entry.statusID); err != nil {
				return fmt.Errorf("PrepareXFromTop: error preparing status with id %s: %s", entry.statusID, err)
			}
			prepared = prepared + 1
			if prepared >= amount {
				// we're done
				break
			}
		}
	}

	return nil
}

func (t *timeline) PrepareXFromTop(amount int) error {
	t.Lock()
	defer t.Unlock()

	t.preparedPosts.data.Init()

	var prepared int
	for e := t.postIndex.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*postIndexEntry)
		if !ok {
			return errors.New("PrepareXFromTop: could not parse e as a postIndexEntry")
		}

		if err := t.prepare(entry.statusID); err != nil {
			return fmt.Errorf("PrepareXFromTop: error preparing status with id %s: %s", entry.statusID, err)
		}

		prepared = prepared + 1
		if prepared >= amount {
			// we're done
			break
		}
	}

	return nil
}

func (t *timeline) GetXFromTop(amount int) ([]*apimodel.Status, error) {
	// make a slice of statuses with the length we need to return
	statuses := make([]*apimodel.Status, 0, amount)

	// if there are no prepared posts, just return the empty slice
	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
	}

	// make sure we have enough posts prepared to return
	if t.preparedPosts.data.Len() < amount {
		if err := t.PrepareXFromTop(amount); err != nil {
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

func (t *timeline) GetXFromID(amount int, fromID string) ([]*apimodel.Status, error) {
	// make a slice of statuses with the length we need to return
	statuses := make([]*apimodel.Status, 0, amount)

	// if there are no prepared posts, just return the empty slice
	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
	}

	// find the position of id
	var position int
	for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXBehindID: could not parse e as a preparedPostsEntry")
		}
		if entry.statusID == fromID {
			break
		}
		position = position + 1
	}

	// make sure we have enough posts prepared behind it to return what we're being asked for
	if t.preparedPosts.data.Len() < amount+position {
		if err := t.PrepareXFromIndex(amount, position); err != nil {
			return nil, err
		}
	}

	// iterate through the modified list until we hit the fromID again
	var serving bool
	var served int
	for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXBehindID: could not parse e as a preparedPostsEntry")
		}

		if !serving {
			// start serving if we've hit the id we're looking for
			if entry.statusID == fromID {
				serving = true
			}
		}

		if serving {
			// serve up to the amount requested
			statuses = append(statuses, entry.prepared)
			served = served + 1
			if served >= amount {
				break
			}
		}
	}

	return statuses, nil
}

func (t *timeline) IndexOne(statusCreatedAt time.Time, statusID string) error {
	t.Lock()
	defer t.Unlock()

	postIndexEntry := &postIndexEntry{
		createdAt: statusCreatedAt,
		statusID:  statusID,
	}

	return t.postIndex.insertIndexed(postIndexEntry)
}

func (t *timeline) Remove(statusID string) error {
	t.Lock()
	defer t.Unlock()

	// remove the entry from the post index
	for e := t.postIndex.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*postIndexEntry)
		if !ok {
			return errors.New("Remove: could not parse e as a postIndexEntry")
		}
		if entry.statusID == statusID {
			t.postIndex.data.Remove(e)
			break // bail once we found and removed it
		}
	}

	// remove the entry from prepared posts
	for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return errors.New("Remove: could not parse e as a preparedPostsEntry")
		}
		if entry.statusID == statusID {
			t.preparedPosts.data.Remove(e)
			break // bail once we found and removed it
		}
	}

	return nil
}

func (t *timeline) Reset() error {
	return nil
}

func (t *timeline) PostIndexLength() int {
	if t.postIndex == nil || t.postIndex.data == nil {
		return 0
	}

	return t.postIndex.data.Len()
}

func (t *timeline) OldestIndexedPostID() (string, error) {
	var id string
	if t.postIndex == nil || t.postIndex.data == nil {
		// return an empty string if postindex hasn't been initialized yet
		return id, nil
	}

	e := t.postIndex.data.Back()

	if e == nil {
		// return an empty string if there's no back entry (ie., the index list hasn't been initialized yet)
		return id, nil
	}

	entry, ok := e.Value.(*postIndexEntry)
	if !ok {
		return id, errors.New("OldestIndexedPostID: could not parse e as a postIndexEntry")
	}
	return entry.statusID, nil
}

func (t *timeline) prepare(statusID string) error {

	// start by getting the status out of the database according to its indexed ID
	gtsStatus := &gtsmodel.Status{}
	if err := t.db.GetByID(statusID, gtsStatus); err != nil {
		return err
	}

	// if the account pointer hasn't been set on this timeline already, set it lazily here
	if t.account == nil {
		timelineOwnerAccount := &gtsmodel.Account{}
		if err := t.db.GetByID(t.accountID, timelineOwnerAccount); err != nil {
			return err
		}
		t.account = timelineOwnerAccount
	}

	// to convert the status we need relevant accounts from it, so pull them out here
	relevantAccounts, err := t.db.PullRelevantAccountsFromStatus(gtsStatus)
	if err != nil {
		return err
	}

	// check if this is a boost...
	var reblogOfStatus *gtsmodel.Status
	if gtsStatus.BoostOfID != "" {
		s := &gtsmodel.Status{}
		if err := t.db.GetByID(gtsStatus.BoostOfID, s); err != nil {
			return err
		}
		reblogOfStatus = s
	}

	// serialize the status (or, at least, convert it to a form that's ready to be serialized)
	apiModelStatus, err := t.tc.StatusToMasto(gtsStatus, relevantAccounts.StatusAuthor, t.account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, reblogOfStatus)
	if err != nil {
		return err
	}

	// shove it in prepared posts as a prepared posts entry
	preparedPostsEntry := &preparedPostsEntry{
		createdAt: gtsStatus.CreatedAt,
		statusID:  statusID,
		prepared:  apiModelStatus,
	}

	return t.preparedPosts.insertPrepared(preparedPostsEntry)
}
