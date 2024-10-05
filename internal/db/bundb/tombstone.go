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

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type tombstoneDB struct {
	db    *bun.DB
	state *state.State
}

func (t *tombstoneDB) GetTombstoneByURI(ctx context.Context, uri string) (*gtsmodel.Tombstone, error) {
	return t.state.Caches.DB.Tombstone.LoadOne("URI", func() (*gtsmodel.Tombstone, error) {
		var tomb gtsmodel.Tombstone

		q := t.db.
			NewSelect().
			Model(&tomb).
			Where("? = ?", bun.Ident("tombstone.uri"), uri)

		if err := q.Scan(ctx); err != nil {
			return nil, err
		}

		return &tomb, nil
	}, uri)
}

func (t *tombstoneDB) TombstoneExistsWithURI(ctx context.Context, uri string) (bool, error) {
	tomb, err := t.GetTombstoneByURI(ctx, uri)
	if err == db.ErrNoEntries {
		err = nil
	}
	return (tomb != nil), err
}

func (t *tombstoneDB) PutTombstone(ctx context.Context, tombstone *gtsmodel.Tombstone) error {
	return t.state.Caches.DB.Tombstone.Store(tombstone, func() error {
		_, err := t.db.
			NewInsert().
			Model(tombstone).
			Exec(ctx)
		return err
	})
}

func (t *tombstoneDB) DeleteTombstone(ctx context.Context, id string) error {
	// Delete tombstone from DB.
	_, err := t.db.NewDelete().
		TableExpr("? AS ?", bun.Ident("tombstones"), bun.Ident("tombstone")).
		Where("? = ?", bun.Ident("tombstone.id"), id).
		Exec(ctx)

	// Invalidate any cached tombstone by given ID.
	t.state.Caches.DB.Tombstone.Invalidate("ID", id)

	return err
}
