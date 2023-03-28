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

package account

import (
	"context"
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// BookmarksGet returns a pageable response of statuses that are bookmarked by requestingAccount.
// Paging for this response is done based on bookmark ID rather than status ID.
func (p *Processor) BookmarksGet(ctx context.Context, requestingAccount *gtsmodel.Account, limit int, maxID string, minID string) (*apimodel.PageableResponse, gtserror.WithCode) {
	bookmarks, err := p.state.DB.GetStatusBookmarks(ctx, requestingAccount.ID, limit, maxID, minID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	var (
		count          = len(bookmarks)
		items          = make([]interface{}, 0, count)
		nextMaxIDValue = id.Highest
		prevMinIDValue = id.Lowest
	)

	for _, bookmark := range bookmarks {
		status, err := p.state.DB.GetStatusByID(ctx, bookmark.StatusID)
		if err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				// We just don't have the status for some reason.
				// Skip this one.
				continue
			}
			return nil, gtserror.NewErrorInternalError(err) // A real error has occurred.
		}

		visible, err := p.filter.StatusVisible(ctx, requestingAccount, status)
		if err != nil {
			log.Errorf(ctx, "error checking bookmarked status visibility: %s", err)
			continue
		}

		if !visible {
			continue
		}

		// Convert the status.
		item, err := p.tc.StatusToAPIStatus(ctx, status, requestingAccount)
		if err != nil {
			log.Errorf(ctx, "error converting bookmarked status to api: %s", err)
			continue
		}
		items = append(items, item)

		// Page based on bookmark ID, not status ID.
		// Note that we only set these values here
		// when we're certain that the caller is able
		// to see the status, *and* we're sure that
		// we can produce an api model representation.
		if bookmark.ID < nextMaxIDValue {
			nextMaxIDValue = bookmark.ID // Lowest ID (for paging down).
		}
		if bookmark.ID > prevMinIDValue {
			prevMinIDValue = bookmark.ID // Highest ID (for paging up).
		}
	}

	if len(items) == 0 {
		return util.EmptyPageableResponse(), nil
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "/api/v1/bookmarks",
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}
