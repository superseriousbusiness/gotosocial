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
			log.Info(ctx, "renaming / re-adding some indexes; this may take some time, please be patient and don't interrupt this!")

			// Remove misnamed indexes from when
			// account_actions renamed to admin_actions
			for _, index := range []string{
				"account_actions_target_category_idx",
				"account_actions_target_id_idx",
				"account_actions_type_idx",
				"account_actions_account_id_idx",
			} {
				if _, err := tx.
					NewDropIndex().
					Index(index).
					IfExists().
					Exec(ctx); err != nil {
					return err
				}
			}

			type spec struct {
				index   string
				table   string
				columns []string
			}

			for _, spec := range []spec{
				// Rename the admin actions indexes.
				{
					index:   "admin_actions_target_category_idx",
					table:   "admin_actions",
					columns: []string{"target_category"},
				},
				{
					index:   "admin_actions_target_id_idx",
					table:   "admin_actions",
					columns: []string{"target_id"},
				},
				{
					index:   "admin_actions_type_idx",
					table:   "admin_actions",
					columns: []string{"type"},
				},
				{
					index:   "admin_actions_account_id_idx",
					table:   "admin_actions",
					columns: []string{"account_id"},
				},

				// Recreate indexes that may have been removed
				// by a bodged version of the previous migration
				// (this PR is my penance -- tobi).
				{
					index:   "list_entries_list_id_idx",
					table:   "list_entries",
					columns: []string{"list_id"},
				},
				{
					index:   "accounts_domain_idx",
					table:   "accounts",
					columns: []string{"domain"},
				},
				{
					index:   "status_faves_account_id_idx",
					table:   "status_faves",
					columns: []string{"account_id"},
				},
				{
					index:   "statuses_account_id_idx",
					table:   "statuses",
					columns: []string{"account_id"},
				},
				{
					index:   "status_to_tags_tag_id_idx",
					table:   "status_to_tags",
					columns: []string{"tag_id"},
				},
			} {
				if _, err := tx.
					NewCreateIndex().
					Table(spec.table).
					Index(spec.index).
					Column(spec.columns...).
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
