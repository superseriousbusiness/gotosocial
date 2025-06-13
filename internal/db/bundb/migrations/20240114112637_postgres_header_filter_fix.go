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
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		// Run the first bit in a transaction
		// since we're not expecting to encounter
		// errors, and any we do encounter will
		// stop us in our tracks.
		err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Drop each of the old versions of the
			// header tables. Normally dropping tables
			// is a big no-no but this migration happens
			// while header filters weren't even in a
			// release yet, so let's go for it.
			for _, table := range []string{
				"header_filter_allows",
				"header_filter_blocks",
			} {
				_, err := tx.NewDropTable().
					IfExists().
					Table(table).
					Exec(ctx)
				if err != nil {
					return err
				}
			}

			// Recreate header tables using
			// the most up-to-date model.
			for _, model := range []any{
				&gtsmodel.HeaderFilterAllow{},
				&gtsmodel.HeaderFilterBlock{},
			} {
				_, err := tx.NewCreateTable().
					IfNotExists().
					Model(model).
					Exec(ctx)
				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return err
		}

		// On Postgres the constraints might still
		// be kicking around from a partial failed
		// migration, so make sure they're gone now.
		// Dropping a constraint will also drop any
		// indexes supporting the constraint, as per:
		//
		// https://www.postgresql.org/docs/16/sql-altertable.html#SQL-ALTERTABLE-DESC-DROP-CONSTRAINT
		//
		// We run this part outside of a transaction
		// because we don't check for errors, and we
		// don't want an error in the first query to
		// foul the transaction and stop the second
		// query from running.
		if db.Dialect().Name() == dialect.PG {
			for _, table := range []string{
				"header_filter_allows",
				"header_filter_blocks",
			} {
				// Just swallow any errors
				// here, we're not bothered.
				_, _ = db.ExecContext(
					ctx,
					"ALTER TABLE ? DROP CONSTRAINT IF EXISTS ?",
					bun.Ident(table),
					bun.Safe("header_regex"),
				)
			}
		}

		return nil
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
