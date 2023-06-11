// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package timeline

import (
	"context"
	"sync"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

const (
	pruneLengthIndexed  = 400
	pruneLengthPrepared = 50
)

// Manager abstracts functions for creating multiple timelines, and adding, removing, and fetching entries from those timelines.
//
// By the time a timelineable hits the manager interface, it should already have been filtered and it should be established that the item indeed
// belongs in the given timeline.
//
// The manager makes a distinction between *indexed* items and *prepared* items.
//
// Indexed items consist of just that item's ID (in the database) and the time it was created. An indexed item takes up very little memory, so
// it's not a huge priority to keep trimming the indexed items list.
//
// Prepared items consist of the item's database ID, the time it was created, AND the apimodel representation of that item, for quick serialization.
// Prepared items of course take up more memory than indexed items, so they should be regularly pruned if they're not being actively served.
type Manager interface {
	// IngestOne takes one timelineable and indexes it into the given timeline, and then immediately prepares it for serving.
	// This is useful in cases where we know the item will need to be shown at the top of a user's timeline immediately (eg., a new status is created).
	//
	// It should already be established before calling this function that the item actually belongs in the timeline!
	//
	// The returned bool indicates whether the item was actually put in the timeline. This could be false in cases where
	// a status is a boost, but a boost of the original status or the status itself already exists recently in the timeline.
	IngestOne(ctx context.Context, timelineID string, item Timelineable) (bool, error)

	// GetTimeline returns limit n amount of prepared entries from the given timeline, in descending chronological order.
	GetTimeline(ctx context.Context, timelineID string, maxID string, sinceID string, minID string, limit int, local bool) ([]Preparable, error)

	// GetIndexedLength returns the amount of items that have been indexed for the given account ID.
	GetIndexedLength(ctx context.Context, timelineID string) int

	// GetOldestIndexedID returns the id ID for the oldest item that we have indexed for the given timeline.
	// Will be an empty string if nothing is (yet) indexed.
	GetOldestIndexedID(ctx context.Context, timelineID string) string

	// Remove removes one item from the given timeline.
	Remove(ctx context.Context, timelineID string, itemID string) (int, error)

	// RemoveTimeline completely removes one timeline.
	RemoveTimeline(ctx context.Context, timelineID string) error

	// WipeItemFromAllTimelines removes one item from the index and prepared items of all timelines
	WipeItemFromAllTimelines(ctx context.Context, itemID string) error

	// WipeStatusesFromAccountID removes all items by the given accountID from the given timeline.
	WipeItemsFromAccountID(ctx context.Context, timelineID string, accountID string) error

	// UnprepareItem unprepares/uncaches the prepared version fo the given itemID from the given timelineID.
	// Use this for cache invalidation when the prepared representation of an item has changed.
	UnprepareItem(ctx context.Context, timelineID string, itemID string) error

	// UnprepareItemFromAllTimelines unprepares/uncaches the prepared version of the given itemID from all timelines.
	// Use this for cache invalidation when the prepared representation of an item has changed.
	UnprepareItemFromAllTimelines(ctx context.Context, itemID string) error

	// Prune manually triggers a prune operation for the given timelineID.
	Prune(ctx context.Context, timelineID string, desiredPreparedItemsLength int, desiredIndexedItemsLength int) (int, error)

	// Start starts hourly cleanup jobs for this timeline manager.
	Start() error

	// Stop stops the timeline manager (currently a stub, doesn't do anything).
	Stop() error
}

// NewManager returns a new timeline manager.
func NewManager(grabFunction GrabFunction, filterFunction FilterFunction, prepareFunction PrepareFunction, skipInsertFunction SkipInsertFunction) Manager {
	return &manager{
		timelines:          sync.Map{},
		grabFunction:       grabFunction,
		filterFunction:     filterFunction,
		prepareFunction:    prepareFunction,
		skipInsertFunction: skipInsertFunction,
	}
}

type manager struct {
	timelines          sync.Map
	grabFunction       GrabFunction
	filterFunction     FilterFunction
	prepareFunction    PrepareFunction
	skipInsertFunction SkipInsertFunction
}

func (m *manager) Start() error {
	// Start a background goroutine which iterates
	// through all stored timelines once per hour,
	// and cleans up old entries if that timeline
	// hasn't been accessed in the last hour.
	go func() {
		for now := range time.NewTicker(1 * time.Hour).C {
			// Define the range function inside here,
			// so that we can use the 'now' returned
			// by the ticker, instead of having to call
			// time.Now() multiple times.
			//
			// Unless it panics, this function always
			// returns 'true', to continue the Range
			// call through the sync.Map.
			f := func(_ any, v any) bool {
				timeline, ok := v.(Timeline)
				if !ok {
					log.Panic(nil, "couldn't parse timeline manager sync map value as Timeline, this should never happen so panic")
				}

				if now.Sub(timeline.LastGot()) < 1*time.Hour {
					// Timeline has been fetched in the
					// last hour, move on to the next one.
					return true
				}

				if amountPruned := timeline.Prune(pruneLengthPrepared, pruneLengthIndexed); amountPruned > 0 {
					log.WithField("accountID", timeline.TimelineID()).Infof("pruned %d indexed and prepared items from timeline", amountPruned)
				}

				return true
			}

			// Execute the function for each timeline.
			m.timelines.Range(f)
		}
	}()

	return nil
}

func (m *manager) Stop() error {
	return nil
}

func (m *manager) IngestOne(ctx context.Context, timelineID string, item Timelineable) (bool, error) {
	return m.getOrCreateTimeline(ctx, timelineID).IndexAndPrepareOne(
		ctx,
		item.GetID(),
		item.GetBoostOfID(),
		item.GetAccountID(),
		item.GetBoostOfAccountID(),
	)
}

func (m *manager) Remove(ctx context.Context, timelineID string, itemID string) (int, error) {
	return m.getOrCreateTimeline(ctx, timelineID).Remove(ctx, itemID)
}

func (m *manager) RemoveTimeline(ctx context.Context, timelineID string) error {
	m.timelines.Delete(timelineID)
	return nil
}

func (m *manager) GetTimeline(ctx context.Context, timelineID string, maxID string, sinceID string, minID string, limit int, local bool) ([]Preparable, error) {
	return m.getOrCreateTimeline(ctx, timelineID).Get(ctx, limit, maxID, sinceID, minID, true)
}

func (m *manager) GetIndexedLength(ctx context.Context, timelineID string) int {
	return m.getOrCreateTimeline(ctx, timelineID).Len()
}

func (m *manager) GetOldestIndexedID(ctx context.Context, timelineID string) string {
	return m.getOrCreateTimeline(ctx, timelineID).OldestIndexedItemID()
}

func (m *manager) WipeItemFromAllTimelines(ctx context.Context, itemID string) error {
	errors := gtserror.MultiError{}

	m.timelines.Range(func(_ any, v any) bool {
		if _, err := v.(Timeline).Remove(ctx, itemID); err != nil {
			errors.Append(err)
		}

		return true // always continue range
	})

	if len(errors) > 0 {
		return gtserror.Newf("error(s) wiping status %s: %w", itemID, errors.Combine())
	}

	return nil
}

func (m *manager) WipeItemsFromAccountID(ctx context.Context, timelineID string, accountID string) error {
	_, err := m.getOrCreateTimeline(ctx, timelineID).RemoveAllByOrBoosting(ctx, accountID)
	return err
}

func (m *manager) UnprepareItemFromAllTimelines(ctx context.Context, itemID string) error {
	errors := gtserror.MultiError{}

	// Work through all timelines held by this
	// manager, and call Unprepare for each.
	m.timelines.Range(func(_ any, v any) bool {
		// nolint:forcetypeassert
		if err := v.(Timeline).Unprepare(ctx, itemID); err != nil {
			errors.Append(err)
		}

		return true // always continue range
	})

	if len(errors) > 0 {
		return gtserror.Newf("error(s) unpreparing status %s: %w", itemID, errors.Combine())
	}

	return nil
}

func (m *manager) UnprepareItem(ctx context.Context, timelineID string, itemID string) error {
	return m.getOrCreateTimeline(ctx, timelineID).Unprepare(ctx, itemID)
}

func (m *manager) Prune(ctx context.Context, timelineID string, desiredPreparedItemsLength int, desiredIndexedItemsLength int) (int, error) {
	return m.getOrCreateTimeline(ctx, timelineID).Prune(desiredPreparedItemsLength, desiredIndexedItemsLength), nil
}

// getOrCreateTimeline returns a timeline with the given id,
// creating a new timeline with that id if necessary.
func (m *manager) getOrCreateTimeline(ctx context.Context, timelineID string) Timeline {
	i, ok := m.timelines.Load(timelineID)
	if ok {
		// Timeline already existed in sync.Map.
		return i.(Timeline) //nolint:forcetypeassert
	}

	// Timeline did not yet exist in sync.Map.
	// Create + store it.
	timeline := NewTimeline(ctx, timelineID, m.grabFunction, m.filterFunction, m.prepareFunction, m.skipInsertFunction)
	m.timelines.Store(timelineID, timeline)

	return timeline
}
