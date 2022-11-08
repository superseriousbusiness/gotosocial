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

// Tombstone contains functionality for storing + retrieving tombstones for remote AP Activities + Objects.
type Tombstone interface {
	// GetTombstoneByURI attempts to fetch a tombstone by the given URI.
	GetTombstoneByURI(ctx context.Context, uri string) (*gtsmodel.Tombstone, Error)

	// TombstoneExistsWithURI returns true if a tombstone with the given URI exists.
	TombstoneExistsWithURI(ctx context.Context, uri string) (bool, Error)

	// PutTombstone creates a new tombstone in the database.
	PutTombstone(ctx context.Context, tombstone *gtsmodel.Tombstone) Error

	// DeleteTombstone deletes a tombstone with the given ID.
	DeleteTombstone(ctx context.Context, id string) Error
}
