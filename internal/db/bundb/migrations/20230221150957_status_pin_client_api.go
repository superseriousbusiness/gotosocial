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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Drop the now unused 'pinned' column in statuses.
			if _, err := tx.ExecContext(ctx, "ALTER TABLE ? DROP COLUMN ?", bun.Ident("statuses"), bun.Ident("pinned")); err != nil &&
				!(strings.Contains(err.Error(), "no such column") || strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "SQLSTATE 42703")) {
				return err
			}

			// Create new (more useful) pinned_at column.
			if _, err := tx.NewAddColumn().Model(&gtsmodel.Status{}).ColumnExpr("? TIMESTAMPTZ", bun.Ident("pinned_at")).Exec(ctx); err != nil &&
				!(strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate column name") || strings.Contains(err.Error(), "SQLSTATE 42701")) {
				return err
			}

			// Index new column appropriately.
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Status{}).
				Index("statuses_account_id_pinned_at_idx").
				Column("account_id", "pinned_at").
				Exec(ctx); err != nil {
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
