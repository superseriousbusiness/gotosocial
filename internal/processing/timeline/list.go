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

// ListTimelineGrab returns a function that satisfies the GrabFunction interface in internal/timeline.
func ListTimelineGrab(database db.DB) timeline.GrabFunction {
	return func(ctx context.Context, listID string, maxID string, sinceID string, minID string, limit int) ([]timeline.Timelineable, bool, error) {
		statuses, err := database.GetListTimeline(ctx, listID, maxID, sinceID, minID, limit)
		if err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				return nil, true, nil // we just don't have enough statuses left in the db so return stop = true
			}
			return nil, false, fmt.Errorf("ListTimelineGrab: error getting statuses from db: %w", err)
		}

		items := make([]timeline.Timelineable, len(statuses))
		for i, s := range statuses {
			items[i] = s
		}

		return items, false, nil
	}
}

// HomeTimelineFilter returns a function that satisfies the FilterFunction interface in internal/timeline.
func ListTimelineFilter(database db.DB, filter *visibility.Filter) timeline.FilterFunction {
	return func(ctx context.Context, listID string, item timeline.Timelineable) (shouldIndex bool, err error) {
		status, ok := item.(*gtsmodel.Status)
		if !ok {
			return false, errors.New("ListTimelineFilter: could not convert item to *gtsmodel.Status")
		}

		list, err := database.GetListByID(ctx, listID)
		if err != nil {
			return false, fmt.Errorf("ListTimelineFilter: error getting list with id %s: %w", listID, err)
		}

		requestingAccount, err := database.GetAccountByID(ctx, list.AccountID)
		if err != nil {
			return false, fmt.Errorf("ListTimelineFilter: error getting account with id %s: %w", list.AccountID, err)
		}

		timelineable, err := filter.StatusHomeTimelineable(ctx, requestingAccount, status)
		if err != nil {
			return false, fmt.Errorf("ListTimelineFilter: error checking hometimelineability of status %s for account %s: %w", status.ID, list.AccountID, err)
		}

		return timelineable, nil
	}
}

// ListTimelineStatusPrepare returns a function that satisfies the PrepareFunction interface in internal/timeline.
func ListTimelineStatusPrepare(database db.DB, tc typeutils.TypeConverter) timeline.PrepareFunction {
	return func(ctx context.Context, listID string, itemID string) (timeline.Preparable, error) {
		status, err := database.GetStatusByID(ctx, itemID)
		if err != nil {
			return nil, fmt.Errorf("ListTimelineStatusPrepare: error getting status with id %s: %w", itemID, err)
		}

		list, err := database.GetListByID(ctx, listID)
		if err != nil {
			return nil, fmt.Errorf("ListTimelineStatusPrepare: error getting list with id %s: %w", listID, err)
		}

		requestingAccount, err := database.GetAccountByID(ctx, list.AccountID)
		if err != nil {
			return nil, fmt.Errorf("ListTimelineStatusPrepare: error getting account with id %s: %w", list.AccountID, err)
		}

		return tc.StatusToAPIStatus(ctx, status, requestingAccount)
	}
}

func (p *Processor) ListTimelineGet(ctx context.Context, authed *oauth.Auth, listID string, maxID string, sinceID string, minID string, limit int) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Ensure list exists + is owned by this account.
	list, err := p.state.DB.GetListByID(ctx, listID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	if list.AccountID != authed.Account.ID {
		err = fmt.Errorf("list with id %s does not belong to account %s", list.ID, authed.Account.ID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	statuses, err := p.ListTimelines.GetTimeline(ctx, listID, maxID, sinceID, minID, limit, false)
	if err != nil {
		err = fmt.Errorf("ListTimelineGet: error getting statuses: %w", err)
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
		Path:           "api/v1/timelines/list/" + listID,
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}
