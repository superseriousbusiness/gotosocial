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
	"fmt"
	"strings"

	oldgtsmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20230328203024_migration_fix"
	newgtsmodel "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"

	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		log.Info(ctx, "migrating account settings to new table, please wait...")

		// Add settings_id to accounts table.
		//
		// Do this outside of the transaction so it doesn't mess
		// up the transaction if the column exists already.
		_, err := db.ExecContext(ctx,
			"ALTER TABLE ? ADD COLUMN ? CHAR(26)",
			bun.Ident("accounts"), bun.Ident("settings_id"),
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
				settingsID, err := id.NewRandomULID()
				if err != nil {
					return fmt.Errorf("error creating settingsID: %w", err)
				}

				settings := &newgtsmodel.AccountSettings{
					ID:                settingsID,
					CreatedAt:         account.CreatedAt,
					Reason:            account.Reason,
					Privacy:           newgtsmodel.Visibility(account.Privacy),
					Sensitive:         util.Ptr(util.PtrValueOr(account.Sensitive, false)),
					Language:          account.Language,
					StatusContentType: account.StatusContentType,
					CustomCSS:         account.CustomCSS,
					EnableRSS:         util.Ptr(util.PtrValueOr(account.EnableRSS, false)),
					HideCollections:   util.Ptr(util.PtrValueOr(account.HideCollections, false)),
				}

				// Insert the settings model.
				if _, err := tx.
					NewInsert().
					Model(settings).
					Exec(ctx); err != nil {
					return err
				}

				// Update account with new settings ID.
				if _, err := tx.
					NewUpdate().
					Table("accounts").
					Set("? = ?", bun.Ident("settings_id"), settings.ID).
					Where("? = ?", bun.Ident("id"), account.ID).
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
