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

	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// List table.
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.List{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Add indexes to the List table.
			for index, columns := range map[string][]string{
				"lists_id_idx":         {"id"},
				"lists_account_id_idx": {"account_id"},
			} {
				if _, err := tx.
					NewCreateIndex().
					Table("lists").
					Index(index).
					Column(columns...).
					Exec(ctx); err != nil {
					return err
				}
			}

			// List entry table.
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.ListEntry{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Add indexes to the List entry table.
			for index, columns := range map[string][]string{
				"list_entries_id_idx":        {"id"},
				"list_entries_list_id_idx":   {"list_id"},
				"list_entries_follow_id_idx": {"follow_id"},
			} {
				if _, err := tx.
					NewCreateIndex().
					Table("list_entries").
					Index(index).
					Column(columns...).
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
