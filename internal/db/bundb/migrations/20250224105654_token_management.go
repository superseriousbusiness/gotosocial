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

	oldmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20211113114307_init"
	newmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20250224105654_token_management"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			// Select all the old model
			// tokens into memory.
			oldTokens := []*oldmodel.Token{}
			if err := tx.
				NewSelect().
				Model(&oldTokens).
				Scan(ctx); err != nil {
				return err
			}

			// Drop the old table.
			if _, err := tx.
				NewDropTable().
				Table("tokens").
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			// Create the new table.
			if _, err := tx.
				NewCreateTable().
				Model((*newmodel.Token)(nil)).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Add access index to new table.
			if _, err := tx.
				NewCreateIndex().
				Table("tokens").
				Index("tokens_access_idx").
				Column("access").
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			if len(oldTokens) == 0 {
				// Nothing left to do.
				return nil
			}

			// Convert all the old model tokens into new ones.
			newTokens := make([]*newmodel.Token, 0, len(oldTokens))
			for _, oldToken := range oldTokens {
				newTokens = append(newTokens, &newmodel.Token{
					ID:                  id.NewULIDFromTime(oldToken.CreatedAt),
					ClientID:            oldToken.ClientID,
					UserID:              oldToken.UserID,
					RedirectURI:         oldToken.RedirectURI,
					Scope:               oldToken.Scope,
					Code:                oldToken.Code,
					CodeChallenge:       oldToken.CodeChallenge,
					CodeChallengeMethod: oldToken.CodeChallengeMethod,
					CodeCreateAt:        oldToken.CodeCreateAt,
					CodeExpiresAt:       oldToken.CodeExpiresAt,
					Access:              oldToken.Access,
					AccessCreateAt:      oldToken.AccessCreateAt,
					AccessExpiresAt:     oldToken.AccessExpiresAt,
					Refresh:             oldToken.Refresh,
					RefreshCreateAt:     oldToken.RefreshCreateAt,
					RefreshExpiresAt:    oldToken.RefreshExpiresAt,
				})
			}

			// Whack all the new tokens in
			// there. Lads lads lads lads!
			if _, err := tx.
				NewInsert().
				Model(&newTokens).
				Exec(ctx); err != nil {
				return err
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
