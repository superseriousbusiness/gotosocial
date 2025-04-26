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

	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		log.Info(ctx, "reindexing accounts (accounts_paging_idx); this may take a few minutes, please don't interrupt this migration!")

		q := db.NewCreateIndex().
			TableExpr("accounts").
			Index("accounts_paging_idx").
			IfNotExists()

		switch d := db.Dialect().Name(); d {
		case dialect.SQLite:
			q = q.ColumnExpr(
				"COALESCE(?, ?) || ? || ?",
				bun.Ident("domain"), "",
				"/@",
				bun.Ident("username"),
			)

		// Specify C collation for Postgres to ensure
		// alphabetic sort order is similar enough to
		// SQLite (which uses BINARY sort by default).
		//
		// See:
		//
		//   - https://www.postgresql.org/docs/current/collation.html#COLLATION-MANAGING-STANDARD
		//   - https://sqlite.org/datatype3.html#collation
		case dialect.PG:
			q = q.ColumnExpr(
				"(COALESCE(?, ?) || ? || ?) COLLATE ?",
				bun.Ident("domain"), "",
				"/@",
				bun.Ident("username"),
				bun.Ident("C"),
			)

		default:
			log.Panicf(ctx, "dialect %s was neither postgres nor sqlite", d)
		}

		if _, err := q.Exec(ctx); err != nil {
			return err
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
