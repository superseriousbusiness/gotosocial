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

	oldgtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20230328203024_migration_fix"
	newgtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20240318115336_account_settings"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/util"

	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		log.Info(ctx, "migrating account settings to new table, please wait...")
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Columns we'll be moving
			// to AccountSettings.
			var columns = []string{
				"reason",
				"privacy",
				"sensitive",
				"language",
				"status_content_type",
				"custom_css",
				"enable_rss",
				"hide_collections",
			}

			// Create the new account settings table.
			if _, err := tx.
				NewCreateTable().
				Model(&newgtsmodel.AccountSettings{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Select each local account.
			accounts := []*oldgtsmodel.Account{}
			if err := tx.
				NewSelect().
				TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
				Column("account.id").
				Column(columns...).
				Join(
					"JOIN ? AS ? ON ? = ?",
					bun.Ident("users"), bun.Ident("user"),
					bun.Ident("user.account_id"), bun.Ident("account.id"),
				).
				Scan(ctx, &accounts); err != nil {
				return err
			}

			// Create a settings entry for each existing account, taking
			// values from the old account model (with sensible defaults).
			for _, account := range accounts {
				settings := &newgtsmodel.AccountSettings{
					AccountID:         account.ID,
					CreatedAt:         account.CreatedAt,
					Reason:            account.Reason,
					Privacy:           newgtsmodel.Visibility(account.Privacy),
					Sensitive:         util.Ptr(util.PtrOrValue(account.Sensitive, false)),
					Language:          account.Language,
					StatusContentType: account.StatusContentType,
					CustomCSS:         account.CustomCSS,
					EnableRSS:         util.Ptr(util.PtrOrValue(account.EnableRSS, false)),
					HideCollections:   util.Ptr(util.PtrOrValue(account.HideCollections, false)),
				}

				// Insert the settings model.
				if _, err := tx.
					NewInsert().
					Model(settings).
					Exec(ctx); err != nil {
					return err
				}
			}

			// Drop now unused columns from accounts table.
			for _, column := range columns {
				if _, err := tx.
					NewDropColumn().
					Table("accounts").
					Column(column).
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
