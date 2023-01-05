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
			// drop indexes for these tables
			if _, err := tx.
				NewDropIndex().
				Index("follow_requests_account_id_idx").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewDropIndex().
				Index("follow_requests_target_account_id_idx").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewDropIndex().
				Index("follows_account_id_idx").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewDropIndex().
				Index("follows_target_account_id_idx").
				Exec(ctx); err != nil {
				return err
			}

			// now re-create them with an ordering
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.FollowRequest{}).
				Index("follow_requests_account_id_idx").
				Column("account_id").
				ColumnExpr("updated_at DESC").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.FollowRequest{}).
				Index("follow_requests_target_account_id_idx").
				Column("target_account_id").
				ColumnExpr("updated_at DESC").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Follow{}).
				Index("follows_account_id_idx").
				Column("account_id").
				ColumnExpr("updated_at DESC").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Follow{}).
				Index("follows_target_account_id_idx").
				Column("target_account_id").
				ColumnExpr("updated_at DESC").
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
