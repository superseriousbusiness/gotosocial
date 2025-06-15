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
	"reflect"

	"code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250614160014_unauthed_web_updates/common"
	newmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250614160014_unauthed_web_updates/new"
	oldmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250614160014_unauthed_web_updates/old"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			var account *newmodel.Account
			accountType := reflect.TypeOf(account)

			// Add new columns to accounts
			// table if they don't exist already.
			for _, new := range []struct {
				dbCol     string
				fieldName string
			}{
				{
					dbCol:     "hides_to_public_from_unauthed_web",
					fieldName: "HidesToPublicFromUnauthedWeb",
				},
				{
					dbCol:     "hides_cc_public_from_unauthed_web",
					fieldName: "HidesCcPublicFromUnauthedWeb",
				},
			} {
				exists, err := doesColumnExist(
					ctx,
					tx,
					"accounts",
					new.dbCol,
				)
				if err != nil {
					return err
				}

				if exists {
					// Column already exists.
					continue
				}

				// Column doesn't exist yet, add it.
				colDef, err := getBunColumnDef(tx, accountType, new.fieldName)
				if err != nil {
					return fmt.Errorf("error making column def: %w", err)
				}

				log.Infof(ctx, "adding accounts.%s column...", new.dbCol)
				if _, err := tx.
					NewAddColumn().
					Model(account).
					ColumnExpr(colDef).
					Exec(ctx); err != nil {
					return fmt.Errorf("error adding column: %w", err)
				}
			}

			// For each account settings we have
			// stored on this instance, set the
			// new account columns to values
			// corresponding to the setting.
			allSettings := []*oldmodel.AccountSettings{}
			if err := tx.
				NewSelect().
				Model(allSettings).
				Scan(ctx); err != nil {
				return fmt.Errorf("error selecting settings: %w", err)
			}

			for _, settings := range allSettings {

				// Derive web visibility.
				var hideToPublic, hideCcPublic bool
				switch settings.WebVisibility {
				case common.VisibilityUnlocked:
					hideToPublic = false
					hideCcPublic = false
				case common.VisibilityPublic:
					hideToPublic = false
					hideCcPublic = true
				case common.VisibilityNone:
					hideToPublic = true
					hideCcPublic = true
				default:
					log.Warn(ctx, "local account had settings.WebVisibility != unlocked|public|none, skipping...")
					continue
				}

				// Update account.
				if _, err := tx.
					NewUpdate().
					Table("accounts").
					Set("? = ?", bun.Ident("hides_to_public_from_unauthed_web"), hideToPublic).
					Set("? = ?", bun.Ident("hides_cc_public_from_unauthed_web"), hideCcPublic).
					Where("? = ?", bun.Ident("id"), settings.AccountID).Exec(ctx); err != nil {
					return fmt.Errorf("error updating local account: %w", err)
				}
			}

			// Drop the old web_visibility column.
			if _, err := tx.
				NewDropColumn().
				Model((*oldmodel.AccountSettings)(nil)).
				Column("web_visibility").
				Exec(ctx); err != nil {
				return fmt.Errorf("error dropping old web_visibility column: %w", err)
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
