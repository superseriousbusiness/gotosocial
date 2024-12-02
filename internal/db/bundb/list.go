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
	"slices"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

type listDB struct {
	db    *bun.DB
	state *state.State
}

/*
	LIST FUNCTIONS
*/

func (l *listDB) GetListByID(ctx context.Context, id string) (*gtsmodel.List, error) {
	return l.getList(
		ctx,
		"ID",
		func(list *gtsmodel.List) error {
			return l.db.NewSelect().
				Model(list).
				Where("? = ?", bun.Ident("list.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (l *listDB) getList(ctx context.Context, lookup string, dbQuery func(*gtsmodel.List) error, keyParts ...any) (*gtsmodel.List, error) {
	list, err := l.state.Caches.DB.List.LoadOne(lookup, func() (*gtsmodel.List, error) {
		var list gtsmodel.List

		// Not cached! Perform database query.
		if err := dbQuery(&list); err != nil {
			return nil, err
		}

		return &list, nil
	}, keyParts...)
	if err != nil {
		// already processed
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return list, nil
	}

	if err := l.state.DB.PopulateList(ctx, list); err != nil {
		return nil, err
	}

	return list, nil
}

func (l *listDB) GetListsByAccountID(ctx context.Context, accountID string) ([]*gtsmodel.List, error) {
	listIDs, err := l.getListIDsByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return l.GetListsByIDs(ctx, listIDs)
}

func (l *listDB) CountListsByAccountID(ctx context.Context, accountID string) (int, error) {
	listIDs, err := l.getListIDsByAccountID(ctx, accountID)
	return len(listIDs), err
}

func (l *listDB) GetListsContainingFollowID(ctx context.Context, followID string) ([]*gtsmodel.List, error) {
	listIDs, err := l.getListIDsWithFollowID(ctx, followID)
	if err != nil {
		return nil, err
	}
	return l.GetListsByIDs(ctx, listIDs)
}

func (l *listDB) GetFollowsInList(ctx context.Context, listID string, page *paging.Page) ([]*gtsmodel.Follow, error) {
	followIDs, err := l.GetFollowIDsInList(ctx, listID, page)
	if err != nil {
		return nil, err
	}
	return l.state.DB.GetFollowsByIDs(ctx, followIDs)
}

func (l *listDB) GetAccountsInList(ctx context.Context, listID string, page *paging.Page) ([]*gtsmodel.Account, error) {
	accountIDs, err := l.GetAccountIDsInList(ctx, listID, page)
	if err != nil {
		return nil, err
	}
	return l.state.DB.GetAccountsByIDs(ctx, accountIDs)
}

func (l *listDB) IsAccountInList(ctx context.Context, listID string, accountID string) (bool, error) {
	accountIDs, err := l.GetAccountIDsInList(ctx, listID, nil)
	return slices.Contains(accountIDs, accountID), err
}

func (l *listDB) PopulateList(ctx context.Context, list *gtsmodel.List) error {
	var (
		err  error
		errs gtserror.MultiError
	)

	if list.Account == nil {
		// List account is not set, fetch from the database.
		list.Account, err = l.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			list.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating list account: %w", err)
		}
	}

	return errs.Combine()
}

func (l *listDB) PutList(ctx context.Context, list *gtsmodel.List) error {
	// note that inserting list will call OnInvalidateList()
	// which will handle clearing caches other than List cache.
	return l.state.Caches.DB.List.Store(list, func() error {
		_, err := l.db.NewInsert().Model(list).Exec(ctx)
		return err
	})
}

func (l *listDB) UpdateList(ctx context.Context, list *gtsmodel.List, columns ...string) error {
	list.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	// Update list in the database, invalidating main list cache.
	if err := l.state.Caches.DB.List.Store(list, func() error {
		_, err := l.db.NewUpdate().
			Model(list).
			Where("? = ?", bun.Ident("list.id"), list.ID).
			Column(columns...).
			Exec(ctx)
		return err
	}); err != nil {
		return err
	}

	// Invalidate this entire list's timeline.
	if err := l.state.Timelines.List.RemoveTimeline(ctx, list.ID); err != nil {
		log.Errorf(ctx, "error invalidating list timeline: %q", err)
	}

	return nil
}

func (l *listDB) DeleteListByID(ctx context.Context, id string) error {
	// Acquire list owner ID.
	var accountID string

	// Gather follow IDs of all
	// entries contained in list.
	var followIDs []string

	// Delete all list entries associated with list, and list itself in transaction.
	if err := l.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewDelete().
			Table("list_entries").
			Where("? = ?", bun.Ident("list_id"), id).
			Returning("?", bun.Ident("follow_id")).
			Exec(ctx, &followIDs); err != nil {
			return err
		}

		_, err := tx.NewDelete().
			Table("lists").
			Where("? = ?", bun.Ident("id"), id).
			Returning("?", bun.Ident("account_id")).
			Exec(ctx, &accountID)
		return err
	}); err != nil {
		return err
	}

	// Invalidate the main list database cache.
	l.state.Caches.DB.List.Invalidate("ID", id)

	// Invalidate cache of list IDs owned by account.
	l.state.Caches.DB.ListIDs.Invalidate("a" + accountID)

	// Invalidate all related entry caches for this list.
	l.invalidateEntryCaches(ctx, []string{id}, followIDs)

	return nil
}

func (l *listDB) getListIDsByAccountID(ctx context.Context, accountID string) ([]string, error) {
	return l.state.Caches.DB.ListIDs.Load("a"+accountID, func() ([]string, error) {
		var listIDs []string

		// List IDs not in cache.
		// Perform the DB query.
		if _, err := l.db.NewSelect().
			Table("lists").
			Column("id").
			Where("? = ?", bun.Ident("account_id"), accountID).
			OrderExpr("? DESC", bun.Ident("created_at")).
			Exec(ctx, &listIDs); err != nil &&
			!errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return listIDs, nil
	})
}

func (l *listDB) getListIDsWithFollowID(ctx context.Context, followID string) ([]string, error) {
	return l.state.Caches.DB.ListIDs.Load("f"+followID, func() ([]string, error) {
		var listIDs []string

		// List IDs not in cache.
		// Perform the DB query.
		if _, err := l.db.NewSelect().
			Table("list_entries").
			Column("list_id").
			Where("? = ?", bun.Ident("follow_id"), followID).
			OrderExpr("? DESC", bun.Ident("created_at")).
			Exec(ctx, &listIDs); err != nil &&
			!errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return listIDs, nil
	})
}

func (l *listDB) GetFollowIDsInList(ctx context.Context, listID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(&l.state.Caches.DB.ListedIDs, "f"+listID, page, func() ([]string, error) {
		var followIDs []string

		// Follow IDs not in cache.
		// Perform the DB query.
		_, err := l.db.NewSelect().
			Table("list_entries").
			Column("follow_id").
			Where("? = ?", bun.Ident("list_id"), listID).
			OrderExpr("? DESC", bun.Ident("created_at")).
			Exec(ctx, &followIDs)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return followIDs, nil
	})
}

func (l *listDB) GetAccountIDsInList(ctx context.Context, listID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(&l.state.Caches.DB.ListedIDs, "a"+listID, page, func() ([]string, error) {
		var accountIDs []string

		// Account IDs not in cache.
		// Perform the DB query.
		_, err := l.db.NewSelect().
			Table("follows").
			Column("follows.target_account_id").
			Join("INNER JOIN ?", bun.Ident("list_entries")).
			JoinOn("? = ?", bun.Ident("follows.id"), bun.Ident("list_entries.follow_id")).
			Where("? = ?", bun.Ident("list_entries.list_id"), listID).
			OrderExpr("? DESC", bun.Ident("list_entries.id")).
			Exec(ctx, &accountIDs)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return accountIDs, nil
	})
}

func (l *listDB) GetListsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.List, error) {
	// Load all list IDs via cache loader callbacks.
	lists, err := l.state.Caches.DB.List.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.List, error) {
			// Preallocate expected length of uncached lists.
			lists := make([]*gtsmodel.List, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := l.db.NewSelect().
				Model(&lists).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return lists, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the lists by their
	// IDs to ensure in correct order.
	getID := func(l *gtsmodel.List) string { return l.ID }
	xslices.OrderBy(lists, ids, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return lists, nil
	}

	// Populate all loaded lists, removing those we fail to
	// populate (removes needing so many nil checks everywhere).
	lists = slices.DeleteFunc(lists, func(list *gtsmodel.List) bool {
		if err := l.PopulateList(ctx, list); err != nil {
			log.Errorf(ctx, "error populating list %s: %v", list.ID, err)
			return true
		}
		return false
	})

	return lists, nil
}

func (l *listDB) PopulateListEntry(ctx context.Context, listEntry *gtsmodel.ListEntry) error {
	var err error

	if listEntry.Follow == nil {
		// ListEntry follow is not set, fetch from the database.
		listEntry.Follow, err = l.state.DB.GetFollowByID(
			gtscontext.SetBarebones(ctx),
			listEntry.FollowID,
		)
		if err != nil {
			return fmt.Errorf("error populating listEntry follow: %w", err)
		}
	}

	return nil
}

func (l *listDB) PutListEntries(ctx context.Context, entries []*gtsmodel.ListEntry) error {
	// Insert all entries into the database in a single transaction (all or nothing!).
	if err := l.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, entry := range entries {
			if _, err := tx.
				NewInsert().
				Model(entry).
				Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Collect unique list IDs from the provided list entries.
	listIDs := xslices.Collate(entries, func(e *gtsmodel.ListEntry) string {
		return e.ListID
	})

	// Collect unique follow IDs from the provided list entries.
	followIDs := xslices.Collate(entries, func(e *gtsmodel.ListEntry) string {
		return e.FollowID
	})

	// Invalidate all related list entry caches.
	l.invalidateEntryCaches(ctx, listIDs, followIDs)

	return nil
}

func (l *listDB) DeleteListEntry(ctx context.Context, listID string, followID string) error {
	// Delete list entry with given
	// ID, returning its list ID.
	if _, err := l.db.NewDelete().
		Table("list_entries").
		Where("? = ?", bun.Ident("list_id"), listID).
		Where("? = ?", bun.Ident("follow_id"), followID).
		Exec(ctx, &listID); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate all related list entry caches.
	l.invalidateEntryCaches(ctx, []string{listID},
		[]string{followID})

	return nil
}

func (l *listDB) DeleteAllListEntriesByFollows(ctx context.Context, followIDs ...string) error {
	var listIDs []string

	// Check for empty list.
	if len(followIDs) == 0 {
		return nil
	}

	// Delete all entries with follow
	// ID, returning IDs and list IDs.
	if _, err := l.db.NewDelete().
		Table("list_entries").
		Where("? IN (?)", bun.Ident("follow_id"), bun.In(followIDs)).
		Returning("?", bun.Ident("list_id")).
		Exec(ctx, &listIDs); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Deduplicate IDs before invalidate.
	listIDs = xslices.Deduplicate(listIDs)

	// Invalidate all related list entry caches.
	l.invalidateEntryCaches(ctx, listIDs, followIDs)

	return nil
}

// invalidateEntryCaches will invalidate all related ListEntry caches for given list IDs and follow IDs, including timelines.
func (l *listDB) invalidateEntryCaches(ctx context.Context, listIDs, followIDs []string) {
	var keys []string

	// Generate ListedID keys to invalidate.
	keys = slices.Grow(keys[:0], 2*len(listIDs))
	for _, listID := range listIDs {
		keys = append(keys,
			"a"+listID,
			"f"+listID,
		)

		// Invalidate the timeline for the list this entry belongs to.
		if err := l.state.Timelines.List.RemoveTimeline(ctx, listID); err != nil {
			log.Errorf(ctx, "error invalidating list timeline: %q", err)
		}
	}

	// Invalidate ListedID slice cache entries.
	l.state.Caches.DB.ListedIDs.Invalidate(keys...)

	// Generate ListID keys to invalidate.
	keys = slices.Grow(keys[:0], len(followIDs))
	for _, followID := range followIDs {
		keys = append(keys, "f"+followID)
	}

	// Invalidate ListID slice cache entries.
	l.state.Caches.DB.ListIDs.Invalidate(keys...)
}
