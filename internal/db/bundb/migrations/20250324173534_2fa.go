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
	"fmt"
	"reflect"

	newmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20250324173534_2fa"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			log.Info(ctx, "adding new 2fa columns to user table...")

			var newUser *newmodel.User
			newUserType := reflect.TypeOf(newUser)

			for _, column := range []string{
				"TwoFactorSecret",
				"TwoFactorBackups",
				"TwoFactorEnabledAt",
			} {
				// Generate new column definition from bun.
				colDef, err := getBunColumnDef(tx, newUserType, column)
				if err != nil {
					return fmt.Errorf("error making column def: %w", err)
				}

				_, err = tx.
					NewAddColumn().
					Model(newUser).
					ColumnExpr(colDef).
					Exec(ctx)
				if err != nil {
					return fmt.Errorf("error adding column: %w", err)
				}
			}

			return nil
		})
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return nil
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
