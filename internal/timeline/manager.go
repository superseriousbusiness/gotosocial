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
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

const (
	desiredPostIndexLength = 400
)

// Manager abstracts functions for creating timelines for multiple accounts, and adding, removing, and fetching entries from those timelines.
//
// By the time a status hits the manager interface, it should already have been filtered and it should be established that the status indeed
// belongs in the home timeline of the given account ID.
//
// The manager makes a distinction between *indexed* posts and *prepared* posts.
//
// Indexed posts consist of just that post's ID (in the database) and the time it was created. An indexed post takes up very little memory, so
// it's not a huge priority to keep trimming the indexed posts list.
//
// Prepared posts consist of the post's database ID, the time it was created, AND the apimodel representation of that post, for quick serialization.
// Prepared posts of course take up more memory than indexed posts, so they should be regularly pruned if they're not being actively served.
type Manager interface {
	// Ingest takes one status and indexes it into the timeline for the given account ID.
	//
	// It should already be established before calling this function that the status/post actually belongs in the timeline!
	//
	// The returned bool indicates whether the status was actually put in the timeline. This could be false in cases where
	// the status is a boost, but a boost of the original post or the post itself already exists recently in the timeline.
	Ingest(status *gtsmodel.Status, timelineAccountID string) (bool, error)
	// IngestAndPrepare takes one status and indexes it into the timeline for the given account ID, and then immediately prepares it for serving.
	// This is useful in cases where we know the status will need to be shown at the top of a user's timeline immediately (eg., a new status is created).
	//
	// It should already be established before calling this function that the status/post actually belongs in the timeline!
	//
	// The returned bool indicates whether the status was actually put in the timeline. This could be false in cases where
	// the status is a boost, but a boost of the original post or the post itself already exists recently in the timeline.
	IngestAndPrepare(status *gtsmodel.Status, timelineAccountID string) (bool, error)
	// HomeTimeline returns limit n amount of entries from the home timeline of the given account ID, in descending chronological order.
	// If maxID is provided, it will return entries from that maxID onwards, inclusive.
	HomeTimeline(accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*apimodel.Status, error)
	// GetIndexedLength returns the amount of posts/statuses that have been *indexed* for the given account ID.
	GetIndexedLength(timelineAccountID string) int
	// GetDesiredIndexLength returns the amount of posts that we, ideally, index for each user.
	GetDesiredIndexLength() int
	// GetOldestIndexedID returns the status ID for the oldest post that we have indexed for the given account.
	GetOldestIndexedID(timelineAccountID string) (string, error)
	// PrepareXFromTop prepares limit n amount of posts, based on their indexed representations, from the top of the index.
	PrepareXFromTop(timelineAccountID string, limit int) error
	// Remove removes one status from the timeline of the given timelineAccountID
	Remove(timelineAccountID string, statusID string) (int, error)
	// WipeStatusFromAllTimelines removes one status from the index and prepared posts of all timelines
	WipeStatusFromAllTimelines(statusID string) error
	// WipeStatusesFromAccountID removes all statuses by the given accountID from the timelineAccountID's timelines.
	WipeStatusesFromAccountID(timelineAccountID string, accountID string) error
}

// NewManager returns a new timeline manager with the given database, typeconverter, config, and log.
func NewManager(db db.DB, tc typeutils.TypeConverter, config *config.Config, log *logrus.Logger) Manager {
	return &manager{
		accountTimelines: sync.Map{},
		db:               db,
		tc:               tc,
		config:           config,
		log:              log,
	}
}

type manager struct {
	accountTimelines sync.Map
	db               db.DB
	tc               typeutils.TypeConverter
	config           *config.Config
	log              *logrus.Logger
}

func (m *manager) Ingest(status *gtsmodel.Status, timelineAccountID string) (bool, error) {
	l := m.log.WithFields(logrus.Fields{
		"func":              "Ingest",
		"timelineAccountID": timelineAccountID,
		"statusID":          status.ID,
	})

	t, err := m.getOrCreateTimeline(timelineAccountID)
	if err != nil {
		return false, err
	}

	l.Trace("ingesting status")
	return t.IndexOne(status.CreatedAt, status.ID, status.BoostOfID, status.AccountID, status.BoostOfAccountID)
}

func (m *manager) IngestAndPrepare(status *gtsmodel.Status, timelineAccountID string) (bool, error) {
	l := m.log.WithFields(logrus.Fields{
		"func":              "IngestAndPrepare",
		"timelineAccountID": timelineAccountID,
		"statusID":          status.ID,
	})

	t, err := m.getOrCreateTimeline(timelineAccountID)
	if err != nil {
		return false, err
	}

	l.Trace("ingesting status")
	return t.IndexAndPrepareOne(status.CreatedAt, status.ID, status.BoostOfID, status.AccountID, status.BoostOfAccountID)
}

func (m *manager) Remove(timelineAccountID string, statusID string) (int, error) {
	l := m.log.WithFields(logrus.Fields{
		"func":              "Remove",
		"timelineAccountID": timelineAccountID,
		"statusID":          statusID,
	})

	t, err := m.getOrCreateTimeline(timelineAccountID)
	if err != nil {
		return 0, err
	}

	l.Trace("removing status")
	return t.Remove(statusID)
}

func (m *manager) HomeTimeline(timelineAccountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*apimodel.Status, error) {
	l := m.log.WithFields(logrus.Fields{
		"func":              "HomeTimelineGet",
		"timelineAccountID": timelineAccountID,
	})

	t, err := m.getOrCreateTimeline(timelineAccountID)
	if err != nil {
		return nil, err
	}

	statuses, err := t.Get(limit, maxID, sinceID, minID, true)
	if err != nil {
		l.Errorf("error getting statuses: %s", err)
	}
	return statuses, nil
}

func (m *manager) GetIndexedLength(timelineAccountID string) int {
	t, err := m.getOrCreateTimeline(timelineAccountID)
	if err != nil {
		return 0
	}

	return t.PostIndexLength()
}

func (m *manager) GetDesiredIndexLength() int {
	return desiredPostIndexLength
}

func (m *manager) GetOldestIndexedID(timelineAccountID string) (string, error) {
	t, err := m.getOrCreateTimeline(timelineAccountID)
	if err != nil {
		return "", err
	}

	return t.OldestIndexedPostID()
}

func (m *manager) PrepareXFromTop(timelineAccountID string, limit int) error {
	t, err := m.getOrCreateTimeline(timelineAccountID)
	if err != nil {
		return err
	}

	return t.PrepareFromTop(limit)
}

func (m *manager) WipeStatusFromAllTimelines(statusID string) error {
	errors := []string{}
	m.accountTimelines.Range(func(k interface{}, i interface{}) bool {
		t, ok := i.(Timeline)
		if !ok {
			panic("couldn't parse entry as Timeline, this should never happen so panic")
		}

		if _, err := t.Remove(statusID); err != nil {
			errors = append(errors, err.Error())
		}

		return true
	})

	var err error
	if len(errors) > 0 {
		err = fmt.Errorf("one or more errors removing status %s from all timelines: %s", statusID, strings.Join(errors, ";"))
	}

	return err
}

func (m *manager) WipeStatusesFromAccountID(timelineAccountID string, accountID string) error {
	t, err := m.getOrCreateTimeline(timelineAccountID)
	if err != nil {
		return err
	}

	_, err = t.RemoveAllBy(accountID)
	return err
}

func (m *manager) getOrCreateTimeline(timelineAccountID string) (Timeline, error) {
	var t Timeline
	i, ok := m.accountTimelines.Load(timelineAccountID)
	if !ok {
		var err error
		t, err = NewTimeline(timelineAccountID, m.db, m.tc, m.log)
		if err != nil {
			return nil, err
		}
		m.accountTimelines.Store(timelineAccountID, t)
	} else {
		t, ok = i.(Timeline)
		if !ok {
			panic("couldn't parse entry as Timeline, this should never happen so panic")
		}
	}

	return t, nil
}
