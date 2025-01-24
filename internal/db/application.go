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

type Application interface {
	// GetApplicationByID fetches the application from the database with corresponding ID value.
	GetApplicationByID(ctx context.Context, id string) (*gtsmodel.Application, error)

	// GetApplicationByClientID fetches the application from the database with corresponding client_id value.
	GetApplicationByClientID(ctx context.Context, clientID string) (*gtsmodel.Application, error)

	// PutApplication places the new application in the database, erroring on non-unique ID or client_id.
	PutApplication(ctx context.Context, app *gtsmodel.Application) error

	// DeleteApplicationByClientID deletes the application with corresponding client_id value from the database.
	DeleteApplicationByClientID(ctx context.Context, clientID string) error

	// GetClientByID ...
	GetClientByID(ctx context.Context, id string) (*gtsmodel.Client, error)

	// PutClient ...
	PutClient(ctx context.Context, client *gtsmodel.Client) error

	// DeleteClientByID ...
	DeleteClientByID(ctx context.Context, id string) error

	// GetAllTokens ...
	GetAllTokens(ctx context.Context) ([]*gtsmodel.Token, error)

	// GetTokenByID ...
	GetTokenByID(ctx context.Context, id string) (*gtsmodel.Token, error)

	// GetTokenByCode ...
	GetTokenByCode(ctx context.Context, code string) (*gtsmodel.Token, error)

	// GetTokenByAccess ...
	GetTokenByAccess(ctx context.Context, access string) (*gtsmodel.Token, error)

	// GetTokenByRefresh ...
	GetTokenByRefresh(ctx context.Context, refresh string) (*gtsmodel.Token, error)

	// PutToken ...
	PutToken(ctx context.Context, token *gtsmodel.Token) error

	// DeleteTokenByID ...
	DeleteTokenByID(ctx context.Context, id string) error

	// DeleteTokenByCode ...
	DeleteTokenByCode(ctx context.Context, code string) error

	// DeleteTokenByAccess ...
	DeleteTokenByAccess(ctx context.Context, access string) error

	// DeleteTokenByRefresh ...
	DeleteTokenByRefresh(ctx context.Context, refresh string) error
}
