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

package migrations

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// new account indexes
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Account{}).
				Index("accounts_url_idx").
				Column("url").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Account{}).
				Index("accounts_public_key_uri_idx").
				Column("public_key_uri").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Account{}).
				Index("accounts_suspended_at_idx").
				Column("suspended_at").
				Exec(ctx); err != nil {
				return err
			}

			// new user indexes
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.User{}).
				Index("users_account_id_idx").
				Column("account_id").
				Exec(ctx); err != nil {
				return err
			}

			// new tags indexes
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Tag{}).
				Index("tags_name_idx").
				Column("name").
				Exec(ctx); err != nil {
				return err
			}

			// new applications indexes
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Application{}).
				Index("applications_client_id_idx").
				Column("client_id").
				Exec(ctx); err != nil {
				return err
			}

			// new status_faves indexes
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.StatusFave{}).
				Index("status_faves_account_id_idx").
				Column("account_id").
				Exec(ctx); err != nil {
				return err
			}

			// new status_mutes indexes
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.StatusMute{}).
				Index("status_mutes_account_id_target_account_id_status_id_idx").
				Column("account_id", "target_account_id", "status_id").
				Exec(ctx); err != nil {
				return err
			}

			return nil
		})
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			return nil
		})
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
