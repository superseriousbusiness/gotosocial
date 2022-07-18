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

	"codeberg.org/gruf/go-cache/v2"
	log "github.com/dsoprea/go-logging"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type mentionDB struct {
	conn  *DBConn
	cache cache.Cache[string, *gtsmodel.Mention]
}

func (m *mentionDB) newMentionQ(i interface{}) *bun.SelectQuery {
	return m.conn.
		NewSelect().
		Model(i).
		Relation("Status").
		Relation("OriginAccount").
		Relation("TargetAccount")
}

func (m *mentionDB) getMentionDB(ctx context.Context, id string) (*gtsmodel.Mention, db.Error) {
	mention := gtsmodel.Mention{}

	q := m.newMentionQ(&mention).
		Where("mention.id = ?", id)

	if err := q.Scan(ctx); err != nil {
		return nil, m.conn.ProcessError(err)
	}

	copy := mention
	m.cache.Set(mention.ID, &copy)

	return &mention, nil
}

func (m *mentionDB) GetMention(ctx context.Context, id string) (*gtsmodel.Mention, db.Error) {
	if mention, ok := m.cache.Get(id); ok {
		return mention, nil
	}
	return m.getMentionDB(ctx, id)
}

func (m *mentionDB) GetMentions(ctx context.Context, ids []string) ([]*gtsmodel.Mention, db.Error) {
	mentions := make([]*gtsmodel.Mention, 0, len(ids))

	for _, id := range ids {
		// Attempt fetch from DB
		mention, err := m.GetMention(ctx, id)
		if err != nil {
			log.Errorf("GetMentions: error getting mention %q: %v", id, err)
			continue
		}

		// Append mention
		mentions = append(mentions, mention)
	}

	return mentions, nil
}
