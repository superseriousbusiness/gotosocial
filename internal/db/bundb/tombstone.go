/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package bundb

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type tombstoneDB struct {
	conn  *DBConn
	state *state.State
}

func (t *tombstoneDB) GetTombstoneByURI(ctx context.Context, uri string) (*gtsmodel.Tombstone, db.Error) {
	return t.state.Caches.GTS.Tombstone().Load("URI", func() (*gtsmodel.Tombstone, error) {
		var tomb gtsmodel.Tombstone

		q := t.conn.
			NewSelect().
			Model(&tomb).
			Where("? = ?", bun.Ident("tombstone.uri"), uri)

		if err := q.Scan(ctx); err != nil {
			return nil, t.conn.ProcessError(err)
		}

		return &tomb, nil
	}, uri)
}

func (t *tombstoneDB) TombstoneExistsWithURI(ctx context.Context, uri string) (bool, db.Error) {
	tomb, err := t.GetTombstoneByURI(ctx, uri)
	if err == db.ErrNoEntries {
		err = nil
	}
	return (tomb != nil), err
}

func (t *tombstoneDB) PutTombstone(ctx context.Context, tombstone *gtsmodel.Tombstone) db.Error {
	return t.state.Caches.GTS.Tombstone().Store(tombstone, func() error {
		_, err := t.conn.
			NewInsert().
			Model(tombstone).
			Exec(ctx)
		return t.conn.ProcessError(err)
	})
}

func (t *tombstoneDB) DeleteTombstone(ctx context.Context, id string) db.Error {
	if _, err := t.conn.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("tombstones"), bun.Ident("tombstone")).
		Where("? = ?", bun.Ident("tombstone.id"), id).
		Exec(ctx); err != nil {
		return t.conn.ProcessError(err)
	}

	// Invalidate from cache by ID
	t.state.Caches.GTS.Tombstone().Invalidate("ID", id)

	return nil
}
