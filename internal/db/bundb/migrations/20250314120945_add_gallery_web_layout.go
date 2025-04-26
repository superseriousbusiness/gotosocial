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

	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			// Add new column to settings.
			if _, err := tx.
				NewAddColumn().
				Table("account_settings").
				ColumnExpr(
					"? SMALLINT NOT NULL DEFAULT ?",
					bun.Ident("web_layout"), 1,
				).
				Exec(ctx); err != nil {
				return err
			}

			// Drop existing statuses web index as it's out of date.
			log.Info(ctx, "updating statuses_profile_web_view_idx, this may take a while, please wait!")
			if _, err := tx.
				NewDropIndex().
				Index("statuses_profile_web_view_idx").
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			// Note: "attachments" field is not included in
			// the index below as SQLite is fussy about using it,
			// and it prevents this index from being used
			// properly in non media-only queries.
			if _, err := tx.
				NewCreateIndex().
				Table("statuses").
				Index("statuses_profile_web_view_idx").
				Column(
					"account_id",
					"visibility",
					"in_reply_to_uri",
					"boost_of_id",
					"federated",
				).
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
