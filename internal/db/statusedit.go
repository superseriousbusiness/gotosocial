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

type StatusEdit interface {

	// GetStatusEditByID fetches the StatusEdit with given ID from the database.
	GetStatusEditByID(ctx context.Context, id string) (*gtsmodel.StatusEdit, error)

	// GetStatusEditsByIDs fetches all StatusEdits with given IDs from database,
	// this is optimized and faster than multiple calls to GetStatusEditByID.
	GetStatusEditsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.StatusEdit, error)

	// PopulateStatusEdit ensures the given StatusEdit's sub-models are populated.
	PopulateStatusEdit(ctx context.Context, edit *gtsmodel.StatusEdit) error

	// PutStatusEdit inserts the given new StatusEdit into the database.
	PutStatusEdit(ctx context.Context, edit *gtsmodel.StatusEdit) error

	// DeleteStatusEdits deletes the StatusEdits with given IDs from the database.
	DeleteStatusEdits(ctx context.Context, ids []string) error
}
