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
	"net/url"
	"slices"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/cache/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
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
	postFilter func(*timeline.StatusMeta) bool,
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
	statuses, err := timeline.Load(ctx,
		page,

		// ...
		loadPage,

		// ...
		func(ids []string) ([]*gtsmodel.Status, error) {
			return p.state.DB.GetStatusesByIDs(ctx, ids)
		},

		// ...
		preFilter,

		// ...
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
		panic(err)
	}
}

func (p *Processor) getTimeline(
	ctx context.Context,
	requester *gtsmodel.Account,
	timeline *cache.TimelineCache[*gtsmodel.Status],
	page *paging.Page,
	pgPath string, // timeline page path
	pgQuery url.Values, // timeline query parameters
	filterCtx statusfilter.FilterContext,
	load func(*paging.Page) (statuses []*gtsmodel.Status, next *paging.Page, err error), // timeline cache load function
	filter func(*gtsmodel.Status) bool, // per-request filtering function, done AFTER timeline caching
) (
	*apimodel.PageableResponse,
	gtserror.WithCode,
) {
	// Load timeline with cache / loader funcs.
	statuses, errWithCode := p.loadTimeline(ctx,
		timeline,
		page,
		load,
		filter,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if len(statuses) == 0 {
		// Check for an empty timeline rsp.
		return paging.EmptyResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := statuses[len(statuses)-1].ID
	hi := statuses[0].ID

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

	// NOTE:
	// Right now this is not ideal, as we perform mute and
	// status filtering *after* the above load loop, so we
	// could end up with no statuses still AFTER all loading.
	//
	// In a PR coming *soon* we will move the filtering and
	// status muting into separate module similar to the visibility
	// filtering and caching which should move it to the above
	// load loop and provided function.

	// API response requires them in interface{} form.
	items := make([]interface{}, 0, len(statuses))

	for _, status := range statuses {
		// Convert internal status model to frontend model.
		apiStatus, err := p.converter.StatusToAPIStatus(ctx,
			status,
			requester,
			filterCtx,
			filters,
			mutes,
		)
		if err != nil && !errors.Is(err, statusfilter.ErrHideStatus) {
			log.Errorf(ctx, "error converting status: %v", err)
			continue
		}

		if apiStatus != nil {
			// Append status to return slice.
			items = append(items, apiStatus)
		}
	}

	// Package converted API statuses as pageable response.
	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Path:  pgPath,
		Query: pgQuery,
	}), nil
}

func (p *Processor) loadTimeline(
	ctx context.Context,
	timeline *cache.TimelineCache[*gtsmodel.Status],
	page *paging.Page,
	load func(*paging.Page) (statuses []*gtsmodel.Status, next *paging.Page, err error),
	filter func(*gtsmodel.Status) bool,
) (
	[]*gtsmodel.Status,
	gtserror.WithCode,
) {
	if load == nil {
		// nil check outside
		// below main loop.
		panic("nil func")
	}

	if page == nil {
		const text = "timeline must be paged"
		return nil, gtserror.NewErrorBadRequest(
			errors.New(text),
			text,
		)
	}

	// Try load statuses from cache.
	statuses := timeline.Select(page)

	// Filter statuses using provided function.
	statuses = slices.DeleteFunc(statuses, filter)

	// Check if more statuses need to be loaded.
	if limit := page.Limit; len(statuses) < limit {

		// Set first page
		// query to load.
		nextPg := page

		for i := 0; i < 5; i++ {
			var err error
			var next []*gtsmodel.Status

			// Load next timeline statuses.
			next, nextPg, err = load(nextPg)
			if err != nil {
				err := gtserror.Newf("error loading timeline: %w", err)
				return nil, gtserror.NewErrorInternalError(err)
			}

			// An empty next page means no more.
			if len(next) == 0 && nextPg == nil {
				break
			}

			// Cache loaded statuses.
			timeline.Insert(next...)

			// Filter statuses using provided function,
			// this must be done AFTER cache insert but
			// BEFORE adding to slice, as this is used
			// for request-specific timeline filtering,
			// as opposed to filtering for entire cache.
			next = slices.DeleteFunc(next, filter)

			// Append loaded statuses to return.
			statuses = append(statuses, next...)

			if len(statuses) >= limit {
				// We loaded all the statuses
				// that were requested of us!
				break
			}
		}
	}

	return statuses, nil
}
