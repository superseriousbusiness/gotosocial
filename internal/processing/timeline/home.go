package timeline

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

// HomeTimelineGrab returns a function that satisfies the GrabFunction interface in internal/timeline.
func HomeTimelineGrab(database db.DB) timeline.GrabFunction {
	return func(ctx context.Context, accountID string, maxID string, sinceID string, minID string, limit int) ([]timeline.Timelineable, bool, error) {
		statuses, err := database.GetHomeTimeline(ctx, accountID, maxID, sinceID, minID, limit, false)
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
	return func(ctx context.Context, accountID string, item timeline.Timelineable) (shouldIndex bool, err error) {
		status, ok := item.(*gtsmodel.Status)
		if !ok {
			return false, errors.New("HomeTimelineFilter: could not convert item to *gtsmodel.Status")
		}

		requestingAccount, err := database.GetAccountByID(ctx, accountID)
		if err != nil {
			return false, fmt.Errorf("HomeTimelineFilter: error getting account with id %s: %w", accountID, err)
		}

		timelineable, err := filter.StatusHomeTimelineable(ctx, requestingAccount, status)
		if err != nil {
			return false, fmt.Errorf("HomeTimelineFilter: error checking hometimelineability of status %s for account %s: %w", status.ID, accountID, err)
		}

		return timelineable, nil
	}
}

// HomeTimelineStatusPrepare returns a function that satisfies the PrepareFunction interface in internal/timeline.
func HomeTimelineStatusPrepare(database db.DB, tc typeutils.TypeConverter) timeline.PrepareFunction {
	return func(ctx context.Context, accountID string, itemID string) (timeline.Preparable, error) {
		status, err := database.GetStatusByID(ctx, itemID)
		if err != nil {
			return nil, fmt.Errorf("StatusPrepare: error getting status with id %s: %w", itemID, err)
		}

		requestingAccount, err := database.GetAccountByID(ctx, accountID)
		if err != nil {
			return nil, fmt.Errorf("StatusPrepare: error getting account with id %s: %w", accountID, err)
		}

		return tc.StatusToAPIStatus(ctx, status, requestingAccount)
	}
}

func (p *Processor) HomeTimelineGet(ctx context.Context, authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) (*apimodel.PageableResponse, gtserror.WithCode) {
	statuses, err := p.HomeTimelines.GetTimeline(ctx, authed.Account.ID, maxID, sinceID, minID, limit, local)
	if err != nil {
		err = fmt.Errorf("HomeTimelineGet: error getting statuses: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(statuses)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	var (
		items          = make([]interface{}, count)
		nextMaxIDValue string
		prevMinIDValue string
	)

	for i, item := range statuses {
		if i == count-1 {
			nextMaxIDValue = item.GetID()
		}

		if i == 0 {
			prevMinIDValue = item.GetID()
		}

		items[i] = item
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "api/v1/timelines/home",
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}
