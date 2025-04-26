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

package bundb

import (
	"context"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type userDB struct {
	db    *bun.DB
	state *state.State
}

func (u *userDB) GetUserByID(ctx context.Context, id string) (*gtsmodel.User, error) {
	return u.getUser(
		ctx,
		"ID",
		func(user *gtsmodel.User) error {
			return u.db.NewSelect().Model(user).Where("? = ?", bun.Ident("id"), id).Scan(ctx)
		},
		id,
	)
}

func (u *userDB) GetUsersByIDs(ctx context.Context, ids []string) ([]*gtsmodel.User, error) {
	var (
		users = make([]*gtsmodel.User, 0, len(ids))

		// Collect errors instead of
		// returning early on any.
		errs gtserror.MultiError
	)

	for _, id := range ids {
		// Attempt to fetch user from DB.
		user, err := u.GetUserByID(ctx, id)
		if err != nil {
			errs.Appendf("error getting user %s: %w", id, err)
			continue
		}

		// Append user to return slice.
		users = append(users, user)
	}

	return users, errs.Combine()
}

func (u *userDB) GetUserByAccountID(ctx context.Context, accountID string) (*gtsmodel.User, error) {
	return u.getUser(
		ctx,
		"AccountID",
		func(user *gtsmodel.User) error {
			return u.db.NewSelect().Model(user).Where("? = ?", bun.Ident("account_id"), accountID).Scan(ctx)
		},
		accountID,
	)
}

func (u *userDB) GetUserByEmailAddress(ctx context.Context, email string) (*gtsmodel.User, error) {
	return u.getUser(
		ctx,
		"Email",
		func(user *gtsmodel.User) error {
			return u.db.NewSelect().Model(user).Where("? = ?", bun.Ident("email"), email).Scan(ctx)
		},
		email,
	)
}

func (u *userDB) GetUserByExternalID(ctx context.Context, id string) (*gtsmodel.User, error) {
	return u.getUser(
		ctx,
		"ExternalID",
		func(user *gtsmodel.User) error {
			return u.db.NewSelect().Model(user).Where("? = ?", bun.Ident("external_id"), id).Scan(ctx)
		},
		id,
	)
}

func (u *userDB) GetUserByConfirmationToken(ctx context.Context, token string) (*gtsmodel.User, error) {
	return u.getUser(
		ctx,
		"ConfirmationToken",
		func(user *gtsmodel.User) error {
			return u.db.NewSelect().Model(user).Where("? = ?", bun.Ident("confirmation_token"), token).Scan(ctx)
		},
		token,
	)
}

func (u *userDB) getUser(ctx context.Context, lookup string, dbQuery func(*gtsmodel.User) error, keyParts ...any) (*gtsmodel.User, error) {
	// Fetch user from database cache with loader callback.
	user, err := u.state.Caches.DB.User.LoadOne(lookup, func() (*gtsmodel.User, error) {
		var user gtsmodel.User

		// Not cached! perform database query.
		if err := dbQuery(&user); err != nil {
			return nil, err
		}

		return &user, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Return without populating.
		return user, nil
	}

	if err := u.PopulateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// PopulateUser ensures that the user's struct fields are populated.
func (u *userDB) PopulateUser(ctx context.Context, user *gtsmodel.User) error {
	var (
		errs = gtserror.NewMultiError(1)
		err  error
	)

	if user.Account == nil {
		// Fetch the related account model for this user.
		user.Account, err = u.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			user.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating user account: %w", err)
		}
	}

	return errs.Combine()
}

func (u *userDB) GetAllUsers(ctx context.Context) ([]*gtsmodel.User, error) {
	var userIDs []string

	// Scan all user IDs into slice.
	if err := u.db.NewSelect().
		Table("users").
		Column("id").
		Scan(ctx, &userIDs); err != nil {
		return nil, err
	}

	// Transform user IDs into user slice.
	return u.GetUsersByIDs(ctx, userIDs)
}

func (u *userDB) PutUser(ctx context.Context, user *gtsmodel.User) error {
	return u.state.Caches.DB.User.Store(user, func() error {
		_, err := u.db.
			NewInsert().
			Model(user).
			Exec(ctx)
		return err
	})
}

func (u *userDB) UpdateUser(ctx context.Context, user *gtsmodel.User, columns ...string) error {
	// Update the user's last-updated
	user.UpdatedAt = time.Now()

	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included
		columns = append(columns, "updated_at")
	}

	return u.state.Caches.DB.User.Store(user, func() error {
		_, err := u.db.
			NewUpdate().
			Model(user).
			Where("? = ?", bun.Ident("user.id"), user.ID).
			Column(columns...).
			Exec(ctx)
		return err
	})
}

func (u *userDB) DeleteUserByID(ctx context.Context, userID string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted gtsmodel.User
	deleted.ID = userID

	// Delete user from DB.
	if _, err := u.db.NewDelete().
		Model(&deleted).
		Where("? = ?", bun.Ident("id"), userID).
		Returning("?", bun.Ident("account_id")).
		Exec(ctx); err != nil {
		return err
	}

	// Invalidate cached user by ID, manually
	// call invalidate hook in case not cached.
	u.state.Caches.DB.User.Invalidate("ID", userID)
	u.state.Caches.OnInvalidateUser(&deleted)

	return nil
}

func (u *userDB) PutDeniedUser(ctx context.Context, deniedUser *gtsmodel.DeniedUser) error {
	_, err := u.db.NewInsert().
		Model(deniedUser).
		Exec(ctx)
	return err
}

func (u *userDB) GetDeniedUserByID(ctx context.Context, id string) (*gtsmodel.DeniedUser, error) {
	deniedUser := new(gtsmodel.DeniedUser)
	if err := u.db.
		NewSelect().
		Model(deniedUser).
		Where("? = ?", bun.Ident("denied_user.id"), id).
		Scan(ctx); err != nil {
		return nil, err
	}

	return deniedUser, nil
}
