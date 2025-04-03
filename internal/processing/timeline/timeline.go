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
	"github.com/superseriousbusiness/gotosocial/internal/id"
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
	cache *timeline.StatusTimeline,
	page *paging.Page,
	pagePath string,
	pageQuery url.Values,
	filterCtx statusfilter.FilterContext,
	loadPage func(*paging.Page) (statuses []*gtsmodel.Status, err error),
	filter func(*gtsmodel.Status) (bool, error),
) (
	*apimodel.PageableResponse,
	gtserror.WithCode,
) {
	var err error
	var filters []*gtsmodel.Filter
	var mutes *usermute.CompiledUserMuteList

	if requester != nil {

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
			nil, // i.e. all
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err := gtserror.Newf("error getting account mutes: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Compile all account mutes to useable form.
		mutes = usermute.NewCompiledUserMuteList(allMutes)
	}

	// Ensure we have valid
	// input paging cursor.
	id.ValidatePage(page)

	// Returned models and page params.
	var apiStatuses []*apimodel.Status
	var lo, hi string

	if cache != nil {
		// Load status page via timeline cache, also
		// getting lo, hi values for next, prev pages.
		apiStatuses, lo, hi, err = cache.Load(ctx,

			// Status page
			// to load.
			page,

			// Caller provided database
			// status page loading function.
			loadPage,

			// Status load function for cached timeline entries.
			func(ids []string) ([]*gtsmodel.Status, error) {
				return p.state.DB.GetStatusesByIDs(ctx, ids)
			},

			// Filtering function,
			// i.e. filter before caching.
			filter,

			// Frontend API model preparation function.
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
	} else {
		// Load status page without a receiving timeline cache.
		// TODO: remove this code path when all support caching.
		apiStatuses, lo, hi, err = timeline.LoadStatusTimeline(ctx,
			page,
			loadPage,
			func(ids []string) ([]*gtsmodel.Status, error) {
				return p.state.DB.GetStatusesByIDs(ctx, ids)
			},
			filter,
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
	}

	if err != nil {
		err := gtserror.Newf("error loading timeline: %w", err)
		return nil, gtserror.WrapWithCode(http.StatusInternalServerError, err)
	}

	// Package returned API statuses as pageable response.
	return paging.PackageResponse(paging.ResponseParams{
		Items: xslices.ToAny(apiStatuses),
		Path:  pagePath,
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: pageQuery,
	}), nil
}
