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

	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20240904084406_fedi_api_reject_interaction"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.SinBinStatus{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			for idx, col := range map[string]string{
				"sin_bin_statuses_account_uri_idx":     "account_uri",
				"sin_bin_statuses_domain_idx":          "domain",
				"sin_bin_statuses_in_reply_to_uri_idx": "in_reply_to_uri",
			} {
				if _, err := tx.
					NewCreateIndex().
					Table("sin_bin_statuses").
					Index(idx).
					Column(col).
					IfNotExists().
					Exec(ctx); err != nil {
					return err
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
