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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/text"
)

// TagTimelineGet gets a pageable timeline for the given
// tagName and given paging parameters. It will ensure
// that each status in the timeline is actually visible
// to requestingAcct before returning it.
func (p *Processor) TagTimelineGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	tagName string,
	maxID string,
	sinceID string,
	minID string,
	limit int,
) (*apimodel.PageableResponse, gtserror.WithCode) {

	// Fetch the requested tag with name.
	tag, errWithCode := p.getTag(ctx, tagName)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Check for a useable returned tag for endpoint.
	if tag == nil || !*tag.Useable || !*tag.Listable {

		// Obey mastodon API by returning 404 for this.
		const text = "tag was not found, or not useable/listable on this instance"
		return nil, gtserror.NewWithCode(http.StatusNotFound, text)
	}

	// Fetch status timeline for tag.
	return p.getStatusTimeline(ctx,

		// Auth'd
		// account.
		requester,

		// No
		// cache.
		nil,

		// Current
		// page.
		&paging.Page{
			Min:   paging.EitherMinID(minID, sinceID),
			Max:   paging.MaxID(maxID),
			Limit: limit,
		},

		// Tag timeline name's endpoint.
		"/api/v1/timelines/tag/"+tagName,

		// No page
		// query.
		nil,

		// Status filter context.
		gtsmodel.FilterContextPublic,

		// Database load function.
		func(pg *paging.Page) (statuses []*gtsmodel.Status, err error) {
			return p.state.DB.GetTagTimeline(ctx, tag.ID, pg)
		},

		// Filtering function,
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

func (p *Processor) getTag(ctx context.Context, tagName string) (*gtsmodel.Tag, gtserror.WithCode) {
	// Normalize + validate tag name.
	tagNameNormal, ok := text.NormalizeHashtag(tagName)
	if !ok {
		err := gtserror.Newf("string '%s' could not be normalized to a valid hashtag", tagName)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Ensure we have tag with this name in the db.
	tag, err := p.state.DB.GetTagByName(ctx, tagNameNormal)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Real db error.
		err = gtserror.Newf("db error getting tag by name: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return tag, nil
}
