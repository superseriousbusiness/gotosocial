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
	"reflect"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			model := &gtsmodel.WebPushSubscription{}

			// Get the column definition for the new policy column.
			modelType := reflect.TypeOf(model)
			columnDef, err := getBunColumnDef(tx, modelType, "Policy")
			if err != nil {
				return err
			}

			// Add the policy column.
			switch tx.Dialect().Name() {
			case dialect.SQLite:
				// Doesn't support Bun feature AlterColumnExists.
				if _, err = tx.
					NewAddColumn().
					Model(model).
					ColumnExpr(columnDef).
					Exec(ctx); // nocollapse
				err != nil && !strings.Contains(err.Error(), "duplicate column name") {
					// Return errors that aren't about this column already existing.
					return err
				}

			case dialect.PG:
				// Supports Bun feature AlterColumnExists.
				if _, err = tx.
					NewAddColumn().
					Model(model).
					ColumnExpr(columnDef).
					IfNotExists().
					Exec(ctx); // nocollapse
				err != nil {
					return err
				}

			default:
				panic("unsupported db type")
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
