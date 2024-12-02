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
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

type statusBookmarkDB struct {
	db    *bun.DB
	state *state.State
}

func (s *statusBookmarkDB) GetStatusBookmarkByID(ctx context.Context, id string) (*gtsmodel.StatusBookmark, error) {
	return s.getStatusBookmark(
		ctx,
		"ID",
		func(bookmark *gtsmodel.StatusBookmark) error {
			return s.db.
				NewSelect().
				Model(bookmark).
				Where("? = ?", bun.Ident("id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (s *statusBookmarkDB) GetStatusBookmark(ctx context.Context, accountID string, statusID string) (*gtsmodel.StatusBookmark, error) {
	return s.getStatusBookmark(
		ctx,
		"AccountID,StatusID",
		func(bookmark *gtsmodel.StatusBookmark) error {
			return s.db.
				NewSelect().
				Model(bookmark).
				Where("? = ?", bun.Ident("account_id"), accountID).
				Where("? = ?", bun.Ident("status_id"), statusID).
				Scan(ctx)
		},
		accountID, statusID,
	)
}

func (s *statusBookmarkDB) GetStatusBookmarksByIDs(ctx context.Context, ids []string) ([]*gtsmodel.StatusBookmark, error) {
	// Load all input bookmark IDs via cache loader callback.
	bookmarks, err := s.state.Caches.DB.StatusBookmark.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.StatusBookmark, error) {
			// Preallocate expected length of uncached bookmarks.
			bookmarks := make([]*gtsmodel.StatusBookmark, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) bookmarks.
			if err := s.db.NewSelect().
				Model(&bookmarks).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return bookmarks, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the bookmarks by their
	// IDs to ensure in correct order.
	getID := func(b *gtsmodel.StatusBookmark) string { return b.ID }
	xslices.OrderBy(bookmarks, ids, getID)

	// Populate all loaded bookmarks, removing those we fail
	// to populate (removes needing so many later nil checks).
	bookmarks = slices.DeleteFunc(bookmarks, func(bookmark *gtsmodel.StatusBookmark) bool {
		if err := s.PopulateStatusBookmark(ctx, bookmark); err != nil {
			log.Errorf(ctx, "error populating bookmark %s: %v", bookmark.ID, err)
			return true
		}
		return false
	})

	return bookmarks, nil
}

func (s *statusBookmarkDB) IsStatusBookmarked(ctx context.Context, statusID string) (bool, error) {
	bookmarkIDs, err := s.getStatusBookmarkIDs(ctx, statusID)
	return (len(bookmarkIDs) > 0), err
}

func (s *statusBookmarkDB) IsStatusBookmarkedBy(ctx context.Context, accountID string, statusID string) (bool, error) {
	bookmark, err := s.GetStatusBookmark(ctx, accountID, statusID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return false, err
	}
	return (bookmark != nil), nil
}

func (s *statusBookmarkDB) getStatusBookmark(ctx context.Context, lookup string, dbQuery func(*gtsmodel.StatusBookmark) error, keyParts ...any) (*gtsmodel.StatusBookmark, error) {
	// Fetch bookmark from database cache with loader callback.
	bookmark, err := s.state.Caches.DB.StatusBookmark.LoadOne(lookup, func() (*gtsmodel.StatusBookmark, error) {
		var bookmark gtsmodel.StatusBookmark

		// Not cached! Perform database query.
		if err := dbQuery(&bookmark); err != nil {
			return nil, err
		}

		return &bookmark, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return bookmark, nil
	}

	// Further populate the bookmark fields where applicable.
	if err := s.PopulateStatusBookmark(ctx, bookmark); err != nil {
		return nil, err
	}

	return bookmark, nil
}

func (s *statusBookmarkDB) PopulateStatusBookmark(ctx context.Context, bookmark *gtsmodel.StatusBookmark) (err error) {
	var errs gtserror.MultiError

	if bookmark.Account == nil {
		// Bookmark author is not set, fetch from database.
		bookmark.Account, err = s.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			bookmark.AccountID,
		)
		if err != nil {
			errs.Appendf("error getting bookmark account %s: %w", bookmark.AccountID, err)
		}
	}

	if bookmark.TargetAccount == nil {
		// Bookmark target account is not set, fetch from database.
		bookmark.TargetAccount, err = s.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			bookmark.TargetAccountID,
		)
		if err != nil {
			errs.Appendf("error getting bookmark target account %s: %w", bookmark.TargetAccountID, err)
		}
	}

	if bookmark.Status == nil {
		// Bookmarked status not set, fetch from database.
		bookmark.Status, err = s.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			bookmark.StatusID,
		)
		if err != nil {
			errs.Appendf("error getting bookmark status %s: %w", bookmark.StatusID, err)
		}
	}

	return errs.Combine()
}

func (s *statusBookmarkDB) GetStatusBookmarks(ctx context.Context, accountID string, limit int, maxID string, minID string) ([]*gtsmodel.StatusBookmark, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Guess size of IDs based on limit.
	ids := make([]string, 0, limit)

	q := s.db.
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
		return nil, err
	}

	return s.GetStatusBookmarksByIDs(ctx, ids)
}

func (s *statusBookmarkDB) getStatusBookmarkIDs(ctx context.Context, statusID string) ([]string, error) {
	return s.state.Caches.DB.StatusBookmarkIDs.Load(statusID, func() ([]string, error) {
		var bookmarkIDs []string

		// Bookmark IDs not cached,
		// perform database query.
		if err := s.db.
			NewSelect().
			Table("status_bookmarks").
			Column("id").Where("? = ?", bun.Ident("status_id"), statusID).
			Order("id DESC").
			Scan(ctx, &bookmarkIDs); err != nil {
			return nil, err
		}

		return bookmarkIDs, nil
	})
}

func (s *statusBookmarkDB) PutStatusBookmark(ctx context.Context, bookmark *gtsmodel.StatusBookmark) error {
	return s.state.Caches.DB.StatusBookmark.Store(bookmark, func() error {
		_, err := s.db.NewInsert().Model(bookmark).Exec(ctx)
		return err
	})
}

func (s *statusBookmarkDB) DeleteStatusBookmarkByID(ctx context.Context, id string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted gtsmodel.StatusBookmark
	deleted.ID = id

	// Delete block with given URI,
	// returning the deleted models.
	if _, err := s.db.NewDelete().
		Model(&deleted).
		Where("? = ?", bun.Ident("id"), id).
		Returning("?", bun.Ident("status_id")).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate cached status bookmark by its ID,
	// manually call invalidate hook in case not cached.
	s.state.Caches.DB.StatusBookmark.Invalidate("ID", id)
	s.state.Caches.OnInvalidateStatusBookmark(&deleted)

	return nil
}

func (s *statusBookmarkDB) DeleteStatusBookmarks(ctx context.Context, targetAccountID string, originAccountID string) error {
	if targetAccountID == "" && originAccountID == "" {
		return gtserror.New("one of targetAccountID or originAccountID must be set")
	}

	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted []*gtsmodel.StatusBookmark

	q := s.db.
		NewDelete().
		Model(&deleted).
		Returning("?", bun.Ident("status_id"))

	if targetAccountID != "" {
		q = q.Where("? = ?", bun.Ident("target_account_id"), targetAccountID)
	}

	if originAccountID != "" {
		q = q.Where("? = ?", bun.Ident("account_id"), originAccountID)
	}

	if _, err := q.Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	for _, deleted := range deleted {
		// Invalidate cached status bookmark by status ID,
		// manually call invalidate hook in case not cached.
		s.state.Caches.DB.StatusBookmark.Invalidate("StatusID", deleted.StatusID)
		s.state.Caches.OnInvalidateStatusBookmark(deleted)
	}

	return nil
}

func (s *statusBookmarkDB) DeleteStatusBookmarksForStatus(ctx context.Context, statusID string) error {
	// Delete status bookmarks
	// from database by status ID.
	q := s.db.NewDelete().
		TableExpr("? AS ?", bun.Ident("status_bookmarks"), bun.Ident("status_bookmark")).
		Where("? = ?", bun.Ident("status_bookmark.status_id"), statusID)
	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	// Wrap provided ID in a bookmark
	// model for calling cache hook.
	var deleted gtsmodel.StatusBookmark
	deleted.StatusID = statusID

	// Invalidate cached status bookmark by status ID,
	// manually call invalidate hook in case not cached.
	s.state.Caches.DB.StatusBookmark.Invalidate("StatusID", statusID)
	s.state.Caches.OnInvalidateStatusBookmark(&deleted)

	return nil
}
