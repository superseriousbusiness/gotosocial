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

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// create the new status pins table
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.StatusPin{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// index appropriately
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.StatusPin{}).
				Index("status_pins_account_id_id_idx").
				Column("account_id").
				ColumnExpr("id DESC").
				Exec(ctx); err != nil {
				return err
			}

			// drop the now unused 'pinned' column in statuses
			if _, err := tx.ExecContext(ctx, "ALTER TABLE ? DROP COLUMN ?", bun.Ident("statuses"), bun.Ident("pinned")); err != nil &&
				!(strings.Contains(err.Error(), "no such column") || strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "SQLSTATE 42703")) {
				return err
			}

			return nil
		})
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return nil
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
