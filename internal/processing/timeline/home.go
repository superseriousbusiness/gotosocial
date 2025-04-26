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
	"net/url"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	statusfilter "code.superseriousbusiness.org/gotosocial/internal/filter/status"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// HomeTimelineGet gets a pageable timeline of statuses
// in the home timeline of the requesting account.
func (p *Processor) HomeTimelineGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	page *paging.Page,
	local bool,
) (
	*apimodel.PageableResponse,
	gtserror.WithCode,
) {

	var pageQuery url.Values
	var postFilter func(*gtsmodel.Status) bool
	if local {
		// Set local = true query.
		pageQuery = localOnlyTrue
		postFilter = func(s *gtsmodel.Status) bool {
			return !*s.Local
		}
	} else {
		// Set local = false query.
		pageQuery = localOnlyFalse
		postFilter = nil
	}
	return p.getStatusTimeline(ctx,

		// Auth'd
		// account.
		requester,

		// Keyed-by-account-ID, home timeline cache.
		p.state.Caches.Timelines.Home.MustGet(requester.ID),

		// Current
		// page.
		page,

		// Home timeline endpoint.
		"/api/v1/timelines/home",

		// Set local-only timeline
		// page query flag, (this map
		// later gets copied before
		// any further usage).
		pageQuery,

		// Status filter context.
		statusfilter.FilterContextHome,

		// Database load function.
		func(pg *paging.Page) (statuses []*gtsmodel.Status, err error) {
			return p.state.DB.GetHomeTimeline(ctx, requester.ID, pg)
		},

		// Filtering function,
		// i.e. filter before caching.
		func(s *gtsmodel.Status) bool {

			// Check the visibility of passed status to requesting user.
			ok, err := p.visFilter.StatusHomeTimelineable(ctx, requester, s)
			if err != nil {
				log.Errorf(ctx, "error filtering status %s: %v", s.URI, err)
			}
			return !ok
		},

		// Post filtering funtion,
		// i.e. filter after caching.
		postFilter,
	)
}
