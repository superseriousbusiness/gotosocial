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

// Tombstone contains functionality for storing + retrieving tombstones for remote AP Activities + Objects.
type Tombstone interface {
	// GetTombstoneByURI attempts to fetch a tombstone by the given URI.
	GetTombstoneByURI(ctx context.Context, uri string) (*gtsmodel.Tombstone, error)

	// TombstoneExistsWithURI returns true if a tombstone with the given URI exists.
	TombstoneExistsWithURI(ctx context.Context, uri string) (bool, error)

	// PutTombstone creates a new tombstone in the database.
	PutTombstone(ctx context.Context, tombstone *gtsmodel.Tombstone) error

	// DeleteTombstone deletes a tombstone with the given ID.
	DeleteTombstone(ctx context.Context, id string) error
}
