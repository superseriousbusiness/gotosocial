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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

type Move interface {
	// GetMoveByID gets one Move with the given internal ID.
	GetMoveByID(ctx context.Context, id string) (*gtsmodel.Move, error)

	// GetMoveByURI gets one Move with the given AP URI.
	GetMoveByURI(ctx context.Context, uri string) (*gtsmodel.Move, error)

	// GetMoveByOriginTarget gets one move with the given originURI and targetURI.
	GetMoveByOriginTarget(ctx context.Context, originURI string, targetURI string) (*gtsmodel.Move, error)

	// PopulateMove parses out the origin and target URIs on the move.
	PopulateMove(ctx context.Context, move *gtsmodel.Move) error

	// GetLatestMoveSuccessInvolvingURIs gets the time of
	// the latest successfully-processed Move that includes
	// either uri1 or uri2 in target or origin positions.
	GetLatestMoveSuccessInvolvingURIs(ctx context.Context, uri1 string, uri2 string) (time.Time, error)

	// GetLatestMoveAttemptInvolvingURIs gets the time
	// of the latest Move attempt that includes either
	// uri1 or uri2 in target or origin positions.
	GetLatestMoveAttemptInvolvingURIs(ctx context.Context, uri1 string, uri2 string) (time.Time, error)

	// PutMove puts the given Move in the database.
	PutMove(ctx context.Context, move *gtsmodel.Move) error

	// UpdateMove updates the given Move by primary key.
	// Updates specific columns if provided, all columns if not.
	UpdateMove(ctx context.Context, move *gtsmodel.Move, columns ...string) error

	// DeleteMoveByID deletes a move with the given internal ID.
	DeleteMoveByID(ctx context.Context, id string) error
}
