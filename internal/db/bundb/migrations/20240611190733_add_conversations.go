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

	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

// Note: this migration has an advanced migration followup.
// See Conversations.MigrateDMs().
func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			for _, model := range []interface{}{
				&gtsmodel.Conversation{},
				&gtsmodel.ConversationToStatus{},
			} {
				if _, err := tx.
					NewCreateTable().
					Model(model).
					IfNotExists().
					Exec(ctx); err != nil {
					return err
				}
			}

			// Add indexes to the conversations table.
			for index, columns := range map[string][]string{
				"conversations_account_id_idx": {
					"account_id",
				},
				"conversations_last_status_id_idx": {
					"last_status_id",
				},
			} {
				if _, err := tx.
					NewCreateIndex().
					Model(&gtsmodel.Conversation{}).
					Index(index).
					Column(columns...).
					IfNotExists().
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
