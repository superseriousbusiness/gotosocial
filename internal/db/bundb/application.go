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
	"errors"
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
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

func (a *applicationDB) GetApplicationsManagedByUserID(
	ctx context.Context,
	userID string,
	page *paging.Page,
) ([]*gtsmodel.Application, error) {
	var (
		// Get paging params.
		minID = page.GetMin()
		maxID = page.GetMax()
		limit = page.GetLimit()
		order = page.GetOrder()

		// Make educated guess for slice size.
		appIDs = make([]string, 0, limit)
	)

	// Ensure user ID.
	if userID == "" {
		return nil, gtserror.New("userID not set")
	}

	q := a.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("applications"), bun.Ident("application")).
		Column("application.id").
		Where("? = ?", bun.Ident("application.managed_by_user_id"), userID)

	if maxID != "" {
		// Return only apps LOWER (ie., older) than maxID.
		q = q.Where("? < ?", bun.Ident("application.id"), maxID)
	}

	if minID != "" {
		// Return only apps HIGHER (ie., newer) than minID.
		q = q.Where("? > ?", bun.Ident("application.id"), minID)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	if order == paging.OrderAscending {
		// Page up.
		q = q.Order("application.id ASC")
	} else {
		// Page down.
		q = q.Order("application.id DESC")
	}

	if err := q.Scan(ctx, &appIDs); err != nil {
		return nil, err
	}

	if len(appIDs) == 0 {
		return nil, nil
	}

	// If we're paging up, we still want apps
	// to be sorted by ID desc (ie., newest to
	// oldest), so reverse ids slice.
	if order == paging.OrderAscending {
		slices.Reverse(appIDs)
	}

	return a.getApplicationsByIDs(ctx, appIDs)
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

func (a *applicationDB) getApplicationsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Application, error) {
	apps, err := a.state.Caches.DB.Application.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.Application, error) {
			// Preallocate expected length of uncached apps.
			apps := make([]*gtsmodel.Application, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) app IDs.
			if err := a.db.NewSelect().
				Model(&apps).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return apps, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the apps by their
	// IDs to ensure in correct order.
	getID := func(t *gtsmodel.Application) string { return t.ID }
	xslices.OrderBy(apps, ids, getID)

	return apps, nil
}

func (a *applicationDB) PutApplication(ctx context.Context, app *gtsmodel.Application) error {
	return a.state.Caches.DB.Application.Store(app, func() error {
		_, err := a.db.NewInsert().Model(app).Exec(ctx)
		return err
	})
}

// DeleteApplicationByID deletes application with the given ID.
//
// The function does not delete tokens owned by the application
// or update statuses/accounts that used the application, since
// the latter can be extremely expensive given the size of the
// statuses table.
//
// Callers to this function should ensure that they do side
// effects themselves (if required) before or after calling.
func (a *applicationDB) DeleteApplicationByID(ctx context.Context, id string) error {
	_, err := a.db.NewDelete().
		Table("applications").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx)
	if err != nil {
		return err
	}

	a.state.Caches.DB.Application.Invalidate("ID", id)
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

func (a *applicationDB) GetAccessTokens(
	ctx context.Context,
	userID string,
	page *paging.Page,
) ([]*gtsmodel.Token, error) {
	var (
		// Get paging params.
		minID = page.GetMin()
		maxID = page.GetMax()
		limit = page.GetLimit()
		order = page.GetOrder()

		// Make educated guess for slice size.
		tokenIDs = make([]string, 0, limit)
	)

	// Ensure user ID.
	if userID == "" {
		return nil, gtserror.New("userID not set")
	}

	q := a.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("tokens"), bun.Ident("token")).
		Column("token.id").
		Where("? = ?", bun.Ident("token.user_id"), userID).
		Where("? != ?", bun.Ident("token.access"), "")

	if maxID != "" {
		// Return only tokens LOWER (ie., older) than maxID.
		q = q.Where("? < ?", bun.Ident("token.id"), maxID)
	}

	if minID != "" {
		// Return only tokens HIGHER (ie., newer) than minID.
		q = q.Where("? > ?", bun.Ident("token.id"), minID)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	if order == paging.OrderAscending {
		// Page up.
		q = q.Order("token.id ASC")
	} else {
		// Page down.
		q = q.Order("token.id DESC")
	}

	if err := q.Scan(ctx, &tokenIDs); err != nil {
		return nil, err
	}

	if len(tokenIDs) == 0 {
		return nil, nil
	}

	// If we're paging up, we still want tokens
	// to be sorted by ID desc (ie., newest to
	// oldest), so reverse ids slice.
	if order == paging.OrderAscending {
		slices.Reverse(tokenIDs)
	}

	return a.getTokensByIDs(ctx, tokenIDs)
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

func (a *applicationDB) getTokensByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Token, error) {
	tokens, err := a.state.Caches.DB.Token.LoadIDs("ID",
		ids,
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

	// Reorder the tokens by their
	// IDs to ensure in correct order.
	getID := func(t *gtsmodel.Token) string { return t.ID }
	xslices.OrderBy(tokens, ids, getID)

	return tokens, nil
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

func (a *applicationDB) UpdateToken(ctx context.Context, token *gtsmodel.Token, columns ...string) error {
	return a.state.Caches.DB.Token.Store(token, func() error {
		_, err := a.db.
			NewUpdate().
			Model(token).
			Column(columns...).
			Where("? = ?", bun.Ident("id"), token.ID).
			Exec(ctx)
		return err
	})
}

func (a *applicationDB) DeleteTokenByID(ctx context.Context, id string) error {
	var token gtsmodel.Token
	token.ID = id

	_, err := a.db.NewDelete().
		Model(&token).
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx)
	if err != nil {
		return err
	}

	a.state.Caches.DB.Token.Invalidate("ID", id)
	a.state.Caches.OnInvalidateToken(&token)
	return nil
}

func (a *applicationDB) DeleteTokenByCode(ctx context.Context, code string) error {
	var token gtsmodel.Token

	_, err := a.db.NewDelete().
		Model(&token).
		Where("? = ?", bun.Ident("code"), code).
		Returning("?", bun.Ident("id")).
		Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	a.state.Caches.DB.Token.Invalidate("Code", code)
	a.state.Caches.OnInvalidateToken(&token)
	return nil
}

func (a *applicationDB) DeleteTokenByAccess(ctx context.Context, access string) error {
	var token gtsmodel.Token

	_, err := a.db.NewDelete().
		Model(&token).
		Where("? = ?", bun.Ident("access"), access).
		Returning("?", bun.Ident("id")).
		Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	a.state.Caches.DB.Token.Invalidate("Access", access)
	a.state.Caches.OnInvalidateToken(&token)
	return nil
}

func (a *applicationDB) DeleteTokenByRefresh(ctx context.Context, refresh string) error {
	var token gtsmodel.Token

	_, err := a.db.NewDelete().
		Model(&token).
		Where("? = ?", bun.Ident("refresh"), refresh).
		Returning("?", bun.Ident("id")).
		Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	a.state.Caches.DB.Token.Invalidate("Refresh", refresh)
	a.state.Caches.OnInvalidateToken(&token)
	return nil
}

func (a *applicationDB) DeleteTokensByClientID(ctx context.Context, clientID string) error {
	var tokens []*gtsmodel.Token

	// Delete tokens owned by
	// clientID and gather token IDs.
	if _, err := a.db.NewDelete().
		Model(&tokens).
		Where("? = ?", bun.Ident("client_id"), clientID).
		Returning("?", bun.Ident("id")).
		Exec(ctx); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate all deleted tokens.
	for _, token := range tokens {
		a.state.Caches.DB.Token.Invalidate("ID", token.ID)
		a.state.Caches.OnInvalidateToken(token)
	}

	return nil
}
