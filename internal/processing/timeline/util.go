package timeline

import (
	"context"
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

const boostReinsertionDepth = 50

// HomeTimelineGrab returns a function that satisfies the GrabFunction interface in internal/timeline.
func HomeTimelineGrab(database db.DB) timeline.GrabFunction {
	return func(ctx context.Context, timelineAccountID string, maxID string, sinceID string, minID string, limit int) ([]timeline.Timelineable, bool, error) {
		statuses, err := database.GetHomeTimeline(ctx, timelineAccountID, maxID, sinceID, minID, limit, false)
		if err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				return nil, true, nil // we just don't have enough statuses left in the db so return stop = true
			}
			return nil, false, fmt.Errorf("HomeTimelineGrab: error getting statuses from db: %w", err)
		}

		items := make([]timeline.Timelineable, len(statuses))
		for i, s := range statuses {
			items[i] = s
		}

		return items, false, nil
	}
}

// HomeTimelineFilter returns a function that satisfies the FilterFunction interface in internal/timeline.
func HomeTimelineFilter(database db.DB, filter *visibility.Filter) timeline.FilterFunction {
	return func(ctx context.Context, timelineAccountID string, item timeline.Timelineable) (shouldIndex bool, err error) {
		status, ok := item.(*gtsmodel.Status)
		if !ok {
			return false, errors.New("HomeTimelineFilter: could not convert item to *gtsmodel.Status")
		}

		requestingAccount, err := database.GetAccountByID(ctx, timelineAccountID)
		if err != nil {
			return false, fmt.Errorf("HomeTimelineFilter: error getting account with id %s: %w", timelineAccountID, err)
		}

		timelineable, err := filter.StatusHomeTimelineable(ctx, requestingAccount, status)
		if err != nil {
			return false, fmt.Errorf("HomeTimelineFilter: error checking hometimelineability of status %s for account %s: %w", status.ID, timelineAccountID, err)
		}

		return timelineable, nil
	}
}

// HomeTimelineSkipInsert returns a function that satisifes the SkipInsertFunction interface in internal/timeline.
func HomeTimelineSkipInsert() timeline.SkipInsertFunction {
	return func(
		ctx context.Context,
		newItemID string,
		newItemAccountID string,
		newItemBoostOfID string,
		newItemBoostOfAccountID string,
		nextItemID string,
		nextItemAccountID string,
		nextItemBoostOfID string,
		nextItemBoostOfAccountID string,
		depth int,
	) (bool, error) {
		// make sure we don't insert a duplicate
		if newItemID == nextItemID {
			return true, nil
		}

		// check if it's a boost
		if newItemBoostOfID != "" {
			// skip if we've recently put another boost of this status in the timeline
			if newItemBoostOfID == nextItemBoostOfID {
				if depth < boostReinsertionDepth {
					return true, nil
				}
			}

			// skip if we've recently put the original status in the timeline
			if newItemBoostOfID == nextItemID {
				if depth < boostReinsertionDepth {
					return true, nil
				}
			}
		}

		// insert the item
		return false, nil
	}
}

// StatusPrepare returns a function that satisfies the PrepareFunction interface in internal/timeline.
func StatusPrepare(database db.DB, tc typeutils.TypeConverter) timeline.PrepareFunction {
	return func(ctx context.Context, timelineAccountID string, itemID string) (timeline.Preparable, error) {
		status, err := database.GetStatusByID(ctx, itemID)
		if err != nil {
			return nil, fmt.Errorf("StatusPrepare: error getting status with id %s: %w", itemID, err)
		}

		requestingAccount, err := database.GetAccountByID(ctx, timelineAccountID)
		if err != nil {
			return nil, fmt.Errorf("StatusPrepare: error getting account with id %s: %w", timelineAccountID, err)
		}

		return tc.StatusToAPIStatus(ctx, status, requestingAccount)
	}
}
