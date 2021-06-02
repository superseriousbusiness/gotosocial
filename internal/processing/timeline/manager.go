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

	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

const (
	preparedPostsMinLength = 80
	desiredPostIndexLength = 400
)

type Manager interface {
	Ingest(status *gtsmodel.Status, timelineAccountID string) error
	HomeTimeline(timelineAccountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*apimodel.Status, error)
	GetIndexedLength(timelineAccountID string) int
	GetDesiredIndexLength() int
	GetOldestIndexedID(timelineAccountID string) (string, error)
	PrepareXFromTop(timelineAccountID string, limit int) error
}

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

func (m *manager) Ingest(status *gtsmodel.Status, timelineAccountID string) error {
	l := m.log.WithFields(logrus.Fields{
		"func":              "Ingest",
		"timelineAccountID": timelineAccountID,
		"statusID":          status.ID,
	})

	var t Timeline
	i, ok := m.accountTimelines.Load(timelineAccountID)
	if !ok {
		t = NewTimeline(timelineAccountID, m.db, m.tc)
		m.accountTimelines.Store(timelineAccountID, t)
	} else {
		t = i.(Timeline)
	}

	l.Trace("ingesting status")

	return t.IndexOne(status.CreatedAt, status.ID)
}

func (m *manager) Remove(statusID string, timelineAccountID string) error {
	l := m.log.WithFields(logrus.Fields{
		"func":              "Remove",
		"timelineAccountID": timelineAccountID,
		"statusID":          statusID,
	})

	var t Timeline
	i, ok := m.accountTimelines.Load(timelineAccountID)
	if !ok {
		t = NewTimeline(timelineAccountID, m.db, m.tc)
		m.accountTimelines.Store(timelineAccountID, t)
	} else {
		t = i.(Timeline)
	}

	l.Trace("removing status")

	return t.Remove(statusID)
}

func (m *manager) HomeTimeline(timelineAccountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*apimodel.Status, error) {
	l := m.log.WithFields(logrus.Fields{
		"func":              "HomeTimelineGet",
		"timelineAccountID": timelineAccountID,
	})

	var t Timeline
	i, ok := m.accountTimelines.Load(timelineAccountID)
	if !ok {
		t = NewTimeline(timelineAccountID, m.db, m.tc)
		m.accountTimelines.Store(timelineAccountID, t)
	} else {
		t = i.(Timeline)
	}

	var err error
	var statuses []*apimodel.Status
	if maxID != "" {
		statuses, err = t.GetXBehindID(limit, maxID)
	} else {
		statuses, err = t.GetXFromTop(limit)
	}

	if err != nil {
		l.Errorf("error getting statuses: %s", err)
	}
	return statuses, nil
}

func (m *manager) GetIndexedLength(timelineAccountID string) int {
	var t Timeline
	i, ok := m.accountTimelines.Load(timelineAccountID)
	if !ok {
		t = NewTimeline(timelineAccountID, m.db, m.tc)
		m.accountTimelines.Store(timelineAccountID, t)
	} else {
		t = i.(Timeline)
	}

	return t.PostIndexLength()
}

func (m *manager) GetDesiredIndexLength() int {
	return desiredPostIndexLength
}

func (m *manager) GetOldestIndexedID(timelineAccountID string) (string, error) {
	var t Timeline
	i, ok := m.accountTimelines.Load(timelineAccountID)
	if !ok {
		t = NewTimeline(timelineAccountID, m.db, m.tc)
		m.accountTimelines.Store(timelineAccountID, t)
	} else {
		t = i.(Timeline)
	}

	return t.OldestIndexedPostID()
}

func (m *manager) PrepareXFromTop(timelineAccountID string, limit int) error {
	var t Timeline
	i, ok := m.accountTimelines.Load(timelineAccountID)
	if !ok {
		t = NewTimeline(timelineAccountID, m.db, m.tc)
		m.accountTimelines.Store(timelineAccountID, t)
	} else {
		t = i.(Timeline)
	}

	return t.PrepareXFromTop(limit)
}
