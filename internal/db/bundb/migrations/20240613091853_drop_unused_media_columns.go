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

	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		log.Info(
			ctx,
			"dropping unused media attachments columns, please wait; "+
				"this may take a long time if your database has lots of media attachments, don't interrupt it!",
		)
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
