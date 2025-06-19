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
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

type Interaction interface {
	// GetInteractionRequestByID gets one request with the given id.
	GetInteractionRequestByID(ctx context.Context, id string) (*gtsmodel.InteractionRequest, error)

	// GetInteractionRequestByID gets one request with the given interaction uri.
	GetInteractionRequestByInteractionURI(ctx context.Context, uri string) (*gtsmodel.InteractionRequest, error)

	// GetInteractionRequestByURI returns one accepted or rejected
	// interaction request with the given URI, if it exists in the db.
	GetInteractionRequestByURI(ctx context.Context, uri string) (*gtsmodel.InteractionRequest, error)

	// PopulateInteractionRequest ensures that the request's struct fields are populated.
	PopulateInteractionRequest(ctx context.Context, request *gtsmodel.InteractionRequest) error

	// PutInteractionRequest puts a new request in the database.
	PutInteractionRequest(ctx context.Context, request *gtsmodel.InteractionRequest) error

	// UpdateInteractionRequest updates the given interaction request.
	UpdateInteractionRequest(ctx context.Context, request *gtsmodel.InteractionRequest, columns ...string) error

	// DeleteInteractionRequestByID deletes one request with the given ID.
	DeleteInteractionRequestByID(ctx context.Context, id string) error

	// DeleteInteractionRequestsByInteractingAccountID deletes all requests
	// originating from the given account ID.
	DeleteInteractionRequestsByInteractingAccountID(ctx context.Context, accountID string) error

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

	// IsInteractionRejected returns true if an rejection exists in the database for an
	// object with the given interactionURI (ie., a status or announce or fave uri).
	IsInteractionRejected(ctx context.Context, interactionURI string) (bool, error)
}
