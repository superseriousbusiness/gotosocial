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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx,
			"ALTER TABLE ? ADD COLUMN ? CHAR(26)",
			bun.Ident("accounts"), bun.Ident("move_id"),
		)
		if err != nil {
			e := err.Error()
			if !(strings.Contains(e, "already exists") ||
				strings.Contains(e, "duplicate column name") ||
				strings.Contains(e, "SQLSTATE 42701")) {
				return err
			}
		}

		// Create "moves" table.
		if _, err := db.NewCreateTable().
			IfNotExists().
			Model(&gtsmodel.Move{}).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return nil
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
