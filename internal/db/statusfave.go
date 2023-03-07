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

type StatusFave interface {
	// GetStatusFave returns one status fave with the given id.
	GetStatusFave(ctx context.Context, id string) (*gtsmodel.StatusFave, Error)

	// GetStatusFaveByAccountID gets one status fave created by the given
	// accountID, targeting the given statusID.
	GetStatusFaveByAccountID(ctx context.Context, accountID string, statusID string) (*gtsmodel.StatusFave, Error)

	// GetStatusFaves returns a slice of faves/likes of the given status.
	// This slice will be unfiltered, not taking account of blocks and whatnot, so filter it before serving it back to a user.
	GetStatusFaves(ctx context.Context, statusID string) ([]*gtsmodel.StatusFave, Error)

	// PutStatusFave inserts the given statusFave into the database.
	PutStatusFave(ctx context.Context, statusFave *gtsmodel.StatusFave) Error

	// DeleteStatusFaves mass deletes status faves targeting targetAccountID
	// and/or originating from originAccountID and/or faving statusID.
	//
	// To delete all faves of statusID from all accounts, just set statusID.
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
	// At least one parameter out of the three id params must not be an empty string.
	DeleteStatusFaves(ctx context.Context, targetAccountID string, originAccountID string, statusID string) Error
}
