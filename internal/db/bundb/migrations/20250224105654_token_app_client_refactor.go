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
	newmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20250224105654_token_app_client_refactor"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			// Drop unused clients table.
			if _, err := tx.
				NewDropTable().
				Table("clients").
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			// Select all old model
			// applications into memory.
			oldApps := []*oldmodel.Application{}
			if err := tx.
				NewSelect().
				Model(&oldApps).
				Scan(ctx); err != nil {
				return err
			}

			// Drop the old applications table.
			if _, err := tx.
				NewDropTable().
				Table("applications").
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			// Create the new applications table.
			if _, err := tx.
				NewCreateTable().
				Model((*newmodel.Application)(nil)).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Add indexes to new applications table.
			if _, err := tx.
				NewCreateIndex().
				Table("applications").
				Index("applications_client_id_idx").
				Column("client_id").
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Table("applications").
				Index("applications_managed_by_user_id_idx").
				Column("managed_by_user_id").
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			if len(oldApps) != 0 {
				// Convert all the old model applications into new ones.
				newApps := make([]*newmodel.Application, 0, len(oldApps))
				for _, oldApp := range oldApps {
					newApps = append(newApps, &newmodel.Application{
						ID:           id.NewULIDFromTime(oldApp.CreatedAt),
						Name:         oldApp.Name,
						Website:      oldApp.Website,
						RedirectURIs: []string{oldApp.RedirectURI},
						ClientID:     oldApp.ClientID,
						ClientSecret: oldApp.ClientSecret,
						Scopes:       oldApp.Scopes,
					})
				}

				// Whack all the new apps in
				// there. Lads lads lads lads!
				if _, err := tx.
					NewInsert().
					Model(&newApps).
					Exec(ctx); err != nil {
					return err
				}
			}

			// Select all the old model
			// tokens into memory.
			oldTokens := []*oldmodel.Token{}
			if err := tx.
				NewSelect().
				Model(&oldTokens).
				Scan(ctx); err != nil {
				return err
			}

			// Drop the old token table.
			if _, err := tx.
				NewDropTable().
				Table("tokens").
				IfExists().
				Exec(ctx); err != nil {
				return err
			}

			// Create the new token table.
			if _, err := tx.
				NewCreateTable().
				Model((*newmodel.Token)(nil)).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Add access index to new token table.
			if _, err := tx.
				NewCreateIndex().
				Table("tokens").
				Index("tokens_access_idx").
				Column("access").
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			if len(oldTokens) != 0 {
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
