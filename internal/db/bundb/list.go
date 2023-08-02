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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type listDB struct {
	db    *WrappedDB
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
	list, err := l.state.Caches.GTS.List().Load(lookup, func() (*gtsmodel.List, error) {
		var list gtsmodel.List

		// Not cached! Perform database query.
		if err := dbQuery(&list); err != nil {
			return nil, l.db.ProcessError(err)
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

func (l *listDB) GetListsForAccountID(ctx context.Context, accountID string) ([]*gtsmodel.List, error) {
	// Fetch IDs of all lists owned by this account.
	var listIDs []string
	if err := l.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("lists"), bun.Ident("list")).
		Column("list.id").
		Where("? = ?", bun.Ident("list.account_id"), accountID).
		Order("list.id DESC").
		Scan(ctx, &listIDs); err != nil {
		return nil, l.db.ProcessError(err)
	}

	if len(listIDs) == 0 {
		return nil, nil
	}

	// Select each list using its ID to ensure cache used.
	lists := make([]*gtsmodel.List, 0, len(listIDs))
	for _, id := range listIDs {
		list, err := l.state.DB.GetListByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error fetching list %q: %v", id, err)
			continue
		}
		lists = append(lists, list)
	}

	return lists, nil
}

func (l *listDB) PopulateList(ctx context.Context, list *gtsmodel.List) error {
	var (
		err  error
		errs = gtserror.NewMultiError(2)
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

	if list.ListEntries == nil {
		// List entries are not set, fetch from the database.
		list.ListEntries, err = l.state.DB.GetListEntries(
			gtscontext.SetBarebones(ctx),
			list.ID,
			"", "", "", 0,
		)
		if err != nil {
			errs.Appendf("error populating list entries: %w", err)
		}
	}

	if err := errs.Combine(); err != nil {
		return gtserror.Newf("%w", err)
	}

	return nil
}

func (l *listDB) PutList(ctx context.Context, list *gtsmodel.List) error {
	return l.state.Caches.GTS.List().Store(list, func() error {
		_, err := l.db.NewInsert().Model(list).Exec(ctx)
		return l.db.ProcessError(err)
	})
}

func (l *listDB) UpdateList(ctx context.Context, list *gtsmodel.List, columns ...string) error {
	list.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	defer func() {
		// Invalidate all entries for this list ID.
		l.state.Caches.GTS.ListEntry().Invalidate("ListID", list.ID)

		// Invalidate this entire list's timeline.
		if err := l.state.Timelines.List.RemoveTimeline(ctx, list.ID); err != nil {
			log.Errorf(ctx, "error invalidating list timeline: %q", err)
		}
	}()

	return l.state.Caches.GTS.List().Store(list, func() error {
		_, err := l.db.NewUpdate().
			Model(list).
			Where("? = ?", bun.Ident("list.id"), list.ID).
			Column(columns...).
			Exec(ctx)
		return l.db.ProcessError(err)
	})
}

func (l *listDB) DeleteListByID(ctx context.Context, id string) error {
	// Load list by ID into cache to ensure we can perform
	// all necessary cache invalidation hooks on removal.
	_, err := l.GetListByID(
		// Don't populate the entry;
		// we only want the list ID.
		gtscontext.SetBarebones(ctx),
		id,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// NOTE: even if db.ErrNoEntries is returned, we
		// still run the below transaction to ensure related
		// objects are appropriately deleted.
		return err
	}

	defer func() {
		// Invalidate this list from cache.
		l.state.Caches.GTS.List().Invalidate("ID", id)

		// Invalidate this entire list's timeline.
		if err := l.state.Timelines.List.RemoveTimeline(ctx, id); err != nil {
			log.Errorf(ctx, "error invalidating list timeline: %q", err)
		}
	}()

	return l.db.RunInTx(ctx, func(tx bun.Tx) error {
		// Delete all entries attached to list.
		if _, err := tx.NewDelete().
			Table("list_entries").
			Where("? = ?", bun.Ident("list_id"), id).
			Exec(ctx); err != nil {
			return err
		}

		// Delete the list itself.
		_, err := tx.NewDelete().
			Table("lists").
			Where("? = ?", bun.Ident("id"), id).
			Exec(ctx)
		return err
	})
}

/*
	LIST ENTRY functions
*/

func (l *listDB) GetListEntryByID(ctx context.Context, id string) (*gtsmodel.ListEntry, error) {
	return l.getListEntry(
		ctx,
		"ID",
		func(listEntry *gtsmodel.ListEntry) error {
			return l.db.NewSelect().
				Model(listEntry).
				Where("? = ?", bun.Ident("list_entry.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (l *listDB) getListEntry(ctx context.Context, lookup string, dbQuery func(*gtsmodel.ListEntry) error, keyParts ...any) (*gtsmodel.ListEntry, error) {
	listEntry, err := l.state.Caches.GTS.ListEntry().Load(lookup, func() (*gtsmodel.ListEntry, error) {
		var listEntry gtsmodel.ListEntry

		// Not cached! Perform database query.
		if err := dbQuery(&listEntry); err != nil {
			return nil, l.db.ProcessError(err)
		}

		return &listEntry, nil
	}, keyParts...)
	if err != nil {
		return nil, err // already processed
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return listEntry, nil
	}

	// Further populate the list entry fields where applicable.
	if err := l.state.DB.PopulateListEntry(ctx, listEntry); err != nil {
		return nil, err
	}

	return listEntry, nil
}

func (l *listDB) GetListEntries(ctx context.Context,
	listID string,
	maxID string,
	sinceID string,
	minID string,
	limit int,
) ([]*gtsmodel.ListEntry, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	var (
		entryIDs    = make([]string, 0, limit)
		frontToBack = true
	)

	q := l.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("list_entries"), bun.Ident("entry")).
		// Select only IDs from table
		Column("entry.id").
		// Select only entries belonging to listID.
		Where("? = ?", bun.Ident("entry.list_id"), listID)

	if maxID != "" {
		// return only entries LOWER (ie., older) than maxID
		q = q.Where("? < ?", bun.Ident("entry.id"), maxID)
	}

	if sinceID != "" {
		// return only entries HIGHER (ie., newer) than sinceID
		q = q.Where("? > ?", bun.Ident("entry.id"), sinceID)
	}

	if minID != "" {
		// return only entries HIGHER (ie., newer) than minID
		q = q.Where("? > ?", bun.Ident("entry.id"), minID)

		// page up
		frontToBack = false
	}

	if limit > 0 {
		// limit amount of entries returned
		q = q.Limit(limit)
	}

	if frontToBack {
		// Page down.
		q = q.Order("entry.id DESC")
	} else {
		// Page up.
		q = q.Order("entry.id ASC")
	}

	if err := q.Scan(ctx, &entryIDs); err != nil {
		return nil, l.db.ProcessError(err)
	}

	if len(entryIDs) == 0 {
		return nil, nil
	}

	// If we're paging up, we still want entries
	// to be sorted by ID desc, so reverse ids slice.
	// https://zchee.github.io/golang-wiki/SliceTricks/#reversing
	if !frontToBack {
		for l, r := 0, len(entryIDs)-1; l < r; l, r = l+1, r-1 {
			entryIDs[l], entryIDs[r] = entryIDs[r], entryIDs[l]
		}
	}

	// Select each list entry using its ID to ensure cache used.
	listEntries := make([]*gtsmodel.ListEntry, 0, len(entryIDs))
	for _, id := range entryIDs {
		listEntry, err := l.state.DB.GetListEntryByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error fetching list entry %q: %v", id, err)
			continue
		}
		listEntries = append(listEntries, listEntry)
	}

	return listEntries, nil
}

func (l *listDB) GetListEntriesForFollowID(ctx context.Context, followID string) ([]*gtsmodel.ListEntry, error) {
	var entryIDs []string

	if err := l.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("list_entries"), bun.Ident("entry")).
		// Select only IDs from table
		Column("entry.id").
		// Select only entries belonging with given followID.
		Where("? = ?", bun.Ident("entry.follow_id"), followID).
		Scan(ctx, &entryIDs); err != nil {
		return nil, l.db.ProcessError(err)
	}

	if len(entryIDs) == 0 {
		return nil, nil
	}

	// Select each list entry using its ID to ensure cache used.
	listEntries := make([]*gtsmodel.ListEntry, 0, len(entryIDs))
	for _, id := range entryIDs {
		listEntry, err := l.state.DB.GetListEntryByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error fetching list entry %q: %v", id, err)
			continue
		}
		listEntries = append(listEntries, listEntry)
	}

	return listEntries, nil
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
	defer func() {
		// Collect unique list IDs from the entries.
		listIDs := collate(func(i int) string {
			return entries[i].ListID
		}, len(entries))

		for _, id := range listIDs {
			// Invalidate the timeline for the list this entry belongs to.
			if err := l.state.Timelines.List.RemoveTimeline(ctx, id); err != nil {
				log.Errorf(ctx, "error invalidating list timeline: %q", err)
			}
		}
	}()

	// Finally, insert each list entry into the database.
	return l.db.RunInTx(ctx, func(tx bun.Tx) error {
		for _, entry := range entries {
			if err := l.state.Caches.GTS.ListEntry().Store(entry, func() error {
				_, err := tx.
					NewInsert().
					Model(entry).
					Exec(ctx)
				return err
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (l *listDB) DeleteListEntry(ctx context.Context, id string) error {
	// Load list entry into cache to ensure we can perform
	// all necessary cache invalidation hooks on removal.
	entry, err := l.GetListEntryByID(
		// Don't populate the entry;
		// we only want the list ID.
		gtscontext.SetBarebones(ctx),
		id,
	)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// Already gone.
			return nil
		}
		return err
	}

	defer func() {
		// Invalidate this list entry upon delete.
		l.state.Caches.GTS.ListEntry().Invalidate("ID", id)

		// Invalidate the timeline for the list this entry belongs to.
		if err := l.state.Timelines.List.RemoveTimeline(ctx, entry.ListID); err != nil {
			log.Errorf(ctx, "error invalidating list timeline: %q", err)
		}
	}()

	// Finally delete the list entry.
	_, err = l.db.NewDelete().
		Table("list_entries").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx)
	return err
}

func (l *listDB) DeleteListEntriesForFollowID(ctx context.Context, followID string) error {
	var entryIDs []string

	// Fetch entry IDs for follow ID.
	if err := l.db.
		NewSelect().
		Table("list_entries").
		Column("id").
		Where("? = ?", bun.Ident("follow_id"), followID).
		Order("id DESC").
		Scan(ctx, &entryIDs); err != nil {
		return l.db.ProcessError(err)
	}

	for _, id := range entryIDs {
		// Delete each separately to trigger cache invalidations.
		if err := l.DeleteListEntry(ctx, id); err != nil {
			return err
		}
	}

	return nil
}

// collate will collect the values of type T from an expected slice of length 'len',
// passing the expected index to each call of 'get' and deduplicating the end result.
func collate[T comparable](get func(int) T, len int) []T {
	ts := make([]T, 0, len)
	tm := make(map[T]struct{}, len)

	for i := 0; i < len; i++ {
		// Get next.
		t := get(i)

		if _, ok := tm[t]; !ok {
			// New value, add
			// to map + slice.
			ts = append(ts, t)
			tm[t] = struct{}{}
		}
	}

	return ts
}
