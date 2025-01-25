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

	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, "ALTER TABLE ? ADD COLUMN ? TEXT", bun.Ident("instances"), bun.Ident("custom_css"))
		if err != nil && !(strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate column name") || strings.Contains(err.Error(), "SQLSTATE 42701")) {
			return err
		}
		return nil
	}

	down := func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, "ALTER TABLE ? DROP COLUMN ?", bun.Ident("instances"), bun.Ident("custom_css"))
		return err
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
