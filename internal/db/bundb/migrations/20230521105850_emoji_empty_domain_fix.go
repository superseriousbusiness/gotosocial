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

	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20230521105850_emoji_empty_domain_fix"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// SQLite doesn't mind creating multiple constraints with
			// the same name, but Postgres balks at it, so remove
			// constraints before we go editing the emoji tables.
			if tx.Dialect().Name() == dialect.PG {
				for _, constraint := range []string{
					"shortcodedomain", // initial constraint name
					"domainshortcode", // later constraint name
				} {
					if _, err := tx.ExecContext(
						ctx,
						"ALTER TABLE ? DROP CONSTRAINT IF EXISTS ?",
						bun.Ident("emojis"),
						bun.Safe(constraint),
					); err != nil {
						return err
					}
				}
			}

			// Set all existing emoji domains to null
			// where domain is '', to fix the bug!
			if _, err := tx.
				NewUpdate().
				Table("emojis").
				Set("? = NULL", bun.Ident("domain")).
				Where("? = ?", bun.Ident("domain"), "").
				Exec(ctx); err != nil {
				return err
			}

			// Create the new emojis table.
			if _, err := tx.
				NewCreateTable().
				ModelTableExpr("new_emojis").
				Model(&gtsmodel.Emoji{}).
				Exec(ctx); err != nil {
				return err
			}

			// Specify columns explicitly to
			// avoid any Postgres shenanigans.
			columns := []string{
				"id",
				"created_at",
				"updated_at",
				"shortcode",
				"domain",
				"image_remote_url",
				"image_static_remote_url",
				"image_url",
				"image_static_url",
				"image_path",
				"image_static_path",
				"image_content_type",
				"image_static_content_type",
				"image_file_size",
				"image_static_file_size",
				"image_updated_at",
				"disabled",
				"uri",
				"visible_in_picker",
				"category_id",
			}

			// Copy existing emojis to the new table.
			if _, err := tx.
				NewInsert().
				Table("new_emojis").
				Table("emojis").
				Column(columns...).
				Exec(ctx); err != nil {
				return err
			}

			// Drop the old table.
			if _, err := tx.
				NewDropTable().
				Table("emojis").
				Exec(ctx); err != nil {
				return err
			}

			// Rename new table to old table.
			if _, err := tx.
				ExecContext(
					ctx,
					"ALTER TABLE ? RENAME TO ?",
					bun.Ident("new_emojis"),
					bun.Ident("emojis"),
				); err != nil {
				return err
			}

			// Add indexes to the new table.
			for index, columns := range map[string][]string{
				"emojis_id_idx":               {"id"},
				"emojis_available_custom_idx": {"visible_in_picker", "disabled", "shortcode"},
				"emojis_image_static_url_idx": {"image_static_url"},
				"emojis_uri_idx":              {"uri"},
			} {
				if _, err := tx.
					NewCreateIndex().
					Table("emojis").
					Index(index).
					Column(columns...).
					Exec(ctx); err != nil {
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
