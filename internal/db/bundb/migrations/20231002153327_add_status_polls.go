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

	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20231002153327_add_status_polls"

	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		// Create `polls` + `poll_votes` tables.
		for _, model := range []any{
			&gtsmodel.Poll{},
			&gtsmodel.PollVote{},
		} {
			_, err := db.NewCreateTable().
				IfNotExists().
				Model(model).
				Exec(ctx)
			if err != nil && !(strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate column name") || strings.Contains(err.Error(), "SQLSTATE 42701")) {
				return err
			}
		}

		// Add the new status `poll_id` column.
		_, err := db.NewAddColumn().
			Model(&gtsmodel.Status{}).
			ColumnExpr("? CHAR(26)", bun.Ident("poll_id")).
			Exec(ctx)
		if err != nil && !(strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate column name") || strings.Contains(err.Error(), "SQLSTATE 42701")) {
			return err
		}

		return nil
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
