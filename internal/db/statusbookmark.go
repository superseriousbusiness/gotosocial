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

package db

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type StatusBookmark interface {
	// GetStatusBookmark gets one status bookmark with the given ID.
	GetStatusBookmark(ctx context.Context, id string) (*gtsmodel.StatusBookmark, Error)

	// GetStatusBookmarkID is a shortcut function for returning just the database ID
	// of a status bookmark created by the given accountID, targeting the given statusID.
	GetStatusBookmarkID(ctx context.Context, accountID string, statusID string) (string, Error)

	// GetStatusBookmarks retrieves status bookmarks created by the given accountID,
	// and using the provided parameters. If limit is < 0 then no limit will be set.
	//
	// This function is primarily useful for paging through bookmarks in a sort of
	// timeline view.
	GetStatusBookmarks(ctx context.Context, accountID string, limit int, maxID string, minID string) ([]*gtsmodel.StatusBookmark, Error)

	// PutStatusBookmark inserts the given statusBookmark into the database.
	PutStatusBookmark(ctx context.Context, statusBookmark *gtsmodel.StatusBookmark) Error

	// DeleteStatusBookmark deletes one status bookmark with the given ID.
	DeleteStatusBookmark(ctx context.Context, id string) Error

	// DeleteStatusBookmarks mass deletes status bookmarks targeting targetAccountID
	// and/or originating from originAccountID and/or bookmarking statusID.
	//
	// If targetAccountID is set and originAccountID isn't, all status bookmarks
	// that target the given account will be deleted.
	//
	// If originAccountID is set and targetAccountID isn't, all status bookmarks
	// originating from the given account will be deleted.
	//
	// If both are set, then status bookmarks that target targetAccountID and
	// originate from originAccountID will be deleted.
	//
	// At least one parameter must not be an empty string.
	DeleteStatusBookmarks(ctx context.Context, targetAccountID string, originAccountID string) Error

	// DeleteStatusBookmarksForStatus deletes all status bookmarks that target the
	// given status ID. This is useful when a status has been deleted, and you need
	// to clean up after it.
	DeleteStatusBookmarksForStatus(ctx context.Context, statusID string) Error
}
