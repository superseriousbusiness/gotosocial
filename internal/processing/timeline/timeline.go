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
)

const (
	fromLatest             = "FROM_LATEST"
	postIndexMinLength     = 200
	postIndexMaxLength     = 400
	preparedPostsMaxLength = 400
	preparedPostsMinLength = 80
)

type timeline struct {
	// 
	postIndex     *list.List
	preparedPosts *list.List
	accountID     string
	db            db.DB
	*sync.Mutex
}

func newTimeline(accountID string, db db.DB, sharedCache *list.List) *timeline {
	return &timeline{
		postIndex:     list.New(),
		preparedPosts: list.New(),
		accountID:     accountID,
		db:            db,
	}
}

func (t *timeline) prepareNextXFromID(amount int, fromID string) error {
	t.Lock()
	defer t.Unlock()

	prepared := make([]*post, 0, amount)

	// find the mark in the index -- we want x statuses after this
	var fromMark *list.Element
	for e := t.postIndex.Front(); e != nil; e = e.Next() {
		p, ok := e.Value.(*post)
		if !ok {
			return errors.New("could not convert interface to post")
		}

		if p.statusID == fromID {
			fromMark = e
			break
		}
	}

	if fromMark == nil {
		// we can't find the given id in the index -_-
		return fmt.Errorf("prepareNextXFromID: fromID %s not found in index", fromID)
	}

	for e := fromMark.Next(); e != nil; e = e.Next() {

	}

	return nil
}

func (t *timeline) getXFromTop(amount int) ([]*apimodel.Status, error) {
	statuses := []*apimodel.Status{}
	if amount == 0 {
		return statuses, nil
	}

	if len(t.readyToGo) < amount {
		if err := t.prepareNextXFromID(amount, fromLatest); err != nil {
			return nil, err
		}
	}

	return t.readyToGo[:amount], nil
}

// getXFromID gets x amount of posts in chronological order from the given ID onwards, NOT including the given id.
// The posts will be taken from the preparedPosts pile, unless nothing is ready to go.
func (t *timeline) getXFromID(amount int, fromID string) ([]*apimodel.Status, error) {
	statuses := []*apimodel.Status{}
	if amount == 0 || fromID == "" {
		return statuses, nil
	}

	// get the position of the given id in the ready to go pile
	var indexOfID *int
	for i, s := range t.readyToGo {
		if s.ID == fromID {
			indexOfID = &i
		}
	}

	// the status isn't in the ready to go pile so prepare it
	if indexOfID == nil {
		if err := t.prepareNextXFromID(amount, fromID); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (t *timeline) insert(status *apimodel.Status) error {
	t.Lock()
	defer t.Unlock()

	createdAt, err := time.Parse(time.RFC3339, status.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert: could not parse time %s: %s", status.CreatedAt, err)
	}

	newPost := &post{
		createdAt:  createdAt,
		statusID:   status.ID,
		serialized: status,
	}

	if t.index == nil {
		t.index.PushFront(newPost)
	}

	for e := t.index.Front(); e != nil; e = e.Next() {
		p, ok := e.Value.(*post)
		if !ok {
			return errors.New("could not convert interface to post")
		}

		if newPost.createdAt.After(p.createdAt) {
			// this is a newer post so insert it just before the post it's newer than
			t.index.InsertBefore(newPost, e)
			return nil
		}
	}

	// if we haven't returned yet it's the oldest post we've seen so shove it at the back
	t.index.PushBack(newPost)
	return nil
}

type preparedPostsEntry struct {
	createdAt  time.Time
	statusID   string
	serialized *apimodel.Status
}

type postIndexEntry struct {
	createdAt  time.Time
	statusID   string
}
