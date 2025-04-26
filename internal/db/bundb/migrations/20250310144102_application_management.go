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

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Add client_id index to token table,
			// needed for invalidation if/when the
			// token's app is deleted.
			if _, err := tx.
				NewCreateIndex().
				Table("tokens").
				Index("tokens_client_id_idx").
				Column("client_id").
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Update users to set all "created_by_application_id"
			// values to the instance application, to correct some
			// past issues where this wasn't set. Skip this if there's
			// no users though, as in that case we probably don't even
			// have an instance application yet.
			usersLen, err := tx.
				NewSelect().
				Table("users").
				Count(ctx)
			if err != nil {
				return err
			}

			if usersLen == 0 {
				// Nothing to do.
				return nil
			}

			// Get Instance account ID.
			var instanceAcctID string
			if err := tx.
				NewSelect().
				Table("accounts").
				Column("id").
				Where("? = ?", bun.Ident("username"), config.GetHost()).
				Where("? IS NULL", bun.Ident("domain")).
				Scan(ctx, &instanceAcctID); err != nil {
				return err
			}

			// Get the instance app ID.
			var instanceAppID string
			if err := tx.
				NewSelect().
				Table("applications").
				Column("id").
				Where("? = ?", bun.Ident("client_id"), instanceAcctID).
				Scan(ctx, &instanceAppID); err != nil {
				return err
			}

			// Set instance app ID on
			// users where it's null.
			if _, err := tx.
				NewUpdate().
				Table("users").
				Set("? = ?", bun.Ident("created_by_application_id"), instanceAppID).
				Where("? IS NULL", bun.Ident("created_by_application_id")).
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
