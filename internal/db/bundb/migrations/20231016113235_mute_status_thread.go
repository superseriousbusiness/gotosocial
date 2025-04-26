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

	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Create thread table.
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.Thread{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Create thread intermediate table.
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.ThreadToStatus{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Drop old pkey constraint from
			// deprecated status mute table.
			//
			// This is only necessary with postgres.
			if tx.Dialect().Name() == dialect.PG {
				if _, err := tx.ExecContext(
					ctx,
					"ALTER TABLE ? DROP CONSTRAINT IF EXISTS ?",
					bun.Ident("status_mutes"),
					bun.Safe("status_mutes_pkey"),
				); err != nil {
					return err
				}
			}

			// Drop old index.
			if _, err := tx.
				NewDropIndex().
				Index("status_mutes_account_id_target_account_id_status_id_idx").
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			// Drop deprecated status mute table.
			if _, err := tx.
				NewDropTable().
				Table("status_mutes").
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			// Create new thread mute table.
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.ThreadMute{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			log.Info(ctx, "creating a new index on the statuses table, please wait and don't interrupt it (this may take a few minutes)")

			// Update statuses to add thread ID column.
			_, err := tx.ExecContext(
				ctx,
				"ALTER TABLE ? ADD COLUMN ? CHAR(26)",
				bun.Ident("statuses"),
				bun.Ident("thread_id"),
			)
			if err != nil && !(strings.Contains(err.Error(), "already exists") ||
				strings.Contains(err.Error(), "duplicate column name") ||
				strings.Contains(err.Error(), "SQLSTATE 42701")) {
				return err
			}

			// Index new + existing tables properly.
			for table, indexes := range map[string]map[string][]string{
				"threads": {
					"threads_id_idx": {"id"},
				},
				"thread_mutes": {
					"thread_mutes_id_idx": {"id"},
					// Eg., check if target thread is muted by account.
					"thread_mutes_thread_id_account_id_idx": {"thread_id", "account_id"},
				},
				"statuses": {
					// Eg., select all statuses in a thread.
					"statuses_thread_id_idx": {"thread_id"},
				},
			} {
				for index, columns := range indexes {
					if _, err := tx.
						NewCreateIndex().
						Table(table).
						Index(index).
						Column(columns...).
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
