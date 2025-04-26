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

type StatusFave interface {
	// GetStatusFaveByAccountID gets one status fave created by the given accountID, targeting the given statusID.
	GetStatusFave(ctx context.Context, accountID string, statusID string) (*gtsmodel.StatusFave, error)

	// GetStatusFaveByID returns one status fave with the given id.
	GetStatusFaveByID(ctx context.Context, id string) (*gtsmodel.StatusFave, error)

	// GetStatusFaveByURI returns one status fave with the given uri.
	GetStatusFaveByURI(ctx context.Context, uri string) (*gtsmodel.StatusFave, error)

	// GetStatusFaves returns a slice of faves/likes of the status with given ID.
	// This slice will be unfiltered, not taking account of blocks and whatnot, so filter it before serving it back to a user.
	GetStatusFaves(ctx context.Context, statusID string) ([]*gtsmodel.StatusFave, error)

	// PopulateStatusFave ensures that all sub-models of a fave are populated (account, status, etc).
	PopulateStatusFave(ctx context.Context, statusFave *gtsmodel.StatusFave) error

	// PutStatusFave inserts the given statusFave into the database.
	PutStatusFave(ctx context.Context, statusFave *gtsmodel.StatusFave) error

	// UpdateStatusFave updates one statusFave in the database.
	UpdateStatusFave(ctx context.Context, statusFave *gtsmodel.StatusFave, columns ...string) error

	// DeleteStatusFave deletes one status fave with the given id.
	DeleteStatusFaveByID(ctx context.Context, id string) error

	// DeleteStatusFaves mass deletes status faves targeting targetAccountID
	// and/or originating from originAccountID and/or faving statusID.
	//
	// If targetAccountID is set and originAccountID isn't, all status faves
	// that target the given account will be deleted.
	//
	// If originAccountID is set and targetAccountID isn't, all status faves
	// originating from the given account will be deleted.
	//
	// If both are set, then status faves that target targetAccountID and
	// originate from originAccountID will be deleted.
	//
	// At least one parameter must not be an empty string.
	DeleteStatusFaves(ctx context.Context, targetAccountID string, originAccountID string) error

	// DeleteStatusFavesForStatus deletes all status faves that target the given status ID.
	// This is useful when a status has been deleted, and you need to clean up after it.
	DeleteStatusFavesForStatus(ctx context.Context, statusID string) error

	// CountStatusFaves returns the number of status favourites registered for status with ID.
	CountStatusFaves(ctx context.Context, statusID string) (int, error)

	// IsStatusFavedBy returns whether the status with ID has been favourited by account with ID.
	IsStatusFavedBy(ctx context.Context, statusID string, accountID string) (bool, error)
}
