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
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

const boostReinsertionDepth = 50

// Timeline represents a timeline for one account, and contains indexed and prepared posts.
type Timeline interface {
	/*
		RETRIEVAL FUNCTIONS
	*/

	Get(amount int, maxID string, sinceID string, minID string) ([]*apimodel.Status, error)
	// GetXFromTop returns x amount of posts from the top of the timeline, from newest to oldest.
	GetXFromTop(amount int) ([]*apimodel.Status, error)
	// GetXBehindID returns x amount of posts from the given id onwards, from newest to oldest.
	// This will NOT include the status with the given ID.
	//
	// This corresponds to an api call to /timelines/home?max_id=WHATEVER
	GetXBehindID(amount int, fromID string) ([]*apimodel.Status, error)
	// GetXBeforeID returns x amount of posts up to the given id, from newest to oldest.
	// This will NOT include the status with the given ID.
	//
	// This corresponds to an api call to /timelines/home?since_id=WHATEVER
	GetXBeforeID(amount int, sinceID string, startFromTop bool) ([]*apimodel.Status, error)
	// GetXBetweenID returns x amount of posts from the given maxID, up to the given id, from newest to oldest.
	// This will NOT include the status with the given IDs.
	//
	// This corresponds to an api call to /timelines/home?since_id=WHATEVER&max_id=WHATEVER_ELSE
	GetXBetweenID(amount int, maxID string, sinceID string) ([]*apimodel.Status, error)

	/*
		INDEXING FUNCTIONS
	*/

	// IndexOne puts a status into the timeline at the appropriate place according to its 'createdAt' property.
	//
	// The returned bool indicates whether or not the status was actually inserted into the timeline. This will be false
	// if the status is a boost and the original post or another boost of it already exists < boostReinsertionDepth back in the timeline.
	IndexOne(statusCreatedAt time.Time, statusID string, boostOfID string, accountID string, boostOfAccountID string) (bool, error)

	// OldestIndexedPostID returns the id of the rearmost (ie., the oldest) indexed post, or an error if something goes wrong.
	// If nothing goes wrong but there's no oldest post, an empty string will be returned so make sure to check for this.
	OldestIndexedPostID() (string, error)

	/*
		PREPARATION FUNCTIONS
	*/

	// PrepareXFromTop instructs the timeline to prepare x amount of posts from the top of the timeline.
	PrepareFromTop(amount int) error
	// PrepareBehind instructs the timeline to prepare the next amount of entries for serialization, from position onwards.
	// If include is true, then the given status ID will also be prepared, otherwise only entries behind it will be prepared.
	PrepareBehind(statusID string, amount int) error
	// IndexOne puts a status into the timeline at the appropriate place according to its 'createdAt' property,
	// and then immediately prepares it.
	//
	// The returned bool indicates whether or not the status was actually inserted into the timeline. This will be false
	// if the status is a boost and the original post or another boost of it already exists < boostReinsertionDepth back in the timeline.
	IndexAndPrepareOne(statusCreatedAt time.Time, statusID string, boostOfID string, accountID string, boostOfAccountID string) (bool, error)
	// OldestPreparedPostID returns the id of the rearmost (ie., the oldest) prepared post, or an error if something goes wrong.
	// If nothing goes wrong but there's no oldest post, an empty string will be returned so make sure to check for this.
	OldestPreparedPostID() (string, error)

	/*
		INFO FUNCTIONS
	*/

	// ActualPostIndexLength returns the actual length of the post index at this point in time.
	PostIndexLength() int

	/*
		UTILITY FUNCTIONS
	*/

	// Reset instructs the timeline to reset to its base state -- cache only the minimum amount of posts.
	Reset() error
	// Remove removes a status from both the index and prepared posts.
	//
	// If a status has multiple entries in a timeline, they will all be removed.
	//
	// The returned int indicates the amount of entries that were removed.
	Remove(statusID string) (int, error)
	// RemoveAllBy removes all statuses by the given accountID, from both the index and prepared posts.
	//
	// The returned int indicates the amount of entries that were removed.
	RemoveAllBy(accountID string) (int, error)
}

// timeline fulfils the Timeline interface
type timeline struct {
	postIndex     *postIndex
	preparedPosts *preparedPosts
	accountID     string
	account       *gtsmodel.Account
	db            db.DB
	filter        visibility.Filter
	tc            typeutils.TypeConverter
	log           *logrus.Logger
	sync.Mutex
}

// NewTimeline returns a new Timeline for the given account ID
func NewTimeline(accountID string, db db.DB, typeConverter typeutils.TypeConverter, log *logrus.Logger) (Timeline, error) {
	timelineOwnerAccount := &gtsmodel.Account{}
	if err := db.GetByID(accountID, timelineOwnerAccount); err != nil {
		return nil, err
	}

	return &timeline{
		postIndex:     &postIndex{},
		preparedPosts: &preparedPosts{},
		accountID:     accountID,
		account:       timelineOwnerAccount,
		db:            db,
		filter:        visibility.NewFilter(db, log),
		tc:            typeConverter,
		log:           log,
	}, nil
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
