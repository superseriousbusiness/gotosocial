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

	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Previous versions of 20240620074530_interaction_policy.go
			// didn't set NOT NULL on gtsmodel.Status.PendingApproval and
			// gtsmodel.StatusFave.PendingApproval, resulting in NULL being
			// set for that column for some statuses. Correct for this.

			log.Info(ctx, "correcting pending_approval on statuses table...")
			res, err := tx.
				NewUpdate().
				Table("statuses").
				Set("? = ?", bun.Ident("pending_approval"), false).
				Where("? IS NULL", bun.Ident("pending_approval")).
				Exec(ctx)
			if err != nil {
				return err
			}

			rows, err := res.RowsAffected()
			if err == nil {
				log.Infof(ctx, "corrected %d entries", rows)
			}

			log.Info(ctx, "correcting pending_approval on status_faves table...")
			res, err = tx.
				NewUpdate().
				Table("status_faves").
				Set("? = ?", bun.Ident("pending_approval"), false).
				Where("? IS NULL", bun.Ident("pending_approval")).
				Exec(ctx)
			if err != nil {
				return err
			}

			rows, err = res.RowsAffected()
			if err == nil {
				log.Infof(ctx, "corrected %d entries", rows)
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
