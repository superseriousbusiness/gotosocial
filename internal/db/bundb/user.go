/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package bundb

import (
	"context"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type userDB struct {
	conn  *DBConn
	cache *cache.UserCache
}

func (u *userDB) newUserQ(user *gtsmodel.User) *bun.SelectQuery {
	return u.conn.
		NewSelect().
		Model(user).
		Relation("Account")
}

func (u *userDB) getUser(ctx context.Context, cacheGet func() (*gtsmodel.User, bool), dbQuery func(*gtsmodel.User) error) (*gtsmodel.User, db.Error) {
	// Attempt to fetch cached user
	user, cached := cacheGet()

	if !cached {
		user = &gtsmodel.User{}

		// Not cached! Perform database query
		err := dbQuery(user)
		if err != nil {
			return nil, u.conn.ProcessError(err)
		}

		// Place in the cache
		u.cache.Put(user)
	}

	return user, nil
}

func (u *userDB) GetUserByID(ctx context.Context, id string) (*gtsmodel.User, db.Error) {
	return u.getUser(
		ctx,
		func() (*gtsmodel.User, bool) {
			return u.cache.GetByID(id)
		},
		func(user *gtsmodel.User) error {
			return u.newUserQ(user).Where("? = ?", bun.Ident("user.id"), id).Scan(ctx)
		},
	)
}

func (u *userDB) GetUserByAccountID(ctx context.Context, accountID string) (*gtsmodel.User, db.Error) {
	return u.getUser(
		ctx,
		func() (*gtsmodel.User, bool) {
			return u.cache.GetByAccountID(accountID)
		},
		func(user *gtsmodel.User) error {
			return u.newUserQ(user).Where("? = ?", bun.Ident("user.account_id"), accountID).Scan(ctx)
		},
	)
}

func (u *userDB) GetUserByEmailAddress(ctx context.Context, emailAddress string) (*gtsmodel.User, db.Error) {
	return u.getUser(
		ctx,
		func() (*gtsmodel.User, bool) {
			return u.cache.GetByEmail(emailAddress)
		},
		func(user *gtsmodel.User) error {
			return u.newUserQ(user).Where("? = ?", bun.Ident("user.email"), emailAddress).Scan(ctx)
		},
	)
}

func (u *userDB) GetUserByConfirmationToken(ctx context.Context, confirmationToken string) (*gtsmodel.User, db.Error) {
	return u.getUser(
		ctx,
		func() (*gtsmodel.User, bool) {
			return u.cache.GetByConfirmationToken(confirmationToken)
		},
		func(user *gtsmodel.User) error {
			return u.newUserQ(user).Where("? = ?", bun.Ident("user.confirmation_token"), confirmationToken).Scan(ctx)
		},
	)
}

func (u *userDB) PutUser(ctx context.Context, user *gtsmodel.User) (*gtsmodel.User, db.Error) {
	if _, err := u.conn.
		NewInsert().
		Model(user).
		Exec(ctx); err != nil {
		return nil, u.conn.ProcessError(err)
	}

	u.cache.Put(user)
	return user, nil
}

func (u *userDB) UpdateUser(ctx context.Context, user *gtsmodel.User, columns ...string) (*gtsmodel.User, db.Error) {
	// Update the user's last-updated
	user.UpdatedAt = time.Now()

	if _, err := u.conn.
		NewUpdate().
		Model(user).
		WherePK().
		Column(columns...).
		Exec(ctx); err != nil {
		return nil, u.conn.ProcessError(err)
	}

	u.cache.Invalidate(user.ID)
	return user, nil
}

func (u *userDB) DeleteUserByID(ctx context.Context, userID string) db.Error {
	if _, err := u.conn.
		NewDelete().
		Model(&gtsmodel.User{ID: userID}).
		WherePK().
		Exec(ctx); err != nil {
		return u.conn.ProcessError(err)
	}

	u.cache.Invalidate(userID)
	return nil
}
