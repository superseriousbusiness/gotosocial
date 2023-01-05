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

// User contains functions related to user getting/setting/creation.
type User interface {
	// GetUserByID returns one user with the given ID, or an error if something goes wrong.
	GetUserByID(ctx context.Context, id string) (*gtsmodel.User, Error)
	// GetUserByAccountID returns one user by its account ID, or an error if something goes wrong.
	GetUserByAccountID(ctx context.Context, accountID string) (*gtsmodel.User, Error)
	// GetUserByID returns one user with the given email address, or an error if something goes wrong.
	GetUserByEmailAddress(ctx context.Context, emailAddress string) (*gtsmodel.User, Error)
	// GetUserByExternalID returns one user with the given external id, or an error if something goes wrong.
	GetUserByExternalID(ctx context.Context, id string) (*gtsmodel.User, Error)
	// GetUserByConfirmationToken returns one user by its confirmation token, or an error if something goes wrong.
	GetUserByConfirmationToken(ctx context.Context, confirmationToken string) (*gtsmodel.User, Error)
	// PutUser will attempt to place user in the database
	PutUser(ctx context.Context, user *gtsmodel.User) Error
	// UpdateUser updates one user by its primary key, updating either only the specified columns, or all of them.
	UpdateUser(ctx context.Context, user *gtsmodel.User, columns ...string) Error
	// DeleteUserByID deletes one user by its ID.
	DeleteUserByID(ctx context.Context, userID string) Error
}
