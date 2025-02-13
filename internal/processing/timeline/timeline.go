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
	"net/http"
	"net/url"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/cache/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
)

var (
	// pre-prepared URL values to be passed in to
	// paging response forms. The paging package always
	// copies values before any modifications so it's
	// safe to only use a single map variable for these.
	localOnlyTrue  = url.Values{"local": {"true"}}
	localOnlyFalse = url.Values{"local": {"false"}}
)

type Processor struct {
	state     *state.State
	converter *typeutils.Converter
	visFilter *visibility.Filter
}

func New(state *state.State, converter *typeutils.Converter, visFilter *visibility.Filter) Processor {
	return Processor{
		state:     state,
		converter: converter,
		visFilter: visFilter,
	}
}

func (p *Processor) getStatusTimeline(
	ctx context.Context,
	requester *gtsmodel.Account,
	timeline *timeline.StatusTimeline,
	page *paging.Page,
	pgPath string, // timeline page path
	pgQuery url.Values, // timeline query parameters
	filterCtx statusfilter.FilterContext,
	loadPage func(*paging.Page) (statuses []*gtsmodel.Status, err error),
	preFilter func(*gtsmodel.Status) (bool, error),
	postFilter func(*gtsmodel.Status) (bool, error),
) (
	*apimodel.PageableResponse,
	gtserror.WithCode,
) {
	var (
		filters []*gtsmodel.Filter
		mutes   *usermute.CompiledUserMuteList
	)

	if requester != nil {
		var err error

		// Fetch all filters relevant for requesting account.
		filters, err = p.state.DB.GetFiltersForAccountID(ctx,
			requester.ID,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err := gtserror.Newf("error getting account filters: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Get a list of all account mutes for requester.
		allMutes, err := p.state.DB.GetAccountMutes(ctx,
			requester.ID,
			nil, // nil page, i.e. all
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err := gtserror.Newf("error getting account mutes: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Compile all account mutes to useable form.
		mutes = usermute.NewCompiledUserMuteList(allMutes)
	}

	// ...
	apiStatuses, lo, hi, err := timeline.Load(ctx,

		page,

		// ...
		loadPage,

		// ...
		func(ids []string) ([]*gtsmodel.Status, error) {
			return p.state.DB.GetStatusesByIDs(ctx, ids)
		},

		// Pre-filtering function,
		// i.e. filter before caching.
		preFilter,

		// Post-filtering function,
		// i.e. filter after caching.
		postFilter,

		// ...
		func(status *gtsmodel.Status) (*apimodel.Status, error) {
			apiStatus, err := p.converter.StatusToAPIStatus(ctx,
				status,
				requester,
				filterCtx,
				filters,
				mutes,
			)
			if err != nil && !errors.Is(err, statusfilter.ErrHideStatus) {
				return nil, err
			}
			return apiStatus, nil
		},
	)
	if err != nil {
		err := gtserror.Newf("error loading timeline: %w", err)
		return nil, gtserror.WrapWithCode(http.StatusInternalServerError, err)
	}

	// Package returned API statuses as pageable response.
	return paging.PackageResponse(paging.ResponseParams{
		Items: xslices.ToAny(apiStatuses),
		Path:  pgPath,
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: pgQuery,
	}), nil
}
