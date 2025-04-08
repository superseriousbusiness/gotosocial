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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
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
		func() url.Values {
			var pageQuery url.Values

			if local {
				// Set local = true query.
				pageQuery = localOnlyTrue
			} else {
				// Set local = false query.
				pageQuery = localOnlyFalse
			}

			return pageQuery
		}(),

		// Status filter context.
		statusfilter.FilterContextHome,

		// Database load function.
		func(pg *paging.Page) (statuses []*gtsmodel.Status, err error) {
			return p.state.DB.GetHomeTimeline(ctx, requester.ID, pg)
		},

		// Pre-filtering function,
		// i.e. filter before caching.
		func(s *gtsmodel.Status) (bool, error) {

			// Check the visibility of passed status to requesting user.
			ok, err := p.visFilter.StatusHomeTimelineable(ctx, requester, s)
			return !ok, err
		},
	)
}
