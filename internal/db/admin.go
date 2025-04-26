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

// Admin contains functions related to instance administration (new signups etc).
type Admin interface {
	// IsUsernameAvailable checks whether a given username is available on our domain.
	// Returns an error if the username is already taken, or something went wrong in the db.
	IsUsernameAvailable(ctx context.Context, username string) (bool, error)

	// IsEmailAvailable checks whether a given email address for a new account is available to be used on our domain.
	// Return an error if:
	// A) the email is already associated with an account
	// B) we block signups from this email domain
	// C) something went wrong in the db
	IsEmailAvailable(ctx context.Context, email string) (bool, error)

	// NewSignup creates a new user + account in the database with the given parameters.
	// By the time this function is called, it should be assumed that all the parameters have passed validation!
	NewSignup(ctx context.Context, newSignup gtsmodel.NewSignup) (*gtsmodel.User, error)

	// CreateInstanceAccount creates an account in the database with the same username as the instance host value.
	// Ie., if the instance is hosted at 'example.org' the instance user will have a username of 'example.org'.
	// This is needed for things like serving files that belong to the instance and not an individual user/account.
	CreateInstanceAccount(ctx context.Context) error

	// CreateInstanceInstance creates an instance in the database with the same domain as the instance host value.
	// Ie., if the instance is hosted at 'example.org' the instance will have a domain of 'example.org'.
	// This is needed for things like serving instance information through /api/v1/instance
	CreateInstanceInstance(ctx context.Context) error

	// CreateInstanceApplication creates an application in the database
	// for use in processing signups etc through the sign-up form.
	CreateInstanceApplication(ctx context.Context) error

	// GetInstanceApplication gets the instance application
	// (ie., the application owned by the instance account).
	GetInstanceApplication(ctx context.Context) (*gtsmodel.Application, error)

	// CountApprovedSignupsSince counts the number of new account
	// sign-ups approved on this instance since the given time.
	CountApprovedSignupsSince(ctx context.Context, since time.Time) (int, error)

	// CountUnhandledSignups counts the number of account sign-ups
	// that have not yet been approved or denied. In other words,
	// the number of pending sign-ups sitting in the backlog.
	CountUnhandledSignups(ctx context.Context) (int, error)

	/*
		ACTION FUNCS
	*/

	// GetAdminAction returns the admin action with the given ID.
	GetAdminAction(ctx context.Context, id string) (*gtsmodel.AdminAction, error)

	// GetAdminActions gets all admin actions from the database.
	GetAdminActions(ctx context.Context) ([]*gtsmodel.AdminAction, error)

	// PutAdminAction puts one admin action in the database.
	PutAdminAction(ctx context.Context, action *gtsmodel.AdminAction) error

	// UpdateAdminAction updates one admin action by its ID.
	UpdateAdminAction(ctx context.Context, action *gtsmodel.AdminAction, columns ...string) error

	// DeleteAdminAction deletes admin action with the given ID.
	DeleteAdminAction(ctx context.Context, id string) error
}
