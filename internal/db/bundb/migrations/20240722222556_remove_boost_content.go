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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/uptrace/bun"
)

// Boosts previously duplicated some columns from their targets.
// This isn't necessary, so we remove them here.
// Admins may want to vacuum after running this migration.
func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		log.Info(
			ctx,
			"dropping duplicated status boost data, please wait; "+
				"this may take a long time if your database has lots of statuses, don't interrupt it!",
		)
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			_, err := tx.NewUpdate().
				Model((*gtsmodel.Status)(nil)).
				Where("boost_of_id IS NOT NULL").
				SetColumn("content", "NULL").
				SetColumn("content_warning", "NULL").
				SetColumn("text", "NULL").
				SetColumn("language", "NULL").
				SetColumn("sensitive", "FALSE").
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
