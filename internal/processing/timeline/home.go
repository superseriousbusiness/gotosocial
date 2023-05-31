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
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

// HomeTimelineGrab returns a function that satisfies GrabFunction for home timelines.
func HomeTimelineGrab(state *state.State) timeline.GrabFunction {
	return func(ctx context.Context, accountID string, maxID string, sinceID string, minID string, limit int) ([]timeline.Timelineable, bool, error) {
		statuses, err := state.DB.GetHomeTimeline(ctx, accountID, maxID, sinceID, minID, limit, false)
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

// HomeTimelineFilter returns a function that satisfies FilterFunction for home timelines.
func HomeTimelineFilter(state *state.State, filter *visibility.Filter) timeline.FilterFunction {
	return func(ctx context.Context, accountID string, item timeline.Timelineable) (shouldIndex bool, err error) {
		status, ok := item.(*gtsmodel.Status)
		if !ok {
			return false, errors.New("HomeTimelineFilter: could not convert item to *gtsmodel.Status")
		}

		requestingAccount, err := state.DB.GetAccountByID(ctx, accountID)
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

// HomeTimelineStatusPrepare returns a function that satisfies PrepareFunction for home timelines.
func HomeTimelineStatusPrepare(state *state.State, tc typeutils.TypeConverter) timeline.PrepareFunction {
	return func(ctx context.Context, accountID string, itemID string) (timeline.Preparable, error) {
		status, err := state.DB.GetStatusByID(ctx, itemID)
		if err != nil {
			return nil, fmt.Errorf("StatusPrepare: error getting status with id %s: %w", itemID, err)
		}

		requestingAccount, err := state.DB.GetAccountByID(ctx, accountID)
		if err != nil {
			return nil, fmt.Errorf("StatusPrepare: error getting account with id %s: %w", accountID, err)
		}

		return tc.StatusToAPIStatus(ctx, status, requestingAccount)
	}
}

func (p *Processor) HomeTimelineGet(ctx context.Context, authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) (*apimodel.PageableResponse, gtserror.WithCode) {
	statuses, err := p.state.Timelines.Home.GetTimeline(ctx, authed.Account.ID, maxID, sinceID, minID, limit, local)
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
