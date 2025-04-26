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
		// Add reason to users table.
		_, err := db.ExecContext(ctx,
			"ALTER TABLE ? ADD COLUMN ? TEXT",
			bun.Ident("users"), bun.Ident("reason"),
		)
		if err != nil {
			e := err.Error()
			if !(strings.Contains(e, "already exists") ||
				strings.Contains(e, "duplicate column name") ||
				strings.Contains(e, "SQLSTATE 42701")) {
				return err
			}
		}

		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Get reasons from
			// account settings.
			type idReason struct {
				AccountID string
				Reason    string
			}

			reasons := []idReason{}
			if err := tx.
				NewSelect().
				Table("account_settings").
				Column("account_id", "reason").
				Scan(ctx, &reasons); err != nil {
				return err
			}

			// Add each reason to appropriate user.
			for _, r := range reasons {
				if _, err := tx.
					NewUpdate().
					Table("users").
					Set("? = ?", bun.Ident("reason"), r.Reason).
					Where("? = ?", bun.Ident("account_id"), r.AccountID).
					Exec(ctx, &reasons); err != nil {
					return err
				}
			}

			// Remove now-unused column
			// from account settings.
			if _, err := tx.
				NewDropColumn().
				Table("account_settings").
				Column("reason").
				Exec(ctx); err != nil {
				return err
			}

			// Remove now-unused columns from users.
			for _, column := range []string{
				"current_sign_in_at",
				"current_sign_in_ip",
				"last_sign_in_at",
				"last_sign_in_ip",
				"sign_in_count",
				"chosen_languages",
				"filtered_languages",
			} {
				if _, err := tx.
					NewDropColumn().
					Table("users").
					Column(column).
					Exec(ctx); err != nil {
					return err
				}
			}

			// Create new UsersDenied table.
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.DeniedUser{}).
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
