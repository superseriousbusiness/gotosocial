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

	gtsmodel "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		var expression string
		switch db.Dialect().Name() {
		case dialect.PG:
			expression = "(mentions IS NULL OR CARDINALITY(mentions) = 0)"
		case dialect.SQLite:
			expression = "(mentions IS NULL OR json_array_length(mentions) = 0)"
		default:
			panic("db conn was neither pg not sqlite")
		}

		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			log.Info(ctx,
				"removing previous statuses_account_view_idx and reindexing statuses; "+
					"this may take a few minutes, please don't interrupt this migration",
			)

			// Remove old index with columns
			// in really awkward order.
			if _, err := tx.
				NewDropIndex().
				Model((*gtsmodel.Status)(nil)).
				Index("statuses_account_view_idx").
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			// Create new index with
			// columns in desired order.
			if _, err := tx.
				NewCreateIndex().
				Model((*gtsmodel.Status)(nil)).
				Index("statuses_account_view_idx").
				Column(
					"account_id",
					"in_reply_to_uri",
					"in_reply_to_account_id",
				).
				ColumnExpr(expression).
				ColumnExpr("id DESC").
				IfNotExists().
				Exec(ctx); err != nil {
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
