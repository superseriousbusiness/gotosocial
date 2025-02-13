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

	// Fetch status timeline for list.
	return p.getStatusTimeline(ctx,

		// Auth'd
		// account.
		requester,

		// Per-account home timeline cache.
		p.state.Caches.Timelines.List.MustGet(requester.ID),

		// Current
		// page.
		page,

		// List timeline ID's endpoint.
		"/api/v1/timelines/list/"+listID,

		// No page
		// query.
		nil,

		// Status filter context.
		statusfilter.FilterContextHome,

		// Database load function.
		func(pg *paging.Page) (statuses []*gtsmodel.Status, err error) {
			return p.state.DB.GetListTimeline(ctx, requester.ID, pg)
		},

		// Pre-filtering function,
		// i.e. filter before caching.
		func(s *gtsmodel.Status) (bool, error) {

			// Check the visibility of passed status to requesting user.
			ok, err := p.visFilter.StatusHomeTimelineable(ctx, requester, s)
			return !ok, err
		},

		// Post-filtering function,
		// i.e. filter after caching.
		nil,
	)
}
