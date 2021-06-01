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
	"sync"
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

type timeline struct {
	index     *list.List
	readyToGo []*apimodel.Status
	sharedCache *list.List
	accountID string
	db        db.DB
	*sync.Mutex
}

func newTimeline(accountID string, db db.DB) *timeline {
	return &timeline{
		index:     list.New(),
		readyToGo: []*apimodel.Status{},
		accountID: accountID,
	}
}

func (t *timeline) prepareXFromID(limit int, statusID string) error {
	t.Lock()
	defer t.Unlock()

	return nil
}

func (t *timeline) getX(limit int) (*apimodel.Status, error) {
	t.Lock()
	defer t.Unlock()
	return nil, nil
}

func (t *timeline) insert(status *apimodel.Status) error {
	t.Lock()
	defer t.Unlock()

	newPost := &post{}

	if t.index == nil {
		t.index.PushFront(newPost)
	}

	for e := t.index.Front(); e != nil; e = e.Next() {
		p, ok := e.Value.(*post)
		if !ok {
			return errors.New("could not convert interface to post")
		}
		if p.createdAt.Before(newPost.createdAt) {

		}
	}
	return nil
}

type post struct {
	createdAt time.Time
	statusID  string
	serialized *apimodel.Status
}
