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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// ListTimelineGrab returns a function that satisfies GrabFunction for list timelines.
func ListTimelineGrab(state *state.State) timeline.GrabFunction {
	return func(ctx context.Context, listID string, maxID string, sinceID string, minID string, limit int) ([]timeline.Timelineable, bool, error) {
		statuses, err := state.DB.GetListTimeline(ctx, listID, maxID, sinceID, minID, limit)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err = gtserror.Newf("error getting statuses from db: %w", err)
			return nil, false, err
		}

		count := len(statuses)
		if count == 0 {
			// We just don't have enough statuses
			// left in the db so return stop = true.
			return nil, true, nil
		}

		items := make([]timeline.Timelineable, count)
		for i, s := range statuses {
			items[i] = s
		}

		return items, false, nil
	}
}

// ListTimelineFilter returns a function that satisfies FilterFunction for list timelines.
func ListTimelineFilter(state *state.State, visFilter *visibility.Filter) timeline.FilterFunction {
	return func(ctx context.Context, listID string, item timeline.Timelineable) (shouldIndex bool, err error) {
		status, ok := item.(*gtsmodel.Status)
		if !ok {
			err = gtserror.New("could not convert item to *gtsmodel.Status")
			return false, err
		}

		list, err := state.DB.GetListByID(ctx, listID)
		if err != nil {
			err = gtserror.Newf("error getting list with id %s: %w", listID, err)
			return false, err
		}

		requestingAccount, err := state.DB.GetAccountByID(ctx, list.AccountID)
		if err != nil {
			err = gtserror.Newf("error getting account with id %s: %w", list.AccountID, err)
			return false, err
		}

		timelineable, err := visFilter.StatusHomeTimelineable(ctx, requestingAccount, status)
		if err != nil {
			err = gtserror.Newf("error checking hometimelineability of status %s for account %s: %w", status.ID, list.AccountID, err)
			return false, err
		}

		return timelineable, nil
	}
}

// ListTimelineStatusPrepare returns a function that satisfies PrepareFunction for list timelines.
func ListTimelineStatusPrepare(state *state.State, converter *typeutils.Converter) timeline.PrepareFunction {
	return func(ctx context.Context, listID string, itemID string) (timeline.Preparable, error) {
		status, err := state.DB.GetStatusByID(ctx, itemID)
		if err != nil {
			err = gtserror.Newf("error getting status with id %s: %w", itemID, err)
			return nil, err
		}

		list, err := state.DB.GetListByID(ctx, listID)
		if err != nil {
			err = gtserror.Newf("error getting list with id %s: %w", listID, err)
			return nil, err
		}

		requestingAccount, err := state.DB.GetAccountByID(ctx, list.AccountID)
		if err != nil {
			err = gtserror.Newf("error getting account with id %s: %w", list.AccountID, err)
			return nil, err
		}

		filters, err := state.DB.GetFiltersForAccountID(ctx, requestingAccount.ID)
		if err != nil {
			err = gtserror.Newf("couldn't retrieve filters for account %s: %w", requestingAccount.ID, err)
			return nil, err
		}

		mutes, err := state.DB.GetAccountMutes(gtscontext.SetBarebones(ctx), requestingAccount.ID, nil)
		if err != nil {
			err = gtserror.Newf("couldn't retrieve mutes for account %s: %w", requestingAccount.ID, err)
			return nil, err
		}
		compiledMutes := usermute.NewCompiledUserMuteList(mutes)

		return converter.StatusToAPIStatus(ctx, status, requestingAccount, statusfilter.FilterContextHome, filters, compiledMutes)
	}
}

func (p *Processor) ListTimelineGet(ctx context.Context, authed *apiutil.Auth, listID string, maxID string, sinceID string, minID string, limit int) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Ensure list exists + is owned by this account.
	list, err := p.state.DB.GetListByID(ctx, listID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	if list.AccountID != authed.Account.ID {
		err = gtserror.Newf("list with id %s does not belong to account %s", list.ID, authed.Account.ID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	statuses, err := p.state.Timelines.List.GetTimeline(ctx, listID, maxID, sinceID, minID, limit, false)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("error getting statuses: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(statuses)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	var (
		items          = make([]interface{}, count)
		nextMaxIDValue = statuses[count-1].GetID()
		prevMinIDValue = statuses[0].GetID()
	)

	for i := range statuses {
		items[i] = statuses[i]
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "/api/v1/timelines/list/" + listID,
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}
