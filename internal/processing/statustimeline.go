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

package processing

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

const boostReinsertionDepth = 50

// StatusGrabFunction returns a function that satisfies the GrabFunction interface in internal/timeline.
func StatusGrabFunction(database db.DB) timeline.GrabFunction {
	return func(ctx context.Context, timelineAccountID string, maxID string, sinceID string, minID string, limit int) ([]timeline.Timelineable, bool, error) {
		statuses, err := database.GetHomeTimeline(ctx, timelineAccountID, maxID, sinceID, minID, limit, false)
		if err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				return nil, true, nil // we just don't have enough statuses left in the db so return stop = true
			}
			return nil, false, fmt.Errorf("statusGrabFunction: error getting statuses from db: %w", err)
		}

		items := make([]timeline.Timelineable, len(statuses))
		for i, s := range statuses {
			items[i] = s
		}

		return items, false, nil
	}
}

// StatusFilterFunction returns a function that satisfies the FilterFunction interface in internal/timeline.
func StatusFilterFunction(database db.DB, filter *visibility.Filter) timeline.FilterFunction {
	return func(ctx context.Context, timelineAccountID string, item timeline.Timelineable) (shouldIndex bool, err error) {
		status, ok := item.(*gtsmodel.Status)
		if !ok {
			return false, errors.New("StatusFilterFunction: could not convert item to *gtsmodel.Status")
		}

		requestingAccount, err := database.GetAccountByID(ctx, timelineAccountID)
		if err != nil {
			return false, fmt.Errorf("StatusFilterFunction: error getting account with id %s: %w", timelineAccountID, err)
		}

		timelineable, err := filter.StatusHomeTimelineable(ctx, requestingAccount, status)
		if err != nil {
			return false, fmt.Errorf("StatusFilterFunction: error checking hometimelineability of status %s for account %s: %w", status.ID, timelineAccountID, err)
		}

		return timelineable, nil
	}
}

// StatusPrepareFunction returns a function that satisfies the PrepareFunction interface in internal/timeline.
func StatusPrepareFunction(database db.DB, tc typeutils.TypeConverter) timeline.PrepareFunction {
	return func(ctx context.Context, timelineAccountID string, itemID string) (timeline.Preparable, error) {
		status, err := database.GetStatusByID(ctx, itemID)
		if err != nil {
			return nil, fmt.Errorf("StatusPrepareFunction: error getting status with id %s: %w", itemID, err)
		}

		requestingAccount, err := database.GetAccountByID(ctx, timelineAccountID)
		if err != nil {
			return nil, fmt.Errorf("StatusPrepareFunction: error getting account with id %s: %w", timelineAccountID, err)
		}

		return tc.StatusToAPIStatus(ctx, status, requestingAccount)
	}
}

// StatusSkipInsertFunction returns a function that satisifes the SkipInsertFunction interface in internal/timeline.
func StatusSkipInsertFunction() timeline.SkipInsertFunction {
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

func (p *Processor) HomeTimelineGet(ctx context.Context, authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) (*apimodel.PageableResponse, gtserror.WithCode) {
	statuses, err := p.statusTimelines.GetTimeline(ctx, authed.Account.ID, maxID, sinceID, minID, limit, local)
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

func (p *Processor) PublicTimelineGet(ctx context.Context, authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) (*apimodel.PageableResponse, gtserror.WithCode) {
	statuses, err := p.state.DB.GetPublicTimeline(ctx, maxID, sinceID, minID, limit, local)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// No statuses (left) in public timeline.
			return util.EmptyPageableResponse(), nil
		}
		// An actual error has occurred.
		err = fmt.Errorf("PublicTimelineGet: db error getting statuses: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(statuses)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	var (
		items          = make([]interface{}, 0, count)
		nextMaxIDValue string
		prevMinIDValue string
	)

	for i, s := range statuses {
		// Set next + prev values before filtering and API
		// converting, so caller can still page properly.
		if i == count-1 {
			nextMaxIDValue = s.ID
		}

		if i == 0 {
			prevMinIDValue = s.ID
		}

		timelineable, err := p.filter.StatusPublicTimelineable(ctx, authed.Account, s)
		if err != nil {
			log.Debugf(ctx, "skipping status %s because of an error checking StatusPublicTimelineable: %s", s.ID, err)
			continue
		}

		if !timelineable {
			continue
		}

		apiStatus, err := p.tc.StatusToAPIStatus(ctx, s, authed.Account)
		if err != nil {
			log.Debugf(ctx, "skipping status %s because it couldn't be converted to its api representation: %s", s.ID, err)
			continue
		}

		items = append(items, apiStatus)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "api/v1/timelines/public",
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}

func (p *Processor) FavedTimelineGet(ctx context.Context, authed *oauth.Auth, maxID string, minID string, limit int) (*apimodel.PageableResponse, gtserror.WithCode) {
	statuses, nextMaxID, prevMinID, err := p.state.DB.GetFavedTimeline(ctx, authed.Account.ID, maxID, minID, limit)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// There are just no entries (left).
			return util.EmptyPageableResponse(), nil
		}
		// An actual error has occurred.
		err = fmt.Errorf("FavedTimelineGet: db error getting statuses: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(statuses)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	filtered, err := p.filterFavedStatuses(ctx, authed, statuses)
	if err != nil {
		err = fmt.Errorf("FavedTimelineGet: error filtering statuses: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	items := make([]interface{}, len(filtered))
	for i, item := range filtered {
		items[i] = item
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "api/v1/favourites",
		NextMaxIDValue: nextMaxID,
		PrevMinIDValue: prevMinID,
		Limit:          limit,
	})
}

func (p *Processor) filterFavedStatuses(ctx context.Context, authed *oauth.Auth, statuses []*gtsmodel.Status) ([]*apimodel.Status, error) {
	apiStatuses := make([]*apimodel.Status, 0, len(statuses))

	for _, s := range statuses {
		if _, err := p.state.DB.GetAccountByID(ctx, s.AccountID); err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				log.Debugf(ctx, "skipping status %s because account %s can't be found in the db", s.ID, s.AccountID)
				continue
			}
			err = fmt.Errorf("filterFavedStatuses: db error getting status author: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		timelineable, err := p.filter.StatusVisible(ctx, authed.Account, s)
		if err != nil {
			log.Debugf(ctx, "skipping status %s because of an error checking status visibility: %s", s.ID, err)
			continue
		}
		if !timelineable {
			continue
		}

		apiStatus, err := p.tc.StatusToAPIStatus(ctx, s, authed.Account)
		if err != nil {
			log.Debugf(ctx, "skipping status %s because it couldn't be converted to its api representation: %s", s.ID, err)
			continue
		}

		apiStatuses = append(apiStatuses, apiStatus)
	}

	return apiStatuses, nil
}
