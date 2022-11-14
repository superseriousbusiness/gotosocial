/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

// Status contains functions for getting statuses, creating statuses, and checking various other fields on statuses.
type Status interface {
	// GetStatusByID returns one status from the database, with no rel fields populated, only their linking ID / URIs
	GetStatusByID(ctx context.Context, id string) (*gtsmodel.Status, Error)

	// GetStatusByURI returns one status from the database, with no rel fields populated, only their linking ID / URIs
	GetStatusByURI(ctx context.Context, uri string) (*gtsmodel.Status, Error)

	// GetStatusByURL returns one status from the database, with no rel fields populated, only their linking ID / URIs
	GetStatusByURL(ctx context.Context, uri string) (*gtsmodel.Status, Error)

	// PutStatus stores one status in the database.
	PutStatus(ctx context.Context, status *gtsmodel.Status) Error

	// UpdateStatus updates one status in the database and returns it to the caller.
	UpdateStatus(ctx context.Context, status *gtsmodel.Status) Error

	// DeleteStatusByID deletes one status from the database.
	DeleteStatusByID(ctx context.Context, id string) Error

	// CountStatusReplies returns the amount of replies recorded for a status, or an error if something goes wrong
	CountStatusReplies(ctx context.Context, status *gtsmodel.Status) (int, Error)

	// CountStatusReblogs returns the amount of reblogs/boosts recorded for a status, or an error if something goes wrong
	CountStatusReblogs(ctx context.Context, status *gtsmodel.Status) (int, Error)

	// CountStatusFaves returns the amount of faves/likes recorded for a status, or an error if something goes wrong
	CountStatusFaves(ctx context.Context, status *gtsmodel.Status) (int, Error)

	// GetStatusParents gets the parent statuses of a given status.
	//
	// If onlyDirect is true, only the immediate parent will be returned.
	GetStatusParents(ctx context.Context, status *gtsmodel.Status, onlyDirect bool) ([]*gtsmodel.Status, Error)

	// GetStatusChildren gets the child statuses of a given status.
	//
	// If onlyDirect is true, only the immediate children will be returned.
	GetStatusChildren(ctx context.Context, status *gtsmodel.Status, onlyDirect bool, minID string) ([]*gtsmodel.Status, Error)

	// IsStatusFavedBy checks if a given status has been faved by a given account ID
	IsStatusFavedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, Error)

	// IsStatusRebloggedBy checks if a given status has been reblogged/boosted by a given account ID
	IsStatusRebloggedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, Error)

	// IsStatusMutedBy checks if a given status has been muted by a given account ID
	IsStatusMutedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, Error)

	// IsStatusBookmarkedBy checks if a given status has been bookmarked by a given account ID
	IsStatusBookmarkedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, Error)

	// GetStatusFaves returns a slice of faves/likes of the given status.
	// This slice will be unfiltered, not taking account of blocks and whatnot, so filter it before serving it back to a user.
	GetStatusFaves(ctx context.Context, status *gtsmodel.Status) ([]*gtsmodel.StatusFave, Error)

	// GetStatusReblogs returns a slice of statuses that are a boost/reblog of the given status.
	// This slice will be unfiltered, not taking account of blocks and whatnot, so filter it before serving it back to a user.
	GetStatusReblogs(ctx context.Context, status *gtsmodel.Status) ([]*gtsmodel.Status, Error)
}
