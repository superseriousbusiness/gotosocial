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

package bundb

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

type applicationDB struct {
	db    *bun.DB
	state *state.State
}

func (a *applicationDB) GetApplicationByID(ctx context.Context, id string) (*gtsmodel.Application, error) {
	return a.getApplication(
		ctx,
		"ID",
		func(app *gtsmodel.Application) error {
			return a.db.NewSelect().Model(app).Where("? = ?", bun.Ident("id"), id).Scan(ctx)
		},
		id,
	)
}

func (a *applicationDB) GetApplicationByClientID(ctx context.Context, clientID string) (*gtsmodel.Application, error) {
	return a.getApplication(
		ctx,
		"ClientID",
		func(app *gtsmodel.Application) error {
			return a.db.NewSelect().Model(app).Where("? = ?", bun.Ident("client_id"), clientID).Scan(ctx)
		},
		clientID,
	)
}

func (a *applicationDB) getApplication(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Application) error, keyParts ...any) (*gtsmodel.Application, error) {
	return a.state.Caches.DB.Application.LoadOne(lookup, func() (*gtsmodel.Application, error) {
		var app gtsmodel.Application

		// Not cached! Perform database query.
		if err := dbQuery(&app); err != nil {
			return nil, err
		}

		return &app, nil
	}, keyParts...)
}

func (a *applicationDB) PutApplication(ctx context.Context, app *gtsmodel.Application) error {
	return a.state.Caches.DB.Application.Store(app, func() error {
		_, err := a.db.NewInsert().Model(app).Exec(ctx)
		return err
	})
}

func (a *applicationDB) DeleteApplicationByClientID(ctx context.Context, clientID string) error {
	// Attempt to delete application.
	if _, err := a.db.NewDelete().
		Table("applications").
		Where("? = ?", bun.Ident("client_id"), clientID).
		Exec(ctx); err != nil {
		return err
	}

	// NOTE about further side effects:
	//
	// We don't need to handle updating any statuses or users
	// (both of which may contain refs to applications), as
	// DeleteApplication__() is only ever called during an
	// account deletion, which handles deletion of the user
	// and all their statuses already.
	//

	// Clear application from the cache.
	a.state.Caches.DB.Application.Invalidate("ClientID", clientID)

	return nil
}

func (a *applicationDB) GetClientByID(ctx context.Context, id string) (*gtsmodel.Client, error) {
	return a.state.Caches.DB.Client.LoadOne("ID", func() (*gtsmodel.Client, error) {
		var client gtsmodel.Client

		if err := a.db.NewSelect().
			Model(&client).
			Where("? = ?", bun.Ident("id"), id).
			Scan(ctx); err != nil {
			return nil, err
		}

		return &client, nil
	}, id)
}

func (a *applicationDB) PutClient(ctx context.Context, client *gtsmodel.Client) error {
	return a.state.Caches.DB.Client.Store(client, func() error {
		_, err := a.db.NewInsert().Model(client).Exec(ctx)
		return err
	})
}

func (a *applicationDB) DeleteClientByID(ctx context.Context, id string) error {
	_, err := a.db.NewDelete().
		Table("clients").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx)
	if err != nil {
		return err
	}

	a.state.Caches.DB.Client.Invalidate("ID", id)
	return nil
}

func (a *applicationDB) GetAllTokens(ctx context.Context) ([]*gtsmodel.Token, error) {
	var tokenIDs []string

	// Select ALL token IDs.
	if err := a.db.NewSelect().
		Table("tokens").
		Column("id").
		Scan(ctx, &tokenIDs); err != nil {
		return nil, err
	}

	// Load all input token IDs via cache loader callback.
	tokens, err := a.state.Caches.DB.Token.LoadIDs("ID",
		tokenIDs,
		func(uncached []string) ([]*gtsmodel.Token, error) {
			// Preallocate expected length of uncached tokens.
			tokens := make([]*gtsmodel.Token, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) token IDs.
			if err := a.db.NewSelect().
				Model(&tokens).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return tokens, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reoroder the tokens by their
	// IDs to ensure in correct order.
	getID := func(t *gtsmodel.Token) string { return t.ID }
	xslices.OrderBy(tokens, tokenIDs, getID)

	return tokens, nil
}

func (a *applicationDB) GetTokenByID(ctx context.Context, code string) (*gtsmodel.Token, error) {
	return a.getTokenBy(
		"ID",
		func(t *gtsmodel.Token) error {
			return a.db.NewSelect().Model(t).Where("? = ?", bun.Ident("id"), code).Scan(ctx)
		},
		code,
	)
}

func (a *applicationDB) GetTokenByCode(ctx context.Context, code string) (*gtsmodel.Token, error) {
	return a.getTokenBy(
		"Code",
		func(t *gtsmodel.Token) error {
			return a.db.NewSelect().Model(t).Where("? = ?", bun.Ident("code"), code).Scan(ctx)
		},
		code,
	)
}

func (a *applicationDB) GetTokenByAccess(ctx context.Context, access string) (*gtsmodel.Token, error) {
	return a.getTokenBy(
		"Access",
		func(t *gtsmodel.Token) error {
			return a.db.NewSelect().Model(t).Where("? = ?", bun.Ident("access"), access).Scan(ctx)
		},
		access,
	)
}

func (a *applicationDB) GetTokenByRefresh(ctx context.Context, refresh string) (*gtsmodel.Token, error) {
	return a.getTokenBy(
		"Refresh",
		func(t *gtsmodel.Token) error {
			return a.db.NewSelect().Model(t).Where("? = ?", bun.Ident("refresh"), refresh).Scan(ctx)
		},
		refresh,
	)
}

func (a *applicationDB) getTokenBy(lookup string, dbQuery func(*gtsmodel.Token) error, keyParts ...any) (*gtsmodel.Token, error) {
	return a.state.Caches.DB.Token.LoadOne(lookup, func() (*gtsmodel.Token, error) {
		var token gtsmodel.Token

		if err := dbQuery(&token); err != nil {
			return nil, err
		}

		return &token, nil
	}, keyParts...)
}

func (a *applicationDB) PutToken(ctx context.Context, token *gtsmodel.Token) error {
	return a.state.Caches.DB.Token.Store(token, func() error {
		_, err := a.db.NewInsert().Model(token).Exec(ctx)
		return err
	})
}

func (a *applicationDB) DeleteTokenByID(ctx context.Context, id string) error {
	_, err := a.db.NewDelete().
		Table("tokens").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx)
	if err != nil {
		return err
	}

	a.state.Caches.DB.Token.Invalidate("ID", id)
	return nil
}

func (a *applicationDB) DeleteTokenByCode(ctx context.Context, code string) error {
	_, err := a.db.NewDelete().
		Table("tokens").
		Where("? = ?", bun.Ident("code"), code).
		Exec(ctx)
	if err != nil {
		return err
	}

	a.state.Caches.DB.Token.Invalidate("Code", code)
	return nil
}

func (a *applicationDB) DeleteTokenByAccess(ctx context.Context, access string) error {
	_, err := a.db.NewDelete().
		Table("tokens").
		Where("? = ?", bun.Ident("access"), access).
		Exec(ctx)
	if err != nil {
		return err
	}

	a.state.Caches.DB.Token.Invalidate("Access", access)
	return nil
}

func (a *applicationDB) DeleteTokenByRefresh(ctx context.Context, refresh string) error {
	_, err := a.db.NewDelete().
		Table("tokens").
		Where("? = ?", bun.Ident("refresh"), refresh).
		Exec(ctx)
	if err != nil {
		return err
	}

	a.state.Caches.DB.Token.Invalidate("Refresh", refresh)
	return nil
}
