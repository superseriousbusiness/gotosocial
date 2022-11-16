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

	"codeberg.org/gruf/go-cache/v3/result"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type userDB struct {
	conn  *DBConn
	cache *result.Cache[*gtsmodel.User]
}

func (u *userDB) init() {
	// Initialize user result cache
	u.cache = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "AccountID"},
		{Name: "Email"},
		{Name: "ConfirmationToken"},
	}, func(u1 *gtsmodel.User) *gtsmodel.User {
		u2 := new(gtsmodel.User)
		*u2 = *u1
		return u2
	}, 1000)

	// Set cache TTL and start sweep routine
	u.cache.SetTTL(time.Minute*5, false)
	u.cache.Start(time.Second * 10)
}

func (u *userDB) GetUserByID(ctx context.Context, id string) (*gtsmodel.User, db.Error) {
	return u.cache.Load("ID", func() (*gtsmodel.User, error) {
		var user gtsmodel.User

		q := u.conn.
			NewSelect().
			Model(&user).
			Relation("Account").
			Where("? = ?", bun.Ident("user.id"), id)

		if err := q.Scan(ctx); err != nil {
			return nil, u.conn.ProcessError(err)
		}

		return &user, nil
	}, id)
}

func (u *userDB) GetUserByAccountID(ctx context.Context, accountID string) (*gtsmodel.User, db.Error) {
	return u.cache.Load("AccountID", func() (*gtsmodel.User, error) {
		var user gtsmodel.User

		q := u.conn.
			NewSelect().
			Model(&user).
			Relation("Account").
			Where("? = ?", bun.Ident("user.account_id"), accountID)

		if err := q.Scan(ctx); err != nil {
			return nil, u.conn.ProcessError(err)
		}

		return &user, nil
	}, accountID)
}

func (u *userDB) GetUserByEmailAddress(ctx context.Context, emailAddress string) (*gtsmodel.User, db.Error) {
	return u.cache.Load("Email", func() (*gtsmodel.User, error) {
		var user gtsmodel.User

		q := u.conn.
			NewSelect().
			Model(&user).
			Relation("Account").
			Where("? = ?", bun.Ident("user.email"), emailAddress)

		if err := q.Scan(ctx); err != nil {
			return nil, u.conn.ProcessError(err)
		}

		return &user, nil
	}, emailAddress)
}

func (u *userDB) GetUserByConfirmationToken(ctx context.Context, confirmationToken string) (*gtsmodel.User, db.Error) {
	return u.cache.Load("ConfirmationToken", func() (*gtsmodel.User, error) {
		var user gtsmodel.User

		q := u.conn.
			NewSelect().
			Model(&user).
			Relation("Account").
			Where("? = ?", bun.Ident("user.confirmation_token"), confirmationToken)

		if err := q.Scan(ctx); err != nil {
			return nil, u.conn.ProcessError(err)
		}

		return &user, nil
	}, confirmationToken)
}

func (u *userDB) PutUser(ctx context.Context, user *gtsmodel.User) db.Error {
	return u.cache.Store(user, func() error {
		_, err := u.conn.
			NewInsert().
			Model(user).
			Exec(ctx)
		return u.conn.ProcessError(err)
	})
}

func (u *userDB) UpdateUser(ctx context.Context, user *gtsmodel.User, columns ...string) db.Error {
	// Update the user's last-updated
	user.UpdatedAt = time.Now()

	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included
		columns = append(columns, "updated_at")
	}

	// Update the user in DB
	_, err := u.conn.
		NewUpdate().
		Model(user).
		Where("? = ?", bun.Ident("user.id"), user.ID).
		Column(columns...).
		Exec(ctx)
	if err != nil {
		return u.conn.ProcessError(err)
	}

	// Invalidate in cache
	u.cache.Invalidate("ID", user.ID)
	return nil
}

func (u *userDB) DeleteUserByID(ctx context.Context, userID string) db.Error {
	if _, err := u.conn.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("users"), bun.Ident("user")).
		Where("? = ?", bun.Ident("user.id"), userID).
		Exec(ctx); err != nil {
		return u.conn.ProcessError(err)
	}

	// Invalidate user from cache
	u.cache.Invalidate("ID", userID)
	return nil
}
