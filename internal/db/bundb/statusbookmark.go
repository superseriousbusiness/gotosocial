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
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type statusBookmarkDB struct {
	conn  *DBConn
	state *state.State
}

func (s *statusBookmarkDB) GetStatusBookmark(ctx context.Context, id string) (*gtsmodel.StatusBookmark, db.Error) {
	bookmark := new(gtsmodel.StatusBookmark)

	err := s.conn.
		NewSelect().
		Model(bookmark).
		Where("? = ?", bun.Ident("status_bookmark.id"), id).
		Scan(ctx)
	if err != nil {
		return nil, s.conn.ProcessError(err)
	}

	bookmark.Account, err = s.state.DB.GetAccountByID(ctx, bookmark.AccountID)
	if err != nil {
		return nil, fmt.Errorf("error getting status bookmark account %q: %w", bookmark.AccountID, err)
	}

	bookmark.TargetAccount, err = s.state.DB.GetAccountByID(ctx, bookmark.TargetAccountID)
	if err != nil {
		return nil, fmt.Errorf("error getting status bookmark target account %q: %w", bookmark.TargetAccountID, err)
	}

	bookmark.Status, err = s.state.DB.GetStatusByID(ctx, bookmark.StatusID)
	if err != nil {
		return nil, fmt.Errorf("error getting status bookmark status %q: %w", bookmark.StatusID, err)
	}

	return bookmark, nil
}

func (s *statusBookmarkDB) GetStatusBookmarkID(ctx context.Context, accountID string, statusID string) (string, db.Error) {
	var id string

	q := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("status_bookmarks"), bun.Ident("status_bookmark")).
		Column("status_bookmark.id").
		Where("? = ?", bun.Ident("status_bookmark.account_id"), accountID).
		Where("? = ?", bun.Ident("status_bookmark.status_id"), statusID).
		Limit(1)

	if err := q.Scan(ctx, &id); err != nil {
		return "", s.conn.ProcessError(err)
	}

	return id, nil
}

func (s *statusBookmarkDB) GetStatusBookmarks(ctx context.Context, accountID string, limit int, maxID string, minID string) ([]*gtsmodel.StatusBookmark, db.Error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Guess size of IDs based on limit.
	ids := make([]string, 0, limit)

	q := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("status_bookmarks"), bun.Ident("status_bookmark")).
		Column("status_bookmark.id").
		Where("? = ?", bun.Ident("status_bookmark.account_id"), accountID).
		Order("status_bookmark.id DESC")

	if accountID == "" {
		return nil, errors.New("must provide an account")
	}

	if maxID != "" {
		q = q.Where("? < ?", bun.Ident("status_bookmark.id"), maxID)
	}

	if minID != "" {
		q = q.Where("? > ?", bun.Ident("status_bookmark.id"), minID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &ids); err != nil {
		return nil, s.conn.ProcessError(err)
	}

	bookmarks := make([]*gtsmodel.StatusBookmark, 0, len(ids))

	for _, id := range ids {
		bookmark, err := s.GetStatusBookmark(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting bookmark %q: %v", id, err)
			continue
		}

		bookmarks = append(bookmarks, bookmark)
	}

	return bookmarks, nil
}

func (s *statusBookmarkDB) PutStatusBookmark(ctx context.Context, statusBookmark *gtsmodel.StatusBookmark) db.Error {
	_, err := s.conn.
		NewInsert().
		Model(statusBookmark).
		Exec(ctx)

	return s.conn.ProcessError(err)
}

func (s *statusBookmarkDB) DeleteStatusBookmark(ctx context.Context, id string) db.Error {
	_, err := s.conn.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("status_bookmarks"), bun.Ident("status_bookmark")).
		Where("? = ?", bun.Ident("status_bookmark.id"), id).
		Exec(ctx)

	return s.conn.ProcessError(err)
}

func (s *statusBookmarkDB) DeleteStatusBookmarks(ctx context.Context, targetAccountID string, originAccountID string) db.Error {
	if targetAccountID == "" && originAccountID == "" {
		return errors.New("DeleteBookmarks: one of targetAccountID or originAccountID must be set")
	}

	// TODO: Capture bookmark IDs in a RETURNING
	// statement (when bookmarks have a cache),
	// + use the IDs to invalidate cache entries.

	q := s.conn.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("status_bookmarks"), bun.Ident("status_bookmark"))

	if targetAccountID != "" {
		q = q.Where("? = ?", bun.Ident("status_bookmark.target_account_id"), targetAccountID)
	}

	if originAccountID != "" {
		q = q.Where("? = ?", bun.Ident("status_bookmark.account_id"), originAccountID)
	}

	if _, err := q.Exec(ctx); err != nil {
		return s.conn.ProcessError(err)
	}

	return nil
}

func (s *statusBookmarkDB) DeleteStatusBookmarksForStatus(ctx context.Context, statusID string) db.Error {
	// TODO: Capture bookmark IDs in a RETURNING
	// statement (when bookmarks have a cache),
	// + use the IDs to invalidate cache entries.

	q := s.conn.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("status_bookmarks"), bun.Ident("status_bookmark")).
		Where("? = ?", bun.Ident("status_bookmark.status_id"), statusID)

	if _, err := q.Exec(ctx); err != nil {
		return s.conn.ProcessError(err)
	}

	return nil
}
