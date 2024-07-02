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

type Interaction interface {
	// GetInteractionApprovalByID gets one approval with the given id.
	GetInteractionApprovalByID(ctx context.Context, id string) (*gtsmodel.InteractionApproval, error)

	// GetInteractionApprovalByID gets one approval with the given uri.
	GetInteractionApprovalByURI(ctx context.Context, id string) (*gtsmodel.InteractionApproval, error)

	// PopulateInteractionApproval ensures that the approval's struct fields are populated.
	PopulateInteractionApproval(ctx context.Context, approval *gtsmodel.InteractionApproval) error

	// PutInteractionApproval puts a new approval in the database.
	PutInteractionApproval(ctx context.Context, approval *gtsmodel.InteractionApproval) error

	// DeleteInteractionApprovalByID deletes one approval with the given ID.
	DeleteInteractionApprovalByID(ctx context.Context, id string) error
}
