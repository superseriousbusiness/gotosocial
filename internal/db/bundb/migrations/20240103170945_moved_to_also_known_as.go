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

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Drop now-unused columns
			// from accounts table.
			for _, column := range []string{
				"also_known_as",
				"moved_to_account_id",
			} {
				if _, err := tx.
					NewDropColumn().
					Table("accounts").
					Column(column).
					Exec(ctx); err != nil {
					return err
				}
			}

			// Create new columns.
			if _, err := tx.
				NewAddColumn().
				Table("accounts").
				ColumnExpr("? VARCHAR", bun.Ident("moved_to_uri")).
				Exec(ctx); err != nil {
				return err
			}

			switch tx.Dialect().Name() {
			case dialect.SQLite:
				if _, err := tx.
					NewAddColumn().
					Table("accounts").
					ColumnExpr("? VARCHAR", bun.Ident("also_known_as_uris")).
					Exec(ctx); err != nil {
					return err
				}
			case dialect.PG:
				if _, err := tx.
					NewAddColumn().
					Table("accounts").
					ColumnExpr("? VARCHAR ARRAY", bun.Ident("also_known_as_uris")).
					Exec(ctx); err != nil {
					return err
				}
			default:
				panic("db conn was neither pg not sqlite")
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
