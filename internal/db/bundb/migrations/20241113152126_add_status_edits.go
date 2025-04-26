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
	"reflect"

	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20241113152126_add_status_edits"
	"code.superseriousbusiness.org/gotosocial/internal/log"

	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			statusType := reflect.TypeOf((*gtsmodel.Status)(nil))

			// Generate new Status.EditIDs column definition from bun.
			colDef, err := getBunColumnDef(tx, statusType, "EditIDs")
			if err != nil {
				return err
			}

			// Add EditIDs column to Status table.
			log.Info(ctx, "adding edits column to statuses table...")
			_, err = tx.NewAddColumn().
				Model((*gtsmodel.Status)(nil)).
				ColumnExpr(colDef).
				Exec(ctx)
			if err != nil {
				return err
			}

			// Create the main StatusEdits table.
			_, err = tx.NewCreateTable().
				IfNotExists().
				Model((*gtsmodel.StatusEdit)(nil)).
				Exec(ctx)
			return err
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
