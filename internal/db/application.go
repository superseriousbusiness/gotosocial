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

type Application interface {
	// GetApplicationByID fetches the application from the database with corresponding ID value.
	GetApplicationByID(ctx context.Context, id string) (*gtsmodel.Application, error)

	// GetApplicationByClientID fetches the application from the database with corresponding client_id value.
	GetApplicationByClientID(ctx context.Context, clientID string) (*gtsmodel.Application, error)

	// GetApplicationsManagedByUserID fetches a page of applications managed by the given userID.
	GetApplicationsManagedByUserID(ctx context.Context, userID string, page *paging.Page) ([]*gtsmodel.Application, error)

	// PutApplication places the new application in the database, erroring on non-unique ID or client_id.
	PutApplication(ctx context.Context, app *gtsmodel.Application) error

	// DeleteApplicationByID deletes the application with corresponding id from the database.
	DeleteApplicationByID(ctx context.Context, id string) error

	// GetAllTokens fetches all client oauth tokens from database.
	GetAllTokens(ctx context.Context) ([]*gtsmodel.Token, error)

	// GetAccessTokens allows paging through a user's access (ie., user-level) tokens.
	GetAccessTokens(ctx context.Context, userID string, page *paging.Page) ([]*gtsmodel.Token, error)

	// GetTokenByID fetches the client oauth token from database with ID.
	GetTokenByID(ctx context.Context, id string) (*gtsmodel.Token, error)

	// GetTokenByCode fetches the client oauth token from database with code.
	GetTokenByCode(ctx context.Context, code string) (*gtsmodel.Token, error)

	// GetTokenByAccess fetches the client oauth token from database with access code.
	GetTokenByAccess(ctx context.Context, access string) (*gtsmodel.Token, error)

	// GetTokenByRefresh fetches the client oauth token from database with refresh code.
	GetTokenByRefresh(ctx context.Context, refresh string) (*gtsmodel.Token, error)

	// PutToken puts given client oauth token in the database.
	PutToken(ctx context.Context, token *gtsmodel.Token) error

	// UpdateToken updates the given token. Update all columns if no specific columns given.
	UpdateToken(ctx context.Context, token *gtsmodel.Token, columns ...string) error

	// DeleteTokenByID deletes client oauth token from database with ID.
	DeleteTokenByID(ctx context.Context, id string) error

	// DeleteTokenByCode deletes client oauth token from database with code.
	DeleteTokenByCode(ctx context.Context, code string) error

	// DeleteTokenByAccess deletes client oauth token from database with access code.
	DeleteTokenByAccess(ctx context.Context, access string) error

	// DeleteTokenByRefresh deletes client oauth token from database with refresh code.
	DeleteTokenByRefresh(ctx context.Context, refresh string) error

	// DeleteTokensByClientID deletes all tokens
	// with the given clientID from the database.
	DeleteTokensByClientID(ctx context.Context, clientID string) error
}
