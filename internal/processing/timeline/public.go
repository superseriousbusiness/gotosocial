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
	"strconv"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *Processor) PublicTimelineGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	maxID string,
	sinceID string,
	minID string,
	limit int,
	local bool,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	const maxAttempts = 3
	var (
		nextMaxIDValue string
		prevMinIDValue string
		items          = make([]any, 0, limit)
	)

	var filters []*gtsmodel.Filter
	var compiledMutes *usermute.CompiledUserMuteList
	if requester != nil {
		var err error
		filters, err = p.state.DB.GetFiltersForAccountID(ctx, requester.ID)
		if err != nil {
			err = gtserror.Newf("couldn't retrieve filters for account %s: %w", requester.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		mutes, err := p.state.DB.GetAccountMutes(gtscontext.SetBarebones(ctx), requester.ID, nil)
		if err != nil {
			err = gtserror.Newf("couldn't retrieve mutes for account %s: %w", requester.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		compiledMutes = usermute.NewCompiledUserMuteList(mutes)
	}

	// Try a few times to select appropriate public
	// statuses from the db, paging up or down to
	// reattempt if nothing suitable is found.
outer:
	for attempts := 1; ; attempts++ {
		// Select slightly more than the limit to try to avoid situations where
		// we filter out all the entries, and have to make another db call.
		// It's cheaper to select more in 1 query than it is to do multiple queries.
		statuses, err := p.state.DB.GetPublicTimeline(ctx, maxID, sinceID, minID, limit+5, local)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err = gtserror.Newf("db error getting statuses: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		count := len(statuses)
		if count == 0 {
			// Nothing relevant (left) in the db.
			return util.EmptyPageableResponse(), nil
		}

		// Page up from first status in slice
		// (ie., one with the highest ID).
		prevMinIDValue = statuses[0].ID

	inner:
		for _, s := range statuses {
			// Push back the next page down ID to
			// this status, regardless of whether
			// we end up filtering it out or not.
			nextMaxIDValue = s.ID

			timelineable, err := p.visFilter.StatusPublicTimelineable(ctx, requester, s)
			if err != nil {
				log.Errorf(ctx, "error checking status visibility: %v", err)
				continue inner
			}

			if !timelineable {
				continue inner
			}

			apiStatus, err := p.converter.StatusToAPIStatus(ctx, s, requester, statusfilter.FilterContextPublic, filters, compiledMutes)
			if errors.Is(err, statusfilter.ErrHideStatus) {
				continue
			}
			if err != nil {
				log.Errorf(ctx, "error converting to api status: %v", err)
				continue inner
			}

			// Looks good, add this.
			items = append(items, apiStatus)

			// We called the db with a little
			// more than the desired limit.
			//
			// Ensure we don't return more
			// than the caller asked for.
			if len(items) == limit {
				break outer
			}
		}

		if len(items) != 0 {
			// We've got some items left after
			// filtering, happily break + return.
			break
		}

		if attempts >= maxAttempts {
			// We reached our attempts limit.
			// Be nice + warn about it.
			log.Warn(ctx, "reached max attempts to find items in public timeline")
			break
		}

		// We filtered out all items before we
		// found anything we could return, but
		// we still have attempts left to try
		// fetching again. Set paging params
		// and allow loop to continue.
		if minID != "" {
			// Paging up.
			minID = prevMinIDValue
		} else {
			// Paging down.
			maxID = nextMaxIDValue
		}
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "/api/v1/timelines/public",
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
		ExtraQueryParams: []string{
			"local=" + strconv.FormatBool(local),
		},
	})
}
