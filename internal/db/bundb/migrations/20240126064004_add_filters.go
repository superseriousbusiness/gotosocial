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

	gtsmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20240126064004_add_filters"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Filter table.
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.Filter{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Filter keyword table.
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.FilterKeyword{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Filter status table.
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.FilterStatus{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Add indexes to the filter tables.
			for table, indexes := range map[string]map[string][]string{
				"filters": {
					"filters_account_id_idx": {"account_id"},
				},
				"filter_keywords": {
					"filter_keywords_account_id_idx": {"account_id"},
					"filter_keywords_filter_id_idx":  {"filter_id"},
				},
				"filter_statuses": {
					"filter_statuses_account_id_idx": {"account_id"},
					"filter_statuses_filter_id_idx":  {"filter_id"},
				},
			} {
				for index, columns := range indexes {
					if _, err := tx.
						NewCreateIndex().
						Table(table).
						Index(index).
						Column(columns...).
						IfNotExists().
						Exec(ctx); err != nil {
						return err
					}
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
