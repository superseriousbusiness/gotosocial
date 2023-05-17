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
	conn  *DBConn
	state *state.State
}

/*
	LIST FUNCTIONS
*/

func (l *listDB) getList(ctx context.Context, lookup string, dbQuery func(*gtsmodel.List) error, keyParts ...any) (*gtsmodel.List, error) {
	list, err := l.state.Caches.GTS.List().Load(lookup, func() (*gtsmodel.List, error) {
		var list gtsmodel.List

		// Not cached! Perform database query.
		if err := dbQuery(&list); err != nil {
			return nil, l.conn.ProcessError(err)
		}

		return &list, nil
	}, keyParts...)
	if err != nil {
		return nil, err // already processed
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

func (l *listDB) GetListByID(ctx context.Context, id string) (*gtsmodel.List, error) {
	return l.getList(
		ctx,
		"ID",
		func(list *gtsmodel.List) error {
			return l.conn.NewSelect().
				Model(list).
				Where("? = ?", bun.Ident("list.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (l *listDB) GetListsForAccountID(ctx context.Context, accountID string) ([]*gtsmodel.List, error) {
	// Fetch IDs of all lists owned by this account.
	var listIDs []string
	if err := l.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("lists"), bun.Ident("list")).
		Column("list.id").
		Where("? = ?", bun.Ident("list.account_id"), accountID).
		Order("list.id DESC").
		Scan(ctx, &listIDs); err != nil {
		return nil, l.conn.ProcessError(err)
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

		// Append list.
		lists = append(lists, list)
	}

	return lists, nil
}

func (l *listDB) PopulateList(ctx context.Context, list *gtsmodel.List) error {
	var (
		err  error
		errs = make(gtserror.MultiError, 0, 2)
	)

	if list.Account == nil {
		// List account is not set, fetch from the database.
		list.Account, err = l.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			list.AccountID,
		)
		if err != nil {
			errs.Append(fmt.Errorf("error populating list account: %w", err))
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
			errs.Append(fmt.Errorf("error populating list entries: %w", err))
		}
	}

	return errs.Combine()
}

func (l *listDB) PutList(ctx context.Context, list *gtsmodel.List) error {
	return l.state.Caches.GTS.List().Store(list, func() error {
		_, err := l.conn.NewInsert().Model(list).Exec(ctx)
		return l.conn.ProcessError(err)
	})
}

func (l *listDB) UpdateList(ctx context.Context, list *gtsmodel.List, columns ...string) error {
	list.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	return l.state.Caches.GTS.List().Store(list, func() error {
		if _, err := l.conn.NewUpdate().
			Model(list).
			Where("? = ?", bun.Ident("list.id"), list.ID).
			Column(columns...).
			Exec(ctx); err != nil {
			return l.conn.ProcessError(err)
		}

		return nil
	})
}

func (l *listDB) DeleteListByID(ctx context.Context, id string) error {
	defer l.state.Caches.GTS.List().Invalidate("ID", id)

	// Select all entries that belong to this list.
	listEntries, err := l.state.DB.GetListEntries(ctx, id, "", "", "", 0)
	if err != nil {
		return fmt.Errorf("error selecting entries from list %q: %w", id, err)
	}

	// Delete each entry.
	for _, listEntry := range listEntries {
		err := l.state.DB.DeleteListEntry(ctx, listEntry.ID)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return err
		}
	}

	// Finally delete list itself from DB.
	_, err = l.conn.NewDelete().
		Table("lists").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx)
	return l.conn.ProcessError(err)
}

/*
	LIST ENTRY functions
*/

func (l *listDB) getListEntry(ctx context.Context, lookup string, dbQuery func(*gtsmodel.ListEntry) error, keyParts ...any) (*gtsmodel.ListEntry, error) {
	listEntry, err := l.state.Caches.GTS.ListEntry().Load(lookup, func() (*gtsmodel.ListEntry, error) {
		var listEntry gtsmodel.ListEntry

		// Not cached! Perform database query.
		if err := dbQuery(&listEntry); err != nil {
			return nil, l.conn.ProcessError(err)
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

func (l *listDB) GetListEntryByID(ctx context.Context, id string) (*gtsmodel.ListEntry, error) {
	return l.getListEntry(
		ctx,
		"ID",
		func(listEntry *gtsmodel.ListEntry) error {
			return l.conn.NewSelect().
				Model(listEntry).
				Where("? = ?", bun.Ident("list_entry.id"), id).
				Scan(ctx)
		},
		id,
	)
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

	q := l.conn.
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
		return nil, l.conn.ProcessError(err)
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

		// Append list entries.
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

func (l *listDB) PutListEntries(ctx context.Context, listEntries []*gtsmodel.ListEntry) error {
	return l.conn.RunInTx(ctx, func(tx bun.Tx) error {
		for _, listEntry := range listEntries {
			if _, err := tx.
				NewInsert().
				Model(listEntry).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	})
}

func (l *listDB) DeleteListEntry(ctx context.Context, id string) error {
	defer l.state.Caches.GTS.ListEntry().Invalidate("ID", id)
	_, err := l.conn.NewDelete().
		Table("list_entries").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx)
	return l.conn.ProcessError(err)
}

func (l *listDB) DeleteListEntriesForFollowID(ctx context.Context, followID string) error {
	// Fetch IDs of all entries that pertain to this follow.
	var listEntryIDs []string
	if err := l.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("list_entries"), bun.Ident("list_entry")).
		Column("list_entry.id").
		Where("? = ?", bun.Ident("list_entry.follow_id"), followID).
		Order("list_entry.id DESC").
		Scan(ctx, &listEntryIDs); err != nil {
		return l.conn.ProcessError(err)
	}

	for _, id := range listEntryIDs {
		if err := l.DeleteListEntry(ctx, id); err != nil {
			return err
		}
	}

	return nil
}
