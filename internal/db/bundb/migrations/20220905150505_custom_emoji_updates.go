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
	"database/sql"

	gtsmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20211113114307_init"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
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
			if _, err := tx.ExecContext(ctx, "ALTER TABLE new_emojis RENAME TO emojis;"); err != nil {
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
