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
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type mentionDB struct {
	conn  *DBConn
	state *state.State
}

func (m *mentionDB) GetMention(ctx context.Context, id string) (*gtsmodel.Mention, db.Error) {
	mention, err := m.state.Caches.GTS.Mention().Load("ID", func() (*gtsmodel.Mention, error) {
		var mention gtsmodel.Mention

		q := m.conn.
			NewSelect().
			Model(&mention).
			Where("? = ?", bun.Ident("mention.id"), id)

		if err := q.Scan(ctx); err != nil {
			return nil, m.conn.ProcessError(err)
		}

		return &mention, nil
	}, id)
	if err != nil {
		return nil, err
	}

	// Set the mention originating status.
	mention.Status, err = m.state.DB.GetStatusByID(
		gtscontext.SetBarebones(ctx),
		mention.StatusID,
	)
	if err != nil {
		return nil, fmt.Errorf("error populating mention status: %w", err)
	}

	// Set the mention origin account model.
	mention.OriginAccount, err = m.state.DB.GetAccountByID(
		gtscontext.SetBarebones(ctx),
		mention.OriginAccountID,
	)
	if err != nil {
		return nil, fmt.Errorf("error populating mention origin account: %w", err)
	}

	// Set the mention target account model.
	mention.TargetAccount, err = m.state.DB.GetAccountByID(
		gtscontext.SetBarebones(ctx),
		mention.TargetAccountID,
	)
	if err != nil {
		return nil, fmt.Errorf("error populating mention target account: %w", err)
	}

	return mention, nil
}

func (m *mentionDB) GetMentions(ctx context.Context, ids []string) ([]*gtsmodel.Mention, db.Error) {
	mentions := make([]*gtsmodel.Mention, 0, len(ids))

	for _, id := range ids {
		// Attempt fetch from DB
		mention, err := m.GetMention(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting mention %q: %v", id, err)
			continue
		}

		// Append mention
		mentions = append(mentions, mention)
	}

	return mentions, nil
}

func (m *mentionDB) PutMention(ctx context.Context, mention *gtsmodel.Mention) error {
	return m.state.Caches.GTS.Mention().Store(mention, func() error {
		_, err := m.conn.NewInsert().Model(mention).Exec(ctx)
		return m.conn.ProcessError(err)
	})
}

func (m *mentionDB) DeleteMentionByID(ctx context.Context, id string) error {
	if _, err := m.conn.
		NewDelete().
		Table("mentions").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil {
		return m.conn.ProcessError(err)
	}

	// Invalidate mention from the lookup cache.
	m.state.Caches.GTS.Mention().Invalidate("ID", id)

	return nil
}
