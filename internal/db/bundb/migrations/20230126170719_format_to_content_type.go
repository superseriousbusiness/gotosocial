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
	"strings"

	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			var err error
			_, err = db.ExecContext(ctx, "ALTER TABLE ? ADD COLUMN ? TEXT", bun.Ident("accounts"), bun.Ident("status_content_type"))
			if err != nil && !(strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate column name") || strings.Contains(err.Error(), "SQLSTATE 42701")) {
				return err
			}
			_, err = db.ExecContext(ctx, "UPDATE ? SET ? = 'text/' || ? WHERE ? IS NOT NULL", bun.Ident("accounts"), bun.Ident("status_content_type"), bun.Ident("status_format"), bun.Ident("status_format"))
			if err != nil {
				return err
			}
			_, err = db.ExecContext(ctx, "ALTER TABLE ? DROP COLUMN ?", bun.Ident("accounts"), bun.Ident("status_format"))
			if err != nil {
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
