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
	fromLatest             = "FROM_LATEST"
	preparedPostsMaxLength = desiredPostIndexLength
)

type Timeline interface {
	// GetXFromTop returns x amount of posts from the top of the timeline, from newest to oldest.
	GetXFromTop(amount int) ([]*apimodel.Status, error)
	// GetXFromTop returns x amount of posts from the given id onwards, from newest to oldest.
	GetXBehindID(amount int, fromID string) ([]*apimodel.Status, error)

	// IndexOne puts a status into the timeline at the appropriate place according to its 'createdAt' property.
	IndexOne(statusCreatedAt time.Time, statusID string) error
	// IndexMany instructs the timeline to index all the given posts.
	IndexMany([]*apimodel.Status) error
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
				fmt.Printf("\n\n\nprepared %d entries\n\n\n", prepared)
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

func (t *timeline) GetXBehindID(amount int, fromID string) ([]*apimodel.Status, error) {
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
			fmt.Printf("\n\n\nfromid %s is at position %d\n\n\n", fromID, position)
			break
		}
		position = position + 1
	}

	// make sure we have enough posts prepared to return
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
			// we're not serving yet but we might on the next time round if we hit our from id
			if entry.statusID == fromID {
				fmt.Printf("\n\n\nwe've hit fromid %s at position %d, will now serve\n\n\n", fromID, position)
				serving = true
				continue
			}
		} else {
			// we're serving now!
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
	return t.postIndex.index(postIndexEntry)
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
			break // bail once we found it
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
			break // bail once we found it
		}
	}

	return nil
}

func (t *timeline) IndexMany(statuses []*apimodel.Status) error {
	t.Lock()
	defer t.Unlock()

	// add statuses to the index
	for _, s := range statuses {
		createdAt, err := time.Parse(s.CreatedAt, time.RFC3339)
		if err != nil {
			return fmt.Errorf("IndexMany: could not parse time %s on status id %s: %s", s.CreatedAt, s.ID, err)
		}
		postIndexEntry := &postIndexEntry{
			createdAt: createdAt,
			statusID:  s.ID,
		}
		if err := t.postIndex.index(postIndexEntry); err != nil {
			return err
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
		return id, nil
	}
	e := t.postIndex.data.Back()

	if e == nil {
		return id, nil
	}

	entry, ok := e.Value.(*postIndexEntry)
	if !ok {
		return id, errors.New("OldestIndexedPostID: could not parse e as a postIndexEntry")
	}

	return entry.statusID, nil
}

func (t *timeline) prepare(statusID string) error {
	gtsStatus := &gtsmodel.Status{}
	if err := t.db.GetByID(statusID, gtsStatus); err != nil {
		return err
	}

	if t.account == nil {
		timelineOwnerAccount := &gtsmodel.Account{}
		if err := t.db.GetByID(t.accountID, timelineOwnerAccount); err != nil {
			return err
		}
		t.account = timelineOwnerAccount
	}

	relevantAccounts, err := t.db.PullRelevantAccountsFromStatus(gtsStatus)
	if err != nil {
		return err
	}

	var reblogOfStatus *gtsmodel.Status
	if gtsStatus.BoostOfID != "" {
		s := &gtsmodel.Status{}
		if err := t.db.GetByID(gtsStatus.BoostOfID, s); err != nil {
			return err
		}
		reblogOfStatus = s
	}

	apiModelStatus, err := t.tc.StatusToMasto(gtsStatus, relevantAccounts.StatusAuthor, t.account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, reblogOfStatus)
	if err != nil {
		return err
	}

	preparedPostsEntry := &preparedPostsEntry{
		createdAt: gtsStatus.CreatedAt,
		statusID:  statusID,
		prepared:  apiModelStatus,
	}

	return t.preparedPosts.insertPrepared(preparedPostsEntry)
}
