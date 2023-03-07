/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package db

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type StatusBookmark interface {
	// GetStatusBookmark gets one status bookmark with the given ID.
	GetStatusBookmark(ctx context.Context, id string) (*gtsmodel.StatusBookmark, Error)

	// GetStatusBookmarkByAccountID gets one status bookmark created by the given
	// accountID, targeting the given statusID.
	GetStatusBookmarkByAccountID(ctx context.Context, accountID string, statusID string) (*gtsmodel.StatusBookmark, Error)

	// PutStatusBookmark inserts the given statusBookmark into the database.
	PutStatusBookmark(ctx context.Context, statusBookmark *gtsmodel.StatusBookmark) Error

	// DeleteStatusBookmarks mass deletes status bookmarks targeting targetAccountID
	// and/or originating from originAccountID and/or bookmarking statusID.
	//
	// To delete all bookmarks of statusID from all accounts, just set statusID.
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
	// At least one parameter out of the three id params must not be an empty string.
	DeleteStatusBookmarks(ctx context.Context, targetAccountID string, originAccountID string, statusID string) Error
}
