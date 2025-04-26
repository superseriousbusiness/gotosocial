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
	"database/sql"

	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20211113114307_init"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// SQLite doesn't mind creating multiple constraints with the same name,
			// but Postgres balks at it, so remove the constraint before we go editing
			// the emoji tables.
			if tx.Dialect().Name() == dialect.PG {
				if _, err := tx.ExecContext(ctx, "ALTER TABLE ? DROP CONSTRAINT ?", bun.Ident("emojis"), bun.Safe("shortcodedomain")); err != nil {
					return err
				}
			}

			// create the new emojis table
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.Emoji{}).
				ModelTableExpr("new_emojis").
				Exec(ctx); err != nil {
				return err
			}

			// move all old emojis to the new table
			currentEmojis := []*gtsmodel.Emoji{}
			if err := tx.
				NewSelect().
				Model(&currentEmojis).
				Scan(ctx); err != nil && err != sql.ErrNoRows {
				return err
			}

			for _, currentEmoji := range currentEmojis {
				if _, err := tx.
					NewInsert().
					Model(currentEmoji).
					ModelTableExpr("new_emojis").
					Exec(ctx); err != nil {
					return err
				}
			}

			// we have all the data we need from the old table, so we can safely drop it now
			if _, err := tx.NewDropTable().Model(&gtsmodel.Emoji{}).Exec(ctx); err != nil {
				return err
			}

			// rename the new table to the same name as the old table was
			if _, err := tx.ExecContext(ctx, "ALTER TABLE ? RENAME TO ?", bun.Ident("new_emojis"), bun.Ident("emojis")); err != nil {
				return err
			}

			// add indexes to the new table
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Emoji{}).
				Index("emojis_id_idx").
				Column("id").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Emoji{}).
				Index("emojis_uri_idx").
				Column("uri").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Emoji{}).
				Index("emojis_available_custom_idx").
				Column("visible_in_picker", "disabled", "shortcode").
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
