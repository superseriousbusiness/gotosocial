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

package migrations

import (
	"context"
	"database/sql"

	newgtsmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20220905150505_custom_emoji_updates"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// create the new emojis table
			if _, err := tx.
				NewCreateTable().
				ModelTableExpr("new_emojis").
				Model(&newgtsmodel.Emoji{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// move all old emojis to the new table
			oldEmojis := []*newgtsmodel.Emoji{}
			if err := tx.
				NewSelect().
				Model(&oldEmojis).
				Scan(ctx); err != nil && err != sql.ErrNoRows {
				return err
			}

			for _, emoji := range oldEmojis {
				if _, err := tx.
					NewInsert().
					Model(emoji).
					Exec(ctx); err != nil {
					return err
				}
			}

			// we have all the data we need from the old table, so we can safely drop it now
			if _, err := tx.NewDropTable().Model(&newgtsmodel.Emoji{}).Exec(ctx); err != nil {
				return err
			}

			// rename the new table to the same name as the old table was
			if _, err := tx.ExecContext(ctx, "ALTER TABLE new_emojis RENAME TO emojis;"); err != nil {
				return err
			}

			// add indexes to the new table
			if _, err := tx.
				NewCreateIndex().
				Model(&newgtsmodel.Emoji{}).
				Index("emojis_id_idx").
				Column("id").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&newgtsmodel.Emoji{}).
				Index("emojis_uri_idx").
				Column("uri").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&newgtsmodel.Emoji{}).
				Index("emojis_available_custom_idx").
				Column("visible_in_picker", "disabled", "domain", "shortcode ASC").
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
