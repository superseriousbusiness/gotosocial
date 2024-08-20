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
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

type Interaction interface {
	// GetInteractionApprovalByID gets one approval with the given id.
	GetInteractionApprovalByID(ctx context.Context, id string) (*gtsmodel.InteractionApproval, error)

	// GetInteractionApprovalByURI gets one approval with the given Accept uri.
	GetInteractionApprovalByURI(ctx context.Context, uri string) (*gtsmodel.InteractionApproval, error)

	// PopulateInteractionApproval ensures that the approval's struct fields are populated.
	PopulateInteractionApproval(ctx context.Context, approval *gtsmodel.InteractionApproval) error

	// PutInteractionApproval puts a new approval in the database.
	PutInteractionApproval(ctx context.Context, approval *gtsmodel.InteractionApproval) error

	// DeleteInteractionApprovalByID deletes one approval with the given ID.
	DeleteInteractionApprovalByID(ctx context.Context, id string) error

	// GetInteractionRejectionByID gets one rejection with the given id.
	GetInteractionRejectionByID(ctx context.Context, id string) (*gtsmodel.InteractionRejection, error)

	// GetInteractionRejectionByURI gets one rejection with the given Reject uri.
	GetInteractionRejectionByURI(ctx context.Context, uri string) (*gtsmodel.InteractionRejection, error)

	// InteractionRejected returns true if an rejection exists in the database for an
	// object with the given interactionURI (ie., a status or announce or fave uri).
	InteractionRejected(ctx context.Context, interactionURI string) (bool, error)

	// PopulateInteractionRejection ensures that the rejection's struct fields are populated.
	PopulateInteractionRejection(ctx context.Context, rejection *gtsmodel.InteractionRejection) error

	// PutInteractionRejection puts a new rejection in the database.
	PutInteractionRejection(ctx context.Context, rejection *gtsmodel.InteractionRejection) error

	// DeleteInteractionRejectionByID deletes one rejection with the given ID.
	DeleteInteractionRejectionByID(ctx context.Context, id string) error

	// GetInteractionRequestByID gets one request with the given id.
	GetInteractionRequestByID(ctx context.Context, id string) (*gtsmodel.InteractionRequest, error)

	// GetInteractionRequestByID gets one request with the given interaction uri.
	GetInteractionRequestByInteractionURI(ctx context.Context, uri string) (*gtsmodel.InteractionRequest, error)

	// PopulateInteractionRequest ensures that the request's struct fields are populated.
	PopulateInteractionRequest(ctx context.Context, request *gtsmodel.InteractionRequest) error

	// PutInteractionRequest puts a new request in the database.
	PutInteractionRequest(ctx context.Context, request *gtsmodel.InteractionRequest) error

	// DeleteInteractionRequestByID deletes one request with the given ID.
	DeleteInteractionRequestByID(ctx context.Context, id string) error

	// GetInteractionsRequestsForAcct returns pending interactions targeting
	// the given (optional) account ID and the given (optional) status ID.
	//
	// At least one of `likes`, `replies`, or `boosts` must be true.
	GetInteractionsRequestsForAcct(
		ctx context.Context,
		acctID string,
		statusID string,
		likes bool,
		replies bool,
		boosts bool,
		page *paging.Page,
	) ([]*gtsmodel.InteractionRequest, error)
}
