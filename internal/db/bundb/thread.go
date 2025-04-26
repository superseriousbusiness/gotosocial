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

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type threadDB struct {
	db    *bun.DB
	state *state.State
}

func (t *threadDB) PutThread(ctx context.Context, thread *gtsmodel.Thread) error {
	_, err := t.db.
		NewInsert().
		Model(thread).
		Exec(ctx)

	return err
}

func (t *threadDB) GetThreadMute(ctx context.Context, id string) (*gtsmodel.ThreadMute, error) {
	return t.state.Caches.DB.ThreadMute.LoadOne("ID", func() (*gtsmodel.ThreadMute, error) {
		var threadMute gtsmodel.ThreadMute

		q := t.db.
			NewSelect().
			Model(&threadMute).
			Where("? = ?", bun.Ident("thread_mute.id"), id)

		if err := q.Scan(ctx); err != nil {
			return nil, err
		}

		return &threadMute, nil
	}, id)
}

func (t *threadDB) GetThreadMutedByAccount(
	ctx context.Context,
	threadID string,
	accountID string,
) (*gtsmodel.ThreadMute, error) {
	return t.state.Caches.DB.ThreadMute.LoadOne("ThreadID,AccountID", func() (*gtsmodel.ThreadMute, error) {
		var threadMute gtsmodel.ThreadMute

		q := t.db.
			NewSelect().
			Model(&threadMute).
			Where("? = ?", bun.Ident("thread_mute.thread_id"), threadID).
			Where("? = ?", bun.Ident("thread_mute.account_id"), accountID)

		if err := q.Scan(ctx); err != nil {
			return nil, err
		}

		return &threadMute, nil
	}, threadID, accountID)
}

func (t *threadDB) IsThreadMutedByAccount(
	ctx context.Context,
	threadID string,
	accountID string,
) (bool, error) {
	if threadID == "" {
		return false, nil
	}

	mute, err := t.GetThreadMutedByAccount(ctx, threadID, accountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return false, err
	}

	return (mute != nil), nil
}

func (t *threadDB) PutThreadMute(ctx context.Context, threadMute *gtsmodel.ThreadMute) error {
	return t.state.Caches.DB.ThreadMute.Store(threadMute, func() error {
		_, err := t.db.NewInsert().Model(threadMute).Exec(ctx)
		return err
	})
}

func (t *threadDB) DeleteThreadMute(ctx context.Context, id string) error {
	if _, err := t.db.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("thread_mutes"), bun.Ident("thread_mute")).
		Where("? = ?", bun.Ident("thread_mute.id"), id).Exec(ctx); err != nil {
		return err
	}

	t.state.Caches.DB.ThreadMute.Invalidate("ID", id)
	return nil
}
