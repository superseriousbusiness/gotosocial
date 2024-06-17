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

			for _, dropcase := range []struct {
				table string
				col   string
			}{
				{table: "media_attachments", col: "file_updated_at"},
				{table: "media_attachments", col: "thumbnail_updated_at"},
				{table: "emojis", col: "thumbnail_updated_at"},
			} {
				// For each case check the column actually exists on database.
				exists, err := doesColumnExist(ctx, tx, dropcase.table, dropcase.col)
				if err != nil {
					return err
				}

				if exists {
					// Now actually drop the column.
					if _, err := tx.NewDropColumn().
						Table(dropcase.table).
						Column(dropcase.col).
						Exec(ctx); err != nil {
						return err
					}
				}
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

func doesColumnExist(ctx context.Context, tx bun.Tx, table, col string) (bool, error) {
	var n int
	var err error
	switch tx.Dialect().Name() {
	case dialect.SQLite:
		err = tx.NewRaw("SELECT COUNT(*) FROM pragma_table_info(?) WHERE name=?", table, col).Scan(ctx, &n)
	case dialect.PG:
		err = tx.NewRaw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name=? and column_name=?", table, col).Scan(ctx, &n)
	default:
		panic("unexpected dialect")
	}
	return (n > 0), err
}
