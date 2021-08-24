/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package pg

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type mentionDB struct {
	config *config.Config
	conn   *bun.DB
	log    *logrus.Logger
	cancel context.CancelFunc
	cache  cache.Cache
}

func (m *mentionDB) cacheMention(id string, mention *gtsmodel.Mention) {
	if m.cache == nil {
		m.cache = cache.New()
	}

	if err := m.cache.Store(id, mention); err != nil {
		m.log.Panicf("mentionDB: error storing in cache: %s", err)
	}
}

func (m *mentionDB) mentionCached(id string) (*gtsmodel.Mention, bool) {
	if m.cache == nil {
		m.cache = cache.New()
		return nil, false
	}

	mI, err := m.cache.Fetch(id)
	if err != nil || mI == nil {
		return nil, false
	}

	mention, ok := mI.(*gtsmodel.Mention)
	if !ok {
		m.log.Panicf("mentionDB: cached interface with key %s was not a mention", id)
	}

	return mention, true
}

func (m *mentionDB) newMentionQ(i interface{}) *bun.SelectQuery {
	return m.conn.
		NewSelect().
		Model(i).
		Relation("Status").
		Relation("OriginAccount").
		Relation("TargetAccount")
}

func (m *mentionDB) GetMention(ctx context.Context, id string) (*gtsmodel.Mention, db.Error) {
	if mention, cached := m.mentionCached(id); cached {
		return mention, nil
	}

	mention := &gtsmodel.Mention{}

	q := m.newMentionQ(mention).
		Where("mention.id = ?", id)

	err := processErrorResponse(q.Scan(ctx))

	if err == nil && mention != nil {
		m.cacheMention(id, mention)
	}

	return mention, err
}

func (m *mentionDB) GetMentions(ctx context.Context, ids []string) ([]*gtsmodel.Mention, db.Error) {
	mentions := []*gtsmodel.Mention{}

	for _, i := range ids {
		mention, err := m.GetMention(ctx, i)
		if err != nil {
			return nil, processErrorResponse(err)
		}
		mentions = append(mentions, mention)
	}

	return mentions, nil
}
