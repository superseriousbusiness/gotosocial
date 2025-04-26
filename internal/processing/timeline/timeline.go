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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	timelinepkg "code.superseriousbusiness.org/gotosocial/internal/cache/timeline"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	statusfilter "code.superseriousbusiness.org/gotosocial/internal/filter/status"
	"code.superseriousbusiness.org/gotosocial/internal/filter/usermute"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
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
	timeline *timelinepkg.StatusTimeline,
	page *paging.Page,
	pagePath string,
	pageQuery url.Values,
	filterCtx statusfilter.FilterContext,
	loadPage func(*paging.Page) (statuses []*gtsmodel.Status, err error),
	filter func(*gtsmodel.Status) (delete bool),
	postFilter func(*gtsmodel.Status) (remove bool),
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

	// Load status page via timeline cache, also
	// getting lo, hi values for next, prev pages.
	//
	// NOTE: this safely handles the case of a nil
	// input timeline, i.e. uncached timeline type.
	apiStatuses, lo, hi, err := timeline.Load(ctx,

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

		// Call provided status
		// filtering function.
		filter,

		// Frontend API model preparation function.
		func(status *gtsmodel.Status) (*apimodel.Status, error) {

			// Check if status needs filtering OUTSIDE of caching stage.
			// TODO: this will be moved to separate postFilter hook when
			// all filtering has been removed from the type converter.
			if postFilter != nil && postFilter(status) {
				return nil, nil
			}

			// Finally, pass status to get converted to API model.
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
		Path:  pagePath,
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: pageQuery,
	}), nil
}
