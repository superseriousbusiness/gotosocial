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

	"github.com/uptrace/bun"

	oldmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20220315160814_admin_account_actions"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Select all old actions.
			var adminAccountActions []*oldmodel.AdminAccountAction
			if err := tx.
				NewSelect().
				Model(&adminAccountActions).
				Scan(ctx); err != nil {
				return err
			}

			// Create the new table.
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.AdminAction{}).
				Exec(ctx); err != nil {
				return err
			}

			// Index new table properly.
			for index, columns := range map[string][]string{
				"account_actions_id_idx": {"id"},
				// Eg., select all actions of given category.
				"account_actions_target_category_idx": {"target_category"},
				// Eg., select all actions targeting given id.
				"account_actions_target_id_idx": {"target_id"},
				// Eg., select all actions of given type.
				"account_actions_type_idx": {"type"},
				// Eg., select all actions by given account id.
				"account_actions_account_id_idx": {"account_id"},
			} {
				if _, err := tx.
					NewCreateIndex().
					Table("admin_actions").
					Index(index).
					Column(columns...).
					Exec(ctx); err != nil {
					return err
				}
			}

			// Insert old format entries into new table.
			for _, oldAction := range adminAccountActions {
				newAction := &gtsmodel.AdminAction{
					ID:             oldAction.ID,
					CreatedAt:      oldAction.CreatedAt,
					UpdatedAt:      oldAction.UpdatedAt,
					TargetCategory: gtsmodel.AdminActionCategoryAccount,
					TargetID:       oldAction.TargetAccountID,
					Type:           gtsmodel.ParseAdminActionType(string(oldAction.Type)),
					AccountID:      oldAction.AccountID,
					Text:           oldAction.Text,
					SendEmail:      util.Ptr(oldAction.SendEmail),
					ReportIDs:      []string{oldAction.ReportID},
				}

				if _, err := tx.
					NewInsert().
					Model(newAction).
					Exec(ctx); err != nil {
					return err
				}
			}

			// Drop the old table.
			if _, err := tx.
				NewDropTable().
				Table("admin_account_actions").
				Exec(ctx); err != nil {
				return err
			}

			// Drop any remaining old indexes.
			for _, idxName := range []string{
				"admin_account_actions_pkey",
				"admin_account_actions_account_id_idx",
				"admin_account_actions_target_account_id_idx",
				"admin_account_actions_type_idx",
			} {
				if _, err := tx.
					NewDropIndex().
					Index(idxName).
					IfExists().
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
