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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			// Add new local statuses count view.
			in := []int16{
				int16(gtsmodel.VisibilityPublic),
				int16(gtsmodel.VisibilityUnlocked),
				int16(gtsmodel.VisibilityFollowersOnly),
				int16(gtsmodel.VisibilityMutualsOnly),
			}
			if _, err := tx.
				NewRaw(
					"CREATE VIEW ? AS "+
						"SELECT COUNT(*) FROM ? "+
						"WHERE (? = ?) AND (? IN (?)) AND (? = ?)",
					bun.Ident("statuses_local_count_view"),
					bun.Ident("statuses"),
					bun.Ident("local"), true,
					bun.Ident("visibility"), bun.In(in),
					bun.Ident("pending_approval"), false,
				).
				Exec(ctx); err != nil {
				return err
			}

			// Drop existing local index.
			if _, err := tx.
				NewDropIndex().
				Index("statuses_local_idx").
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			// Add new multicolumn local statuses
			// index that works for the local count
			// view and the local timeline query.
			if _, err := tx.
				NewCreateIndex().
				Table("statuses").
				Index("statuses_local_idx").
				Column("local", "visibility", "pending_approval").
				ColumnExpr("? DESC", bun.Ident("id")).
				IfNotExists().
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
