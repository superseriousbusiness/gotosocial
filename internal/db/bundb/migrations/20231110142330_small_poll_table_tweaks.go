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
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Add new poll expiry column WITHOUT
			// the previously set NULL constraint.
			if _, err := tx.NewAddColumn().
				Table("polls").
				ColumnExpr("? TIMESTAMPTZ", bun.Ident("expires_at_new")).
				Exec(ctx); err != nil {
				return err
			}

			// Copy all data from old to new,
			// this won't cause anyn issues as
			// old column is NOT NULL and it's
			// only the new column that drops
			// this constraint to allow NULL.
			if _, err := tx.NewUpdate().
				Table("polls").
				Column("expires_at_new").
				Set("? = ?", bun.Ident("expires_at_new"), bun.Ident("expires_at")).
				Where("1"). // bun gets angry performing update over all rows
				Exec(ctx); err != nil {
				return err
			}

			// Drop the old poll expiry column.
			if _, err := tx.NewDropColumn().
				Table("polls").
				ColumnExpr("?", bun.Ident("expires_at")).
				Exec(ctx); err != nil {
				return err
			}

			// Rename the new expiry column
			// to the correct name (of the old).
			if _, err := tx.NewRaw(
				"ALTER TABLE ? RENAME COLUMN ? TO ?",
				bun.Ident("polls"),
				bun.Ident("expires_at_new"),
				bun.Ident("expires_at"),
			).Exec(ctx); err != nil {
				return err
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
