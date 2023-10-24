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

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type threadDB struct {
	db    *DB
	state *state.State
}

func (t *threadDB) GetThread(ctx context.Context, id string) (*gtsmodel.Thread, error) {
	thread, err := t.state.Caches.GTS.Thread().Load("ID", func() (*gtsmodel.Thread, error) {
		var thread gtsmodel.Thread

		q := t.db.
			NewSelect().
			Model(&thread).
			Where("? = ?", bun.Ident("thread.id"), id)

		if err := q.Scan(ctx); err != nil {
			return nil, err
		}

		return &thread, nil
	}, id)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return thread, nil
	}

	// Use intermediate table to select
	// all status IDs in this thread.
	if err := t.db.
		NewSelect().
		Table("thread_to_statuses").
		Column("thread_to_statuses.status_id").
		Where("? = ?", bun.Ident("thread_to_statuses.thread_id"), id).
		Scan(ctx, &thread.StatusIDs); err != nil {
		return nil, gtserror.Newf("db error populating thread: %w", err)
	}

	return thread, nil
}

func (t *threadDB) PutThread(ctx context.Context, thread *gtsmodel.Thread) error {
	return t.state.Caches.GTS.Thread().Store(thread, func() error {
		_, err := t.db.NewInsert().Model(thread).Exec(ctx)
		return err
	})
}

func (t *threadDB) DeleteThread(ctx context.Context, id string) error {
	if err := t.db.RunInTx(ctx, func(tx Tx) error {
		// Delete the thread itself.
		_, err := tx.
			NewDelete().
			TableExpr("? AS ?", bun.Ident("threads"), bun.Ident("thread")).
			Where("? = ?", bun.Ident("thread.id"), id).
			Exec(ctx)
		if err != nil {
			return err
		}

		// Update statuses that use the given thread id.
		//
		// Note: deleting a thread (which is just a grouping)
		// doesn't delete statuses in the thread, it just
		// de-threads them.
		_, err = tx.
			NewUpdate().
			Table("statuses").
			Set("? = NULL", bun.Ident("thread_id")).
			Where("? = ?", bun.Ident("thread_id"), id).
			Exec(ctx)
		if err != nil {
			return err
		}

		// Delete entries from intermediary table.
		_, err = tx.
			NewDelete().
			TableExpr("? AS ?", bun.Ident("thread_to_statuses"), bun.Ident("thread_to_status")).
			Where("? = ?", bun.Ident("thread_to_status.thread_id"), id).
			Exec(ctx)
		if err != nil {
			return err
		}

		// Delete thread mutes that target the given thread.
		_, err = tx.
			NewDelete().
			TableExpr("? AS ?", bun.Ident("thread_mutes"), bun.Ident("thread_mute")).
			Where("? = ?", bun.Ident("thread_mute.thread_id"), id).
			Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// Operation successful. Invalidate cache
	// entries related to now-deleted thread.
	t.state.Caches.GTS.Thread().Invalidate("ID", id)
	t.state.Caches.GTS.Status().Invalidate("ThreadID", id)
	t.state.Caches.GTS.ThreadMute().Invalidate("ThreadID", id)

	return nil
}

func (t *threadDB) GetThreadMute(ctx context.Context, id string) (*gtsmodel.ThreadMute, error) {
	return t.state.Caches.GTS.ThreadMute().Load("ID", func() (*gtsmodel.ThreadMute, error) {
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
	return t.state.Caches.GTS.ThreadMute().Load("ThreadID.AccountID", func() (*gtsmodel.ThreadMute, error) {
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
	return t.state.Caches.GTS.ThreadMute().Store(threadMute, func() error {
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

	t.state.Caches.GTS.ThreadMute().Invalidate("ID", id)
	return nil
}
