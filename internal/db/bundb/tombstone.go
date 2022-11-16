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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"

	"codeberg.org/gruf/go-cache/v3/result"
)

type tombstoneDB struct {
	conn  *DBConn
	cache *result.Cache[*gtsmodel.Tombstone]
}

func (t *tombstoneDB) init() {
	// Initialize tombstone result cache
	t.cache = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
	}, func(t1 *gtsmodel.Tombstone) *gtsmodel.Tombstone {
		t2 := new(gtsmodel.Tombstone)
		*t2 = *t1
		return t2
	}, 100)

	// Set cache TTL and start sweep routine
	t.cache.SetTTL(time.Minute*5, false)
	t.cache.Start(time.Second * 10)
}

func (t *tombstoneDB) GetTombstoneByURI(ctx context.Context, uri string) (*gtsmodel.Tombstone, db.Error) {
	return t.cache.Load("URI", func() (*gtsmodel.Tombstone, error) {
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
	return t.cache.Store(tombstone, func() error {
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
	t.cache.Invalidate("ID", id)

	return nil
}
