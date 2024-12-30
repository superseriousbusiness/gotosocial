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
	"strconv"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// PublicTimelineGet ...
func (p *Processor) PublicTimelineGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	page *paging.Page,
	local bool,
) (
	*apimodel.PageableResponse,
	gtserror.WithCode,
) {

	// Load timeline data.
	return p.getTimeline(ctx,

		// Auth'd
		// account.
		requester,

		// Global public timeline cache.
		&p.state.Caches.Timelines.Public,

		// Current
		// page.
		page,

		// Public timeline endpoint.
		"/api/v1/timelines/public",

		// Set local-only timeline page query flag.
		url.Values{"local": {strconv.FormatBool(local)}},

		// Status filter context.
		statusfilter.FilterContextPublic,

		// Timeline cache load function, used to further hydrate cache where necessary.
		func(page *paging.Page) (statuses []*gtsmodel.Status, next *paging.Page, err error) {

			// Fetch the global public status timeline page.
			statuses, err = p.state.DB.GetPublicTimeline(ctx,
				page,
			)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				return nil, nil, gtserror.Newf("error getting statuses: %w", err)
			}

			if len(statuses) == 0 {
				// No more to load.
				return nil, nil, nil
			}

			// Get the lowest and highest
			// ID values, used for next pg.
			lo := statuses[len(statuses)-1].ID
			hi := statuses[0].ID

			// Set next paging value.
			page = page.Next(lo, hi)

			for i := 0; i < len(statuses); {
				// Get status at idx.
				status := statuses[i]

				// Check whether status should be show on public timeline.
				visible, err := p.visFilter.StatusPublicTimelineable(ctx,
					requester,
					status,
				)
				if err != nil {
					return nil, nil, gtserror.Newf("error checking visibility: %w", err)
				}

				if !visible {
					// Status not visible to home timeline.
					statuses = slices.Delete(statuses, i, i+1)
					continue
				}

				// Iter.
				i++
			}

			return
		},

		// Per-request filtering function.
		func(s *gtsmodel.Status) bool {
			if local {
				return !*s.Local
			}
			return false
		},
	)
}
