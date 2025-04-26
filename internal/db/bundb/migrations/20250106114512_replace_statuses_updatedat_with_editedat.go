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
	"fmt"
	"reflect"

	oldmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250106114512_replace_statuses_updatedat_with_editedat"
	newmodel "code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			var newStatus *newmodel.Status
			newStatusType := reflect.TypeOf(newStatus)

			// Generate new Status.EditedAt column definition from bun.
			colDef, err := getBunColumnDef(tx, newStatusType, "EditedAt")
			if err != nil {
				return fmt.Errorf("error making column def: %w", err)
			}

			log.Info(ctx, "adding statuses.edited_at column...")
			_, err = tx.NewAddColumn().Model(newStatus).
				ColumnExpr(colDef).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("error adding column: %w", err)
			}

			var whereSQL string
			var whereArg []any

			// Check for an empty length
			// EditIDs JSON array, with different
			// SQL depending on connected database.
			switch tx.Dialect().Name() {
			case dialect.SQLite:
				whereSQL = "NOT (json_array_length(?) = 0 OR ? IS NULL)"
				whereArg = []any{bun.Ident("edits"), bun.Ident("edits")}
			case dialect.PG:
				whereSQL = "NOT (CARDINALITY(?) = 0 OR ? IS NULL)"
				whereArg = []any{bun.Ident("edits"), bun.Ident("edits")}
			default:
				panic("unsupported db type")
			}

			log.Info(ctx, "setting edited_at = updated_at where not empty(edits)...")
			res, err := tx.NewUpdate().Model(newStatus).Where(whereSQL, whereArg...).
				Set("? = ?",
					bun.Ident("edited_at"),
					bun.Ident("updated_at"),
				).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("error updating columns: %w", err)
			}

			count, _ := res.RowsAffected()
			log.Infof(ctx, "updated %d statuses", count)

			log.Info(ctx, "removing statuses.updated_at column...")
			_, err = tx.NewDropColumn().Model((*oldmodel.Status)(nil)).
				Column("updated_at").
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("error dropping column: %w", err)
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
