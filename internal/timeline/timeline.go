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
)

// GrabFunction is used by a Timeline to grab more items to index.
//
// It should be provided to NewTimeline when the caller is creating a timeline
// (of statuses, notifications, etc).
//
//   - timelineID: ID of the timeline.
//   - maxID: the maximum item ID desired.
//   - sinceID: the minimum item ID desired.
//   - minID: see sinceID
//   - limit: the maximum amount of items to be returned
//
// If an error is returned, the timeline will stop processing whatever request called GrabFunction,
// and return the error. If no error is returned, but stop = true, this indicates to the caller of GrabFunction
// that there are no more items to return, and processing should continue with the items already grabbed.
type GrabFunction func(ctx context.Context, timelineID string, maxID string, sinceID string, minID string, limit int) (items []Timelineable, stop bool, err error)

// FilterFunction is used by a Timeline to filter whether or not a grabbed item should be indexed.
type FilterFunction func(ctx context.Context, timelineID string, item Timelineable) (shouldIndex bool, err error)

// PrepareFunction converts a Timelineable into a Preparable.
//
// For example, this might result in the converstion of a *gtsmodel.Status with the given itemID into a serializable *apimodel.Status.
type PrepareFunction func(ctx context.Context, timelineID string, itemID string) (Preparable, error)

// SkipInsertFunction indicates whether a new item about to be inserted in the prepared list should be skipped,
// based on the item itself, the next item in the timeline, and the depth at which nextItem has been found in the list.
//
// This will be called for every item found while iterating through a timeline, so callers should be very careful
// not to do anything expensive here.
type SkipInsertFunction func(ctx context.Context,
	newItemID string,
	newItemAccountID string,
	newItemBoostOfID string,
	newItemBoostOfAccountID string,
	nextItemID string,
	nextItemAccountID string,
	nextItemBoostOfID string,
	nextItemBoostOfAccountID string,
	depth int) (bool, error)

// Timeline represents a timeline for one account, and contains indexed and prepared items.
type Timeline interface {
	/*
		RETRIEVAL FUNCTIONS
	*/

	// Get returns an amount of prepared items with the given parameters.
	// If prepareNext is true, then the next predicted query will be prepared already in a goroutine,
	// to make the next call to Get faster.
	Get(ctx context.Context, amount int, maxID string, sinceID string, minID string, prepareNext bool) ([]Preparable, error)

	/*
		INDEXING + PREPARATION FUNCTIONS
	*/

	// IndexAndPrepareOne puts a item into the timeline at the appropriate place
	// according to its id, and then immediately prepares it.
	//
	// The returned bool indicates whether or not the item was actually inserted
	// into the timeline. This will be false if the item is a boost and the original
	// item, or a boost of it, already exists recently in the timeline.
	IndexAndPrepareOne(ctx context.Context, itemID string, boostOfID string, accountID string, boostOfAccountID string) (bool, error)

	// Unprepare clears the prepared version of the given item (and any boosts
	// thereof) from the timeline, but leaves the indexed version in place.
	//
	// This is useful for cache invalidation when the prepared version of the
	// item has changed for some reason (edits, updates, etc), but the item does
	// not need to be removed: it will be prepared again next time Get is called.
	Unprepare(ctx context.Context, itemID string) error

	/*
		INFO FUNCTIONS
	*/

	// TimelineID returns the id of this timeline.
	TimelineID() string

	// Len returns the length of the item index at this point in time.
	Len() int

	// OldestIndexedItemID returns the id of the rearmost (ie., the oldest) indexed item.
	// If there's no oldest item, an empty string will be returned so make sure to check for this.
	OldestIndexedItemID() string

	/*
		UTILITY FUNCTIONS
	*/

	// LastGot returns the time that Get was last called.
	LastGot() time.Time

	// Prune prunes prepared and indexed items in this timeline to the desired lengths.
	// This will be a no-op if the lengths are already < the desired values.
	//
	// The returned int indicates the amount of entries that were removed or unprepared.
	Prune(desiredPreparedItemsLength int, desiredIndexedItemsLength int) int

	// Remove removes an item with the given ID.
	//
	// If a item has multiple entries in a timeline, they will all be removed.
	//
	// The returned int indicates the amount of entries that were removed.
	Remove(ctx context.Context, itemID string) (int, error)

	// RemoveAllByOrBoosting removes all items created by or boosting the given accountID.
	//
	// The returned int indicates the amount of entries that were removed.
	RemoveAllByOrBoosting(ctx context.Context, accountID string) (int, error)
}

// timeline fulfils the Timeline interface
type timeline struct {
	items           *indexedItems
	grabFunction    GrabFunction
	filterFunction  FilterFunction
	prepareFunction PrepareFunction
	timelineID      string
	lastGot         time.Time
	sync.Mutex
}

func (t *timeline) TimelineID() string {
	return t.timelineID
}

// NewTimeline returns a new Timeline with
// the given ID, using the given functions.
func NewTimeline(
	ctx context.Context,
	timelineID string,
	grabFunction GrabFunction,
	filterFunction FilterFunction,
	prepareFunction PrepareFunction,
	skipInsertFunction SkipInsertFunction,
) Timeline {
	return &timeline{
		items: &indexedItems{
			skipInsert: skipInsertFunction,
		},
		grabFunction:    grabFunction,
		filterFunction:  filterFunction,
		prepareFunction: prepareFunction,
		timelineID:      timelineID,
		lastGot:         time.Time{},
	}
}
