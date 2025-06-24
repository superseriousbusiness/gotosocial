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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// PublicTimelineGet gets a pageable timeline of public statuses
// for the given requesting account. It ensures that each status
// in timeline is visible to the account before returning it.
//
// The local argument limits this to local-only statuses.
func (p *Processor) PublicTimelineGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	page *paging.Page,
	local bool,
) (
	*apimodel.PageableResponse,
	gtserror.WithCode,
) {
	if local {
		return p.localTimelineGet(ctx, requester, page)
	}
	return p.publicTimelineGet(ctx, requester, page)
}

func (p *Processor) publicTimelineGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	page *paging.Page,
) (
	*apimodel.PageableResponse,
	gtserror.WithCode,
) {
	return p.getStatusTimeline(ctx,

		// Auth acconut,
		// can be nil.
		requester,

		// No cache.
		nil,

		// Current
		// page.
		page,

		// Public timeline endpoint.
		"/api/v1/timelines/public",

		// Set local-only timeline
		// page query flag, (this map
		// later gets copied before
		// any further usage).
		localOnlyFalse,

		// Status filter context.
		gtsmodel.FilterContextPublic,

		// Database load function.
		func(pg *paging.Page) (statuses []*gtsmodel.Status, err error) {
			return p.state.DB.GetPublicTimeline(ctx, pg)
		},

		// Pre-filtering function,
		// i.e. filter before caching.
		func(s *gtsmodel.Status) bool {

			// Check the visibility of passed status to requesting user.
			ok, err := p.visFilter.StatusPublicTimelineable(ctx, requester, s)
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

func (p *Processor) localTimelineGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	page *paging.Page,
) (
	*apimodel.PageableResponse,
	gtserror.WithCode,
) {
	return p.getStatusTimeline(ctx,

		// Auth acconut,
		// can be nil.
		requester,

		// No cache.
		nil,

		// Current
		// page.
		page,

		// Public timeline endpoint.
		"/api/v1/timelines/public",

		// Set local-only timeline
		// page query flag, (this map
		// later gets copied before
		// any further usage).
		localOnlyTrue,

		// Status filter context.
		gtsmodel.FilterContextPublic,

		// Database load function.
		func(pg *paging.Page) (statuses []*gtsmodel.Status, err error) {
			return p.state.DB.GetLocalTimeline(ctx, pg)
		},

		// Filtering function,
		// i.e. filter before caching.
		func(s *gtsmodel.Status) bool {

			// Check the visibility of passed status to requesting user.
			ok, err := p.visFilter.StatusPublicTimelineable(ctx, requester, s)
			if err != nil {
				log.Errorf(ctx, "error checking status %s visibility: %v", s.URI, err)
			} else if !ok {
				return true
			}

			// Check if status been muted by requester from timelines.
			muted, err := p.muteFilter.StatusMuted(ctx, requester, s)
			if err != nil {
				log.Errorf(ctx, "error checking status %s mutes: %v", s.URI, err)
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
