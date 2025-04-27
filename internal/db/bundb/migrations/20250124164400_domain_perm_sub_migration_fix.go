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

	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20241022153016_domain_permission_subscriptions"
	"github.com/uptrace/bun"
)

// This file exists because tobi is a silly billy and named two migration files
// with the same date part, so the latter migration didn't always run properly.
// The file just repeats migrations in 20250119112745_domain_permission_subscriptions.go,
// which will be a noop in most cases, but will fix some issues for those who
// were running snapshots between GtS v0.17.0 and GtS v0.18.0.
//
// See https://codeberg.org/superseriousbusiness/gotosocial/pulls/3679.
func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Create `domain_permission_subscriptions`.
			if _, err := tx.
				NewCreateTable().
				Model((*gtsmodel.DomainPermissionSubscription)(nil)).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Create indexes. Indices. Indie sexes.
			if _, err := tx.
				NewCreateIndex().
				Table("domain_permission_subscriptions").
				// Filter on permission type.
				Index("domain_permission_subscriptions_permission_type_idx").
				Column("permission_type").
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Table("domain_permission_subscriptions").
				// Sort by priority DESC.
				Index("domain_permission_subscriptions_priority_order_idx").
				ColumnExpr("? DESC", bun.Ident("priority")).
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
