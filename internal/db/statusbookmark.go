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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

type StatusBookmark interface {
	// GetStatusBookmark gets one status bookmark with the given ID.
	GetStatusBookmarkByID(ctx context.Context, id string) (*gtsmodel.StatusBookmark, error)

	// GetStatusBookmark fetches a status bookmark by the given account ID on the given status ID, if it exists.
	GetStatusBookmark(ctx context.Context, accountID string, statusID string) (*gtsmodel.StatusBookmark, error)

	// IsStatusBookmarked returns whether status has been bookmarked by any account.
	IsStatusBookmarked(ctx context.Context, statusID string) (bool, error)

	// IsStatusBookmarkedBy returns whether status ID is bookmarked by the given account ID.
	IsStatusBookmarkedBy(ctx context.Context, accountID string, statusID string) (bool, error)

	// GetStatusBookmarks retrieves status bookmarks created by the given accountID,
	// and using the provided parameters. If limit is < 0 then no limit will be set.
	//
	// This function is primarily useful for paging through bookmarks in a sort of
	// timeline view.
	GetStatusBookmarks(ctx context.Context, accountID string, limit int, maxID string, minID string) ([]*gtsmodel.StatusBookmark, error)

	// PutStatusBookmark inserts the given statusBookmark into the database.
	PutStatusBookmark(ctx context.Context, statusBookmark *gtsmodel.StatusBookmark) error

	// DeleteStatusBookmark deletes one status bookmark with the given ID.
	DeleteStatusBookmarkByID(ctx context.Context, id string) error

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
	DeleteStatusBookmarks(ctx context.Context, targetAccountID string, originAccountID string) error

	// DeleteStatusBookmarksForStatus deletes all status bookmarks that target the
	// given status ID. This is useful when a status has been deleted, and you need
	// to clean up after it.
	DeleteStatusBookmarksForStatus(ctx context.Context, statusID string) error
}
