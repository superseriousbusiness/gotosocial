/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type tombstoneDB struct {
	conn  *DBConn
	cache *cache.TombstoneCache
}

func (t *tombstoneDB) TombstoneExists(ctx context.Context, uri string) (bool, db.Error) {
	if _, err := t.getTombstone(
		ctx,
		func() (*gtsmodel.Tombstone, bool) {
			return t.cache.GetByURI(uri)
		},
		func(status *gtsmodel.Tombstone) error {
			tombstone := &gtsmodel.Tombstone{}
			return t.conn.
				NewSelect().
				Model(tombstone).
				Where("? = ?", bun.Ident("tombstone.uri"), uri).
				Scan(ctx)
		},
	); err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// doesn't exist
			return false, nil
		}
		// there's a real error
		return false, err
	}

	return true, nil
}

func (t *tombstoneDB) PutTombstone(ctx context.Context, tombstone *gtsmodel.Tombstone) (*gtsmodel.Tombstone, db.Error) {
	if _, err := t.conn.
		NewInsert().
		Model(tombstone).
		Exec(ctx); err != nil {
		return nil, t.conn.ProcessError(err)
	}

	t.cache.Put(tombstone)
	return tombstone, nil
}

func (t *tombstoneDB) DeleteTombstone(ctx context.Context, id string) db.Error {
	if _, err := t.conn.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("tombstones"), bun.Ident("tombstone")).
		Where("? = ?", bun.Ident("tombstone.id"), id).
		Exec(ctx); err != nil {
		return t.conn.ProcessError(err)
	}

	t.cache.Invalidate(id)
	return nil
}

func (t *tombstoneDB) getTombstone(ctx context.Context, cacheGet func() (*gtsmodel.Tombstone, bool), dbQuery func(*gtsmodel.Tombstone) error) (*gtsmodel.Tombstone, db.Error) {
	// Attempt to fetch cached tombstone
	tombstone, cached := cacheGet()

	if !cached {
		tombstone = &gtsmodel.Tombstone{}

		// Not cached! Perform database query
		if err := dbQuery(tombstone); err != nil {
			return nil, t.conn.ProcessError(err)
		}

		// Place in the cache
		t.cache.Put(tombstone)
	}

	return tombstone, nil
}
