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
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Create a new column for status_content_type.
			if _, err := tx.ExecContext(ctx, "ALTER TABLE ? ADD COLUMN ? TEXT", bun.Ident("accounts"), bun.Ident("status_content_type")); err != nil &&
				!(strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate column name") || strings.Contains(err.Error(), "SQLSTATE 42701")) {
				return err
			}

			// Port values into this column from the old status_format column,
			// prepending 'text/' to the value to derive the mime type.
			//
			// Eg, 'plain' status_format becomes 'text/plain' status_content_type.
			q := tx.
				NewUpdate().
				Table("accounts").
				Where("? IS NOT NULL", bun.Ident("status_format"))

			// We need to switch here because Postgres and SQLite
			// have different syntaxes for concatenation.
			switch tx.Dialect().Name() {
			case dialect.SQLite:
				q = q.
					Set("? = ? || ?", bun.Ident("status_content_type"), "text/", bun.Ident("status_format"))
			case dialect.PG:
				q = q.
					Set("? = CONCAT(?, ?)", bun.Ident("status_content_type"), "text/", bun.Ident("status_format"))
			default:
				panic("db conn was neither pg not sqlite")
			}

			if _, err := q.Exec(ctx); err != nil {
				return err
			}

			// Drop the old status_format column.
			if _, err := tx.ExecContext(ctx, "ALTER TABLE ? DROP COLUMN ?", bun.Ident("accounts"), bun.Ident("status_format")); err != nil &&
				!(strings.Contains(err.Error(), "no such column") || strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "SQLSTATE 42703")) {
				return err
			}

			return nil
		})
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			var err error
			_, err = db.ExecContext(ctx, "ALTER TABLE ? ADD COLUMN ? TEXT", bun.Ident("accounts"), bun.Ident("status_format"))
			if err != nil && !(strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate column name") || strings.Contains(err.Error(), "SQLSTATE 42701")) {
				return err
			}
			_, err = db.ExecContext(ctx, "UPDATE ? SET ? = SUBSTR(?, 6) WHERE ? IS NOT NULL", bun.Ident("accounts"), bun.Ident("status_format"), bun.Ident("status_content_type"), bun.Ident("status_content_type"))
			if err != nil {
				return err
			}
			_, err = db.ExecContext(ctx, "ALTER TABLE ? DROP COLUMN ?", bun.Ident("accounts"), bun.Ident("status_content_type"))
			if err != nil {
				return err
			}
			return nil
		})
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
