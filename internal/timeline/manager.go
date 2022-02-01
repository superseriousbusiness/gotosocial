/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	desiredPostIndexLength = 400
)

// Manager abstracts functions for creating timelines for multiple accounts, and adding, removing, and fetching entries from those timelines.
//
// By the time a timelineable hits the manager interface, it should already have been filtered and it should be established that the item indeed
// belongs in the timeline of the given account ID.
//
// The manager makes a distinction between *indexed* items and *prepared* items.
//
// Indexed items consist of just that item's ID (in the database) and the time it was created. An indexed item takes up very little memory, so
// it's not a huge priority to keep trimming the indexed items list.
//
// Prepared items consist of the item's database ID, the time it was created, AND the apimodel representation of that item, for quick serialization.
// Prepared items of course take up more memory than indexed items, so they should be regularly pruned if they're not being actively served.
type Manager interface {
	// Ingest takes one item and indexes it into the timeline for the given account ID.
	//
	// It should already be established before calling this function that the item actually belongs in the timeline!
	//
	// The returned bool indicates whether the item was actually put in the timeline. This could be false in cases where
	// the item is a boosted status, but a boost of the original status or the status itself already exists recently in the timeline.
	Ingest(ctx context.Context, item Timelineable, timelineAccountID string) (bool, error)
	// IngestAndPrepare takes one timelineable and indexes it into the timeline for the given account ID, and then immediately prepares it for serving.
	// This is useful in cases where we know the item will need to be shown at the top of a user's timeline immediately (eg., a new status is created).
	//
	// It should already be established before calling this function that the item actually belongs in the timeline!
	//
	// The returned bool indicates whether the item was actually put in the timeline. This could be false in cases where
	// a status is a boost, but a boost of the original status or the status itself already exists recently in the timeline.
	IngestAndPrepare(ctx context.Context, item Timelineable, timelineAccountID string) (bool, error)
	// GetTimeline returns limit n amount of prepared entries from the timeline of the given account ID, in descending chronological order.
	// If maxID is provided, it will return prepared entries from that maxID onwards, inclusive.
	GetTimeline(ctx context.Context, accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]Preparable, error)
	// GetIndexedLength returns the amount of items that have been *indexed* for the given account ID.
	GetIndexedLength(ctx context.Context, timelineAccountID string) int
	// GetDesiredIndexLength returns the amount of items that we, ideally, index for each user.
	GetDesiredIndexLength(ctx context.Context) int
	// GetOldestIndexedID returns the id ID for the oldest item that we have indexed for the given account.
	GetOldestIndexedID(ctx context.Context, timelineAccountID string) (string, error)
	// PrepareXFromTop prepares limit n amount of items, based on their indexed representations, from the top of the index.
	PrepareXFromTop(ctx context.Context, timelineAccountID string, limit int) error
	// Remove removes one item from the timeline of the given timelineAccountID
	Remove(ctx context.Context, timelineAccountID string, itemID string) (int, error)
	// WipeItemFromAllTimelines removes one item from the index and prepared items of all timelines
	WipeItemFromAllTimelines(ctx context.Context, itemID string) error
	// WipeStatusesFromAccountID removes all items by the given accountID from the timelineAccountID's timelines.
	WipeItemsFromAccountID(ctx context.Context, timelineAccountID string, accountID string) error
}

// NewManager returns a new timeline manager.
func NewManager(grabFunction GrabFunction, filterFunction FilterFunction, prepareFunction PrepareFunction, skipInsertFunction SkipInsertFunction) Manager {
	return &manager{
		accountTimelines:   sync.Map{},
		grabFunction:       grabFunction,
		filterFunction:     filterFunction,
		prepareFunction:    prepareFunction,
		skipInsertFunction: skipInsertFunction,
	}
}

type manager struct {
	accountTimelines   sync.Map
	grabFunction       GrabFunction
	filterFunction     FilterFunction
	prepareFunction    PrepareFunction
	skipInsertFunction SkipInsertFunction
}

func (m *manager) Ingest(ctx context.Context, item Timelineable, timelineAccountID string) (bool, error) {
	l := logrus.WithFields(logrus.Fields{
		"func":              "Ingest",
		"timelineAccountID": timelineAccountID,
		"itemID":            item.GetID(),
	})

	t, err := m.getOrCreateTimeline(ctx, timelineAccountID)
	if err != nil {
		return false, err
	}

	l.Trace("ingesting item")
	return t.IndexOne(ctx, item.GetID(), item.GetBoostOfID(), item.GetAccountID(), item.GetBoostOfAccountID())
}

func (m *manager) IngestAndPrepare(ctx context.Context, item Timelineable, timelineAccountID string) (bool, error) {
	l := logrus.WithFields(logrus.Fields{
		"func":              "IngestAndPrepare",
		"timelineAccountID": timelineAccountID,
		"itemID":            item.GetID(),
	})

	t, err := m.getOrCreateTimeline(ctx, timelineAccountID)
	if err != nil {
		return false, err
	}

	l.Trace("ingesting item")
	return t.IndexAndPrepareOne(ctx, item.GetID(), item.GetBoostOfID(), item.GetAccountID(), item.GetBoostOfAccountID())
}

func (m *manager) Remove(ctx context.Context, timelineAccountID string, itemID string) (int, error) {
	l := logrus.WithFields(logrus.Fields{
		"func":              "Remove",
		"timelineAccountID": timelineAccountID,
		"itemID":            itemID,
	})

	t, err := m.getOrCreateTimeline(ctx, timelineAccountID)
	if err != nil {
		return 0, err
	}

	l.Trace("removing item")
	return t.Remove(ctx, itemID)
}

func (m *manager) GetTimeline(ctx context.Context, timelineAccountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]Preparable, error) {
	l := logrus.WithFields(logrus.Fields{
		"func":              "GetTimeline",
		"timelineAccountID": timelineAccountID,
	})

	t, err := m.getOrCreateTimeline(ctx, timelineAccountID)
	if err != nil {
		return nil, err
	}

	items, err := t.Get(ctx, limit, maxID, sinceID, minID, true)
	if err != nil {
		l.Errorf("error getting statuses: %s", err)
	}
	return items, nil
}

func (m *manager) GetIndexedLength(ctx context.Context, timelineAccountID string) int {
	t, err := m.getOrCreateTimeline(ctx, timelineAccountID)
	if err != nil {
		return 0
	}

	return t.ItemIndexLength(ctx)
}

func (m *manager) GetDesiredIndexLength(ctx context.Context) int {
	return desiredPostIndexLength
}

func (m *manager) GetOldestIndexedID(ctx context.Context, timelineAccountID string) (string, error) {
	t, err := m.getOrCreateTimeline(ctx, timelineAccountID)
	if err != nil {
		return "", err
	}

	return t.OldestIndexedItemID(ctx)
}

func (m *manager) PrepareXFromTop(ctx context.Context, timelineAccountID string, limit int) error {
	t, err := m.getOrCreateTimeline(ctx, timelineAccountID)
	if err != nil {
		return err
	}

	return t.PrepareFromTop(ctx, limit)
}

func (m *manager) WipeItemFromAllTimelines(ctx context.Context, statusID string) error {
	errors := []string{}
	m.accountTimelines.Range(func(k interface{}, i interface{}) bool {
		t, ok := i.(Timeline)
		if !ok {
			panic("couldn't parse entry as Timeline, this should never happen so panic")
		}

		if _, err := t.Remove(ctx, statusID); err != nil {
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

func (m *manager) WipeItemsFromAccountID(ctx context.Context, timelineAccountID string, accountID string) error {
	t, err := m.getOrCreateTimeline(ctx, timelineAccountID)
	if err != nil {
		return err
	}

	_, err = t.RemoveAllBy(ctx, accountID)
	return err
}

func (m *manager) getOrCreateTimeline(ctx context.Context, timelineAccountID string) (Timeline, error) {
	var t Timeline
	i, ok := m.accountTimelines.Load(timelineAccountID)
	if !ok {
		var err error
		t, err = NewTimeline(ctx, timelineAccountID, m.grabFunction, m.filterFunction, m.prepareFunction, m.skipInsertFunction)
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
