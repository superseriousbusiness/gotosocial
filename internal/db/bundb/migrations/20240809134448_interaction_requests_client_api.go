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

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"

	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			// Drop interaction approvals table if it exists,
			// ie., if instance was running on main between now
			// and 2024-07-16.
			//
			// We might lose some interaction approvals this way,
			// but since they weren't *really* used much yet this
			// it's not a big deal, that's the running-on-main life!
			if _, err := tx.NewDropTable().
				Table("interaction_approvals").
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			// Add `interaction_requests`
			// table and new indexes.
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.InteractionRequest{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			for idx, col := range map[string]string{
				"interaction_requests_status_id_idx":              "status_id",
				"interaction_requests_target_account_id_idx":      "target_account_id",
				"interaction_requests_interacting_account_id_idx": "interacting_account_id",
			} {
				if _, err := tx.
					NewCreateIndex().
					Table("interaction_requests").
					Index(idx).
					Column(col).
					IfNotExists().
					Exec(ctx); err != nil {
					return err
				}
			}

			// Select all pending statuses (replies or boosts).
			pendingStatuses := []*gtsmodel.Status{}
			err := tx.
				NewSelect().
				Model(&pendingStatuses).
				Column(
					"created_at",
					"in_reply_to_id",
					"boost_of_id",
					"in_reply_to_account_id",
					"boost_of_account_id",
					"account_id",
					"uri",
				).
				Where("? = ?", bun.Ident("pending_approval"), true).
				Scan(ctx)
			if err != nil {
				return err
			}

			// For each currently pending status, check whether it's a reply or
			// a boost, and insert a corresponding interaction request into the db.
			for _, pendingStatus := range pendingStatuses {
				req := typeutils.StatusToInteractionRequest(pendingStatus)
				if _, err := tx.
					NewInsert().
					Model(req).
					Exec(ctx); err != nil {
					return err
				}
			}

			// Now do the same thing for pending faves.
			pendingFaves := []*gtsmodel.StatusFave{}
			err = tx.
				NewSelect().
				Model(&pendingFaves).
				Column(
					"created_at",
					"status_id",
					"target_account_id",
					"account_id",
					"uri",
				).
				Where("? = ?", bun.Ident("pending_approval"), true).
				Scan(ctx)
			if err != nil {
				return err
			}

			for _, pendingFave := range pendingFaves {
				req := typeutils.StatusFaveToInteractionRequest(pendingFave)

				if _, err := tx.
					NewInsert().
					Model(req).
					Exec(ctx); err != nil {
					return err
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
