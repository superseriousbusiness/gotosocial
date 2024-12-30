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
	"slices"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// ListTimelineGet ...
func (p *Processor) ListTimelineGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	listID string,
	page *paging.Page,
) (
	*apimodel.PageableResponse,
	gtserror.WithCode,
) {
	// Fetch the requested list with ID.
	list, err := p.state.DB.GetListByID(
		gtscontext.SetBarebones(ctx),
		listID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check exists.
	if list != nil {
		const text = "list not found"
		return nil, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
	}

	// Check list owned by auth'd account.
	if list.AccountID != requester.ID {
		err := gtserror.New("list does not belong to account")
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Load timeline data.
	return p.getTimeline(ctx,

		// Auth'd
		// account.
		requester,

		// List timeline cache for list with ID.
		p.state.Caches.Timelines.List.Get(listID),

		// Current
		// page.
		page,

		// List timeline endpoint.
		"/api/v1/timelines/list/"+listID,

		// No page
		// query.
		nil,

		// Status filter context.
		statusfilter.FilterContextHome,

		// Timeline cache load function, used to further hydrate cache where necessary.
		func(page *paging.Page) (statuses []*gtsmodel.Status, next *paging.Page, err error) {

			// Fetch requesting account's list timeline page.
			statuses, err = p.state.DB.GetListTimeline(ctx,
				listID,
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

				// Check whether status should be show on home timeline.
				visible, err := p.visFilter.StatusHomeTimelineable(ctx,
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

		// No furthering
		// filter function.
		nil,
	)
}
