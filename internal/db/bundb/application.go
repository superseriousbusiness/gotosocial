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
	"github.com/uptrace/bun"
)

type applicationDB struct {
	db    *WrappedDB
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
	return a.state.Caches.GTS.Application().Load(lookup, func() (*gtsmodel.Application, error) {
		var app gtsmodel.Application

		// Not cached! Perform database query.
		if err := dbQuery(&app); err != nil {
			return nil, a.db.ProcessError(err)
		}

		return &app, nil
	}, keyParts...)
}

func (a *applicationDB) PutApplication(ctx context.Context, app *gtsmodel.Application) error {
	return a.state.Caches.GTS.Application().Store(app, func() error {
		_, err := a.db.NewInsert().Model(app).Exec(ctx)
		return a.db.ProcessError(err)
	})
}

func (a *applicationDB) DeleteApplicationByClientID(ctx context.Context, clientID string) error {
	// Attempt to delete application.
	if _, err := a.db.NewDelete().
		Table("applications").
		Where("? = ?", bun.Ident("client_id"), clientID).
		Exec(ctx); err != nil {
		return a.db.ProcessError(err)
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
	a.state.Caches.GTS.Application().Invalidate("ClientID", clientID)

	return nil
}
