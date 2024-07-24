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
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

type tagDB struct {
	db    *bun.DB
	state *state.State
}

func (t *tagDB) GetTag(ctx context.Context, id string) (*gtsmodel.Tag, error) {
	return t.state.Caches.DB.Tag.LoadOne("ID", func() (*gtsmodel.Tag, error) {
		var tag gtsmodel.Tag

		q := t.db.
			NewSelect().
			Model(&tag).
			Where("? = ?", bun.Ident("tag.id"), id)

		if err := q.Scan(ctx); err != nil {
			return nil, err
		}

		return &tag, nil
	}, id)
}

func (t *tagDB) GetTagByName(ctx context.Context, name string) (*gtsmodel.Tag, error) {
	// Normalize 'name' string.
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)

	return t.state.Caches.DB.Tag.LoadOne("Name", func() (*gtsmodel.Tag, error) {
		var tag gtsmodel.Tag

		q := t.db.
			NewSelect().
			Model(&tag).
			Where("? = ?", bun.Ident("tag.name"), name)

		if err := q.Scan(ctx); err != nil {
			return nil, err
		}

		return &tag, nil
	}, name)
}

func (t *tagDB) GetTags(ctx context.Context, ids []string) ([]*gtsmodel.Tag, error) {
	// Load all tag IDs via cache loader callbacks.
	tags, err := t.state.Caches.DB.Tag.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.Tag, error) {
			// Preallocate expected length of uncached tags.
			tags := make([]*gtsmodel.Tag, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := t.db.NewSelect().
				Model(&tags).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return tags, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the tags by their
	// IDs to ensure in correct order.
	getID := func(t *gtsmodel.Tag) string { return t.ID }
	util.OrderBy(tags, ids, getID)

	return tags, nil
}

func (t *tagDB) PutTag(ctx context.Context, tag *gtsmodel.Tag) error {
	// Normalize 'name' string before it enters
	// the db, without changing tag we were given.
	//
	// First copy tag to new pointer.
	t2 := new(gtsmodel.Tag)
	*t2 = *tag

	// Normalize name on new pointer.
	t2.Name = strings.TrimSpace(t2.Name)
	t2.Name = strings.ToLower(t2.Name)

	// Insert the copy.
	if err := t.state.Caches.DB.Tag.Store(t2, func() error {
		_, err := t.db.NewInsert().Model(t2).Exec(ctx)
		return err
	}); err != nil {
		return err // err already processed
	}

	// Update original tag with
	// field values populated by db.
	tag.CreatedAt = t2.CreatedAt
	tag.UpdatedAt = t2.UpdatedAt
	tag.Useable = t2.Useable
	tag.Listable = t2.Listable

	return nil
}
