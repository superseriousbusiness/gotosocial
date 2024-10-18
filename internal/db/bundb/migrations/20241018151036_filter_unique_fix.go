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

package migrations

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			// Create the new filters table
			// with the unique constraint
			// set on AccountID + Title.
			if _, err := tx.
				NewCreateTable().
				ModelTableExpr("new_filters").
				Model((*gtsmodel.Filter)(nil)).
				Exec(ctx); err != nil {
				return err
			}

			// Explicitly specify columns to bring
			// from old table to new, to avoid any
			// potential Postgres shenanigans.
			columns := []string{
				"id",
				"created_at",
				"updated_at",
				"expires_at",
				"account_id",
				"title",
				"action",
				"context_home",
				"context_notifications",
				"context_public",
				"context_thread",
				"context_account",
			}

			// Copy all data for existing
			// filters to the new table.
			if _, err := tx.
				NewInsert().
				Table("new_filters").
				Table("filters").
				Column(columns...).
				Exec(ctx); err != nil {
				return err
			}

			// Drop the old table.
			if _, err := tx.
				NewDropTable().
				Table("filters").
				Exec(ctx); err != nil {
				return err
			}

			// Rename new table to old table.
			if _, err := tx.
				ExecContext(
					ctx,
					"ALTER TABLE ? RENAME TO ?",
					bun.Ident("new_filters"),
					bun.Ident("filters"),
				); err != nil {
				return err
			}

			// Index the new version
			// of the filters table.
			if _, err := tx.
				NewCreateIndex().
				Table("filters").
				Index("filters_account_id_idx").
				Column("account_id").
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			if db.Dialect().Name() == dialect.PG {
				// Rename "new_filters_pkey" from the
				// new table to just "filters_pkey".
				// This is only necessary on Postgres.
				if _, err := tx.ExecContext(
					ctx,
					"ALTER TABLE ? RENAME CONSTRAINT ? TO ?",
					bun.Ident("public.filters"),
					bun.Safe("new_filters_pkey"),
					bun.Safe("filters_pkey"),
				); err != nil {
					return err
				}
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
