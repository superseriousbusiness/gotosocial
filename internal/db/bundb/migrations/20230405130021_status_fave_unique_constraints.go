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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		// To update fave constraints, we need to migrate faves into a new table.
		// See section 7 here: https://www.sqlite.org/lang_altertable.html

		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Remove any duplicate faves that were created before constraints.
			// We need to ensure that we don't just delete all faves that are
			// duplicates--we should keep the original, non-duplicate fave.
			// So, produced query will look something like this:
			//
			//	DELETE FROM "status_faves"
			//	WHERE id IN (
			//		WITH cte AS (
			//			SELECT
			//				"id",
			//				ROW_NUMBER() OVER(
			//					PARTITION BY "account_id", "status_id"
			//					ORDER BY "account_id", "status_id"
			//				) AS "row_number"
			//			FROM status_faves
			//		)
			//		SELECT "id" FROM cte
			//		WHERE "row_number" > 1
			//	)
			//
			// The above query only deletes status_faves with ids that are
			// in the subquery. The subquery selects the IDs of all duplicate
			// status faves past the first one, where 'duplicate' means 'has
			// the same account id and status id'.
			overQ := tx.NewRaw(
				"PARTITION BY ?, ? ORDER BY ?, ?",
				bun.Ident("account_id"),
				bun.Ident("status_id"),
				bun.Ident("account_id"),
				bun.Ident("status_id"),
			)

			rowNumberQ := tx.NewRaw(
				"SELECT ?, ROW_NUMBER() OVER(?) AS ? FROM status_faves",
				bun.Ident("id"),
				overQ,
				bun.Ident("row_number"),
			)

			inQ := tx.NewRaw(
				"WITH cte AS (?) SELECT ? FROM cte WHERE ? > 1",
				rowNumberQ,
				bun.Ident("id"),
				bun.Ident("row_number"),
			)

			if _, err := tx.
				NewDelete().
				Table("status_faves").
				Where("id IN (?)", inQ).
				Exec(ctx); err != nil {
				return err
			}

			// Create the new faves table.
			if _, err := tx.
				NewCreateTable().
				ModelTableExpr("new_status_faves").
				Model(&gtsmodel.StatusFave{}).
				Exec(ctx); err != nil {
				return err
			}

			// Specify columns explicitly to
			// avoid any Postgres shenanigans.
			columns := []string{
				"id",
				"created_at",
				"updated_at",
				"account_id",
				"target_account_id",
				"status_id",
				"uri",
			}

			// Copy remaining faves to the new table.
			if _, err := tx.
				NewInsert().
				Table("new_status_faves").
				Table("status_faves").
				Column(columns...).
				Exec(ctx); err != nil {
				return err
			}

			// Drop the old table.
			if _, err := tx.
				NewDropTable().
				Table("status_faves").
				Exec(ctx); err != nil {
				return err
			}

			// Rename new table to old table.
			if _, err := tx.
				ExecContext(
					ctx,
					"ALTER TABLE ? RENAME TO ?",
					bun.Ident("new_status_faves"),
					bun.Ident("status_faves"),
				); err != nil {
				return err
			}

			// Add indexes to the new table.
			for index, columns := range map[string][]string{
				"status_faves_id_idx":         {"id"},
				"status_faves_account_id_idx": {"account_id"},
				"status_faves_status_id_idx":  {"status_id"},
			} {
				if _, err := tx.
					NewCreateIndex().
					Table("status_faves").
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
