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
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type tagDB struct {
	conn  *DBConn
	state *state.State
}

func (m *tagDB) GetTag(ctx context.Context, id string) (*gtsmodel.Tag, db.Error) {
	return m.state.Caches.GTS.Tag().Load("ID", func() (*gtsmodel.Tag, error) {
		var tag gtsmodel.Tag

		q := m.conn.
			NewSelect().
			Model(&tag).
			Where("? = ?", bun.Ident("tag.id"), id)

		if err := q.Scan(ctx); err != nil {
			return nil, m.conn.ProcessError(err)
		}

		return &tag, nil
	}, id)
}

func (m *tagDB) GetTagByName(ctx context.Context, name string) (*gtsmodel.Tag, db.Error) {
	// Normalize 'name' string.
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)

	return m.state.Caches.GTS.Tag().Load("Name", func() (*gtsmodel.Tag, error) {
		var tag gtsmodel.Tag

		q := m.conn.
			NewSelect().
			Model(&tag).
			Where("? = ?", bun.Ident("tag.name"), name)

		if err := q.Scan(ctx); err != nil {
			return nil, m.conn.ProcessError(err)
		}

		return &tag, nil
	}, name)
}

func (m *tagDB) GetOrCreateTag(ctx context.Context, name string) (*gtsmodel.Tag, db.Error) {
	tag, err := m.GetTagByName(ctx, name)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Real error.
		return nil, err
	}

	if tag != nil {
		// Tag existed, return it.
		return tag, nil
	}

	// Tag did not exist, create it.
	tag = &gtsmodel.Tag{
		ID:   id.NewULID(),
		Name: name,
	}

	return tag, m.putTag(ctx, tag)
}

func (m *tagDB) GetTags(ctx context.Context, ids []string) ([]*gtsmodel.Tag, db.Error) {
	tags := make([]*gtsmodel.Tag, 0, len(ids))

	for _, id := range ids {
		// Attempt fetch from DB
		tag, err := m.GetTag(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting tag %q: %v", id, err)
			continue
		}

		// Append tag
		tags = append(tags, tag)
	}

	return tags, nil
}

func (m *tagDB) putTag(ctx context.Context, tag *gtsmodel.Tag) error {
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
	if err := m.state.Caches.GTS.Tag().Store(t2, func() error {
		_, err := m.conn.NewInsert().Model(t2).Exec(ctx)
		return m.conn.ProcessError(err)
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
