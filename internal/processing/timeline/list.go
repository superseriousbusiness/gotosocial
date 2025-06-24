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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// ListTimelineGet gets a pageable timeline of statuses
// in the list timeline of ID by the requesting account.
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
	if list == nil {
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

	// Fetch status timeline for list.
	return p.getStatusTimeline(ctx,

		// Auth'd
		// account.
		requester,

		// Keyed-by-list-ID, list timeline cache.
		p.state.Caches.Timelines.List.MustGet(listID),

		// Current
		// page.
		page,

		// List timeline ID's endpoint.
		"/api/v1/timelines/list/"+listID,

		// No page
		// query.
		nil,

		// Status filter context.
		gtsmodel.FilterContextHome,

		// Database load function.
		func(pg *paging.Page) (statuses []*gtsmodel.Status, err error) {
			return p.state.DB.GetListTimeline(ctx, listID, pg)
		},

		// Filtering function,
		// i.e. filter before caching.
		func(s *gtsmodel.Status) bool {

			// Check the visibility of passed status to requesting user.
			ok, err := p.visFilter.StatusHomeTimelineable(ctx, requester, s)
			if err != nil {
				log.Errorf(ctx, "error checking status %s visibility: %v", s.URI, err)
				return true // default assume not visible
			} else if !ok {
				return true
			}

			// Check if status been muted by requester from timelines.
			muted, err := p.muteFilter.StatusMuted(ctx, requester, s)
			if err != nil {
				log.Errorf(ctx, "error checking status %s mutes: %v", s.URI, err)
				return true // default assume muted
			} else if muted {
				return true
			}

			return false
		},

		// Post filtering funtion,
		// i.e. filter after caching.
		nil,
	)
}
