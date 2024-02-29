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
	"slices"
	"time"

	"codeberg.org/gruf/go-structr"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

type filterDB struct {
	db    *bun.DB
	state *state.State
}

//region Filter methods

func (f *filterDB) GetFilterByID(ctx context.Context, id string) (*gtsmodel.Filter, error) {
	filter, err := f.state.Caches.GTS.Filter.LoadOne(
		"ID",
		func() (*gtsmodel.Filter, error) {
			var filter gtsmodel.Filter
			err := f.db.NewSelect().Model(&filter).Where("? = ?", bun.Ident("id"), id).Scan(ctx)
			return &filter, err
		},
		id,
	)
	if err != nil {
		// already processed
		return nil, err
	}

	if !gtscontext.Barebones(ctx) {
		if err := f.populateFilter(ctx, filter); err != nil {
			return nil, err
		}
	}

	return filter, nil
}

func (f *filterDB) GetFiltersForAccountID(ctx context.Context, accountID string) ([]*gtsmodel.Filter, error) {
	// Fetch IDs of all filters owned by this account.
	var filterIDs []string
	if err := f.db.
		NewSelect().
		Model((*gtsmodel.Filter)(nil)).
		Column("id").
		Where("? = ?", bun.Ident("account_id"), accountID).
		Scan(ctx, &filterIDs); err != nil {
		return nil, err
	}
	if len(filterIDs) == 0 {
		return nil, nil
	}

	// Get each filter by ID from the cache or DB.
	uncachedFilterIDs := make([]string, 0, len(filterIDs))
	filters, err := f.state.Caches.GTS.Filter.Load(
		"ID",
		func(load func(keyParts ...any) bool) {
			for _, id := range filterIDs {
				if !load(id) {
					uncachedFilterIDs = append(uncachedFilterIDs, id)
				}
			}
		},
		func() ([]*gtsmodel.Filter, error) {
			uncachedFilters := make([]*gtsmodel.Filter, 0, len(uncachedFilterIDs))
			if err := f.db.
				NewSelect().
				Model(&uncachedFilters).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncachedFilterIDs)).
				Scan(ctx); err != nil {
				return nil, err
			}
			return uncachedFilters, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Put the filter structs in the same order as the filter IDs.
	util.OrderBy(filters, filterIDs, func(filter *gtsmodel.Filter) string { return filter.ID })

	if gtscontext.Barebones(ctx) {
		return filters, nil
	}

	// Populate the filters. Remove any that we can't populate from the return slice.
	errs := gtserror.NewMultiError(len(filters))
	filters = slices.DeleteFunc(filters, func(filter *gtsmodel.Filter) bool {
		if err := f.populateFilter(ctx, filter); err != nil {
			// %w is allowed here.
			//goland:noinspection GoPrintFunctions
			errs.Appendf("error populating filter %s: %w", filter.ID, err)
			return true
		}
		return false
	})

	return filters, errs.Combine()
}

func (f *filterDB) populateFilter(ctx context.Context, filter *gtsmodel.Filter) error {
	var err error
	errs := gtserror.NewMultiError(2)

	if filter.Keywords == nil {
		// Filter keywords are not set, fetch from the database.
		filter.Keywords, err = f.state.DB.GetFilterKeywordsForFilterID(
			gtscontext.SetBarebones(ctx),
			filter.ID,
		)
		if err != nil {
			// %w is allowed here.
			//goland:noinspection GoPrintFunctions
			errs.Appendf("error populating filter keywords: %w", err)
		}
		for i := range filter.Keywords {
			filter.Keywords[i].Filter = filter
		}
	}

	if filter.Statuses == nil {
		// Filter statuses are not set, fetch from the database.
		filter.Statuses, err = f.state.DB.GetFilterStatusesForFilterID(
			gtscontext.SetBarebones(ctx),
			filter.ID,
		)
		if err != nil {
			// %w is allowed here.
			//goland:noinspection GoPrintFunctions
			errs.Appendf("error populating filter statuses: %w", err)
		}
		for i := range filter.Statuses {
			filter.Statuses[i].Filter = filter
		}
	}

	return errs.Combine()
}

func (f *filterDB) PutFilter(ctx context.Context, filter *gtsmodel.Filter) error {
	// Update database.
	if err := f.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(filter).Exec(ctx); err != nil {
			return err
		}

		if len(filter.Keywords) > 0 {
			if _, err := tx.NewInsert().Model(&filter.Keywords).Exec(ctx); err != nil {
				return err
			}
		}

		if len(filter.Statuses) > 0 {
			if _, err := tx.NewInsert().Model(&filter.Statuses).Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	// Update cache.
	f.state.Caches.GTS.Filter.Put(filter)
	f.state.Caches.GTS.FilterKeyword.Put(filter.Keywords...)
	f.state.Caches.GTS.FilterStatus.Put(filter.Statuses...)

	return nil
}

func (f *filterDB) UpdateFilter(
	ctx context.Context,
	filter *gtsmodel.Filter,
	filterColumns []string,
	filterKeywordColumns []string,
	filterStatusColumns []string,
	deleteFilterKeywordIDs []string,
	deleteFilterStatusIDs []string,
) error {
	updatedAt := time.Now()
	filter.UpdatedAt = updatedAt
	for _, filterKeyword := range filter.Keywords {
		filterKeyword.UpdatedAt = updatedAt
	}
	for _, filterStatus := range filter.Statuses {
		filterStatus.UpdatedAt = updatedAt
	}

	// If we're updating by column, ensure "updated_at" is included.
	if len(filterColumns) > 0 {
		filterColumns = append(filterColumns, "updated_at")
	}
	if len(filterKeywordColumns) > 0 {
		filterKeywordColumns = append(filterKeywordColumns, "updated_at")
	}
	if len(filterStatusColumns) > 0 {
		filterStatusColumns = append(filterStatusColumns, "updated_at")
	}

	// Update database.
	if err := f.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.
			NewUpdate().
			Model(filter).
			Column(filterColumns...).
			Where("? = ?", bun.Ident("id"), filter.ID).
			Exec(ctx); err != nil {
			return err
		}

		if len(filter.Keywords) > 0 {
			if _, err := NewUpsert(tx).
				Model(&filter.Keywords).
				Constraint("id").
				Column(filterKeywordColumns...).
				Exec(ctx); err != nil {
				return err
			}
		}

		if len(filter.Statuses) > 0 {
			if _, err := NewUpsert(tx).
				Model(&filter.Statuses).
				Constraint("id").
				Column(filterStatusColumns...).
				Exec(ctx); err != nil {
				return err
			}
		}

		if len(deleteFilterKeywordIDs) > 0 {
			if _, err := tx.
				NewDelete().
				Model((*gtsmodel.FilterKeyword)(nil)).
				Where("? = (?)", bun.Ident("id"), bun.In(deleteFilterKeywordIDs)).
				Exec(ctx); err != nil {
				return err
			}
		}

		if len(deleteFilterStatusIDs) > 0 {
			if _, err := tx.
				NewDelete().
				Model((*gtsmodel.FilterStatus)(nil)).
				Where("? = (?)", bun.Ident("id"), bun.In(deleteFilterStatusIDs)).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	// Update cache.
	f.state.Caches.GTS.Filter.Put(filter)
	f.state.Caches.GTS.FilterKeyword.Put(filter.Keywords...)
	f.state.Caches.GTS.FilterStatus.Put(filter.Statuses...)
	// TODO: (Vyr) replace with cache multi-invalidate call
	for _, id := range deleteFilterKeywordIDs {
		f.state.Caches.GTS.FilterKeyword.Invalidate("ID", id)
	}
	for _, id := range deleteFilterStatusIDs {
		f.state.Caches.GTS.FilterStatus.Invalidate("ID", id)
	}

	return nil
}

func (f *filterDB) DeleteFilterByID(ctx context.Context, id string) error {
	if err := f.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Delete all keywords attached to filter.
		if _, err := tx.
			NewDelete().
			Model((*gtsmodel.FilterKeyword)(nil)).
			Where("? = ?", bun.Ident("filter_id"), id).
			Exec(ctx); err != nil {
			return err
		}

		// Delete all statuses attached to filter.
		if _, err := tx.
			NewDelete().
			Model((*gtsmodel.FilterStatus)(nil)).
			Where("? = ?", bun.Ident("filter_id"), id).
			Exec(ctx); err != nil {
			return err
		}

		// Delete the filter itself.
		_, err := tx.
			NewDelete().
			Model((*gtsmodel.Filter)(nil)).
			Where("? = ?", bun.Ident("id"), id).
			Exec(ctx)
		return err
	}); err != nil {
		return err
	}

	// Invalidate this filter.
	f.state.Caches.GTS.Filter.Invalidate("ID", id)

	// Invalidate all keywords and statuses for this filter.
	f.state.Caches.GTS.FilterKeyword.Invalidate("FilterID", id)
	f.state.Caches.GTS.FilterStatus.Invalidate("FilterID", id)

	return nil
}

//endregion

//region Filter entries

// entry wraps a pointer to a model extending [gtsmodel.FilterEntry]
// and provides operations common across those models.
type entry[Model any, Self any] interface {
	name() string
	cache(f *filterDB) *structr.Cache[*Model]
	new() Self
	get() *Model
	set(model *Model)
	common() *gtsmodel.FilterEntry
}

// newEntry creates a new entry.
// This is a workaround for Go's inability to define a self-referential interface type.
// See https://appliedgo.com/blog/generic-interface-functions
func newEntry[Entry entry[Model, Entry], Model any]() Entry {
	var entry Entry
	return entry.new()
}

//region keywordEntry

// keywordEntry implements entry for a filter keyword.
type keywordEntry struct {
	*gtsmodel.FilterKeyword
}

func (e *keywordEntry) new() *keywordEntry {
	return &keywordEntry{}
}

func (e *keywordEntry) name() string {
	return "keyword"
}

func (e *keywordEntry) cache(f *filterDB) *structr.Cache[*gtsmodel.FilterKeyword] {
	return &f.state.Caches.GTS.FilterKeyword
}

func (e *keywordEntry) get() *gtsmodel.FilterKeyword {
	return e.FilterKeyword
}

func (e *keywordEntry) set(model *gtsmodel.FilterKeyword) {
	e.FilterKeyword = model
}

func (e *keywordEntry) common() *gtsmodel.FilterEntry {
	return &e.FilterEntry
}

//endregion

//region statusEntry

// statusEntry implements entry for a filter status.
type statusEntry struct {
	*gtsmodel.FilterStatus
}

func (e *statusEntry) name() string {
	return "status"
}

func (e *statusEntry) cache(f *filterDB) *structr.Cache[*gtsmodel.FilterStatus] {
	return &f.state.Caches.GTS.FilterStatus
}

func (e *statusEntry) new() *statusEntry {
	return &statusEntry{}
}

func (e *statusEntry) get() *gtsmodel.FilterStatus {
	return e.FilterStatus
}

func (e *statusEntry) set(model *gtsmodel.FilterStatus) {
	e.FilterStatus = model
}

func (e *statusEntry) common() *gtsmodel.FilterEntry {
	return &e.FilterEntry
}

//endregion

// Note that methods can't be generic, so these have to be free functions.

func getFilterEntryByID[Entry entry[Model, Entry], Model any](ctx context.Context, f *filterDB, id string) (*Model, error) {
	entry := newEntry[Entry, Model]()
	model, err := entry.cache(f).LoadOne(
		"ID",
		func() (*Model, error) {
			var model Model
			err := f.db.NewSelect().Model(&model).Where("? = ?", bun.Ident("id"), id).Scan(ctx)
			return &model, err
		},
		id,
	)
	if err != nil {
		return nil, err
	}
	entry.set(model)

	if !gtscontext.Barebones(ctx) {
		if err := populateFilterEntry[Entry, Model](ctx, f, entry); err != nil {
			return nil, err
		}
	}

	return entry.get(), nil
}

func populateFilterEntry[Entry entry[Model, Entry], Model any](ctx context.Context, f *filterDB, entry Entry) error {
	common := entry.common()
	if common.Filter == nil {
		// Filter is not set, fetch from the cache or database.
		filter, err := f.state.DB.GetFilterByID(
			// Don't populate the filter with all of its entries or we'll just end up back here.
			gtscontext.SetBarebones(ctx),
			common.FilterID,
		)
		if err != nil {
			return err
		}
		common.Filter = filter
	}

	return nil
}

func getFilterEntries[Entry entry[Model, Entry], Model any](ctx context.Context, f *filterDB, idColumn string, id string) ([]*Model, error) {
	// Intentionally doesn't contain a model; we only need it for its type.
	var entry Entry

	var filterEntryIDs []string
	if err := f.db.
		NewSelect().
		Model((*Model)(nil)).
		Column("id").
		Where("? = ?", bun.Ident(idColumn), id).
		Scan(ctx, &filterEntryIDs); err != nil {
		return nil, err
	}
	if len(filterEntryIDs) == 0 {
		return nil, nil
	}

	// Get each filter entry by ID from the cache or DB.
	uncachedFilterEntryIDs := make([]string, 0, len(filterEntryIDs))
	filterEntries, err := entry.cache(f).Load(
		"ID",
		func(load func(keyParts ...any) bool) {
			for _, id := range filterEntryIDs {
				if !load(id) {
					uncachedFilterEntryIDs = append(uncachedFilterEntryIDs, id)
				}
			}
		},
		func() ([]*Model, error) {
			uncachedFilterEntries := make([]*Model, 0, len(uncachedFilterEntryIDs))
			if err := f.db.
				NewSelect().
				Model(&uncachedFilterEntries).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncachedFilterEntryIDs)).
				Scan(ctx); err != nil {
				return nil, err
			}
			return uncachedFilterEntries, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Put the entry structs in the same order as the entry IDs.
	util.OrderBy(filterEntries, filterEntryIDs, func(model *Model) string {
		entry := newEntry[Entry, Model]()
		entry.set(model)
		return entry.common().ID
	})

	if gtscontext.Barebones(ctx) {
		return filterEntries, nil
	}

	// Populate the entries. Remove any that we can't populate from the return slice.
	errs := gtserror.NewMultiError(len(filterEntries))
	filterEntries = slices.DeleteFunc(filterEntries, func(model *Model) bool {
		entry := newEntry[Entry, Model]()
		entry.set(model)
		if err := populateFilterEntry[Entry, Model](ctx, f, entry); err != nil {
			// %w is allowed here.
			//goland:noinspection GoPrintFunctions
			errs.Appendf(
				"error populating filter %s %s: %w",
				entry.name(),
				entry.common().ID,
				err,
			)
			return true
		}
		return false
	})

	return filterEntries, errs.Combine()
}

func putFilterEntry[Entry entry[Model, Entry], Model any](ctx context.Context, f *filterDB, model *Model) error {
	// Intentionally doesn't contain a model; we only need it for its type.
	var entry Entry

	return entry.cache(f).Store(model, func() error {
		_, err := f.db.NewInsert().Model(model).Exec(ctx)
		return err
	})
}

func updateFilterEntry[Entry entry[Model, Entry], Model any](ctx context.Context, f *filterDB, model *Model, columns []string) error {
	entry := newEntry[Entry, Model]()
	entry.set(model)

	entry.common().UpdatedAt = time.Now()
	if len(columns) > 0 {
		columns = append(columns, "updated_at")
	}

	return entry.cache(f).Store(model, func() error {
		_, err := f.db.
			NewUpdate().
			Model(model).
			Where("? = ?", bun.Ident("id"), entry.common().ID).
			Column(columns...).
			Exec(ctx)
		return err
	})
}

func deleteFilterEntryByID[Entry entry[Model, Entry], Model any](ctx context.Context, f *filterDB, id string) error {
	// Intentionally doesn't contain a model; we only need it for its type.
	var entry Entry

	if _, err := f.db.
		NewDelete().
		Model((*Model)(nil)).
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil {
		return err
	}

	entry.cache(f).Invalidate("ID", id)

	return nil
}

//endregion

//region Filter keyword methods

func (f *filterDB) GetFilterKeywordByID(ctx context.Context, id string) (*gtsmodel.FilterKeyword, error) {
	return getFilterEntryByID[*keywordEntry, gtsmodel.FilterKeyword](ctx, f, id)
}

func (f *filterDB) GetFilterKeywordsForFilterID(ctx context.Context, filterID string) ([]*gtsmodel.FilterKeyword, error) {
	return getFilterEntries[*keywordEntry, gtsmodel.FilterKeyword](ctx, f, "filter_id", filterID)
}

func (f *filterDB) GetFilterKeywordsForAccountID(ctx context.Context, accountID string) ([]*gtsmodel.FilterKeyword, error) {
	return getFilterEntries[*keywordEntry, gtsmodel.FilterKeyword](ctx, f, "account_id", accountID)
}

func (f *filterDB) PutFilterKeyword(ctx context.Context, filterKeyword *gtsmodel.FilterKeyword) error {
	return putFilterEntry[*keywordEntry](ctx, f, filterKeyword)
}

func (f *filterDB) UpdateFilterKeyword(ctx context.Context, filterKeyword *gtsmodel.FilterKeyword, columns ...string) error {
	return updateFilterEntry[*keywordEntry](ctx, f, filterKeyword, columns)
}

func (f *filterDB) DeleteFilterKeywordByID(ctx context.Context, id string) error {
	return deleteFilterEntryByID[*keywordEntry, gtsmodel.FilterKeyword](ctx, f, id)
}

//endregion

//region Filter status methods

func (f *filterDB) GetFilterStatusByID(ctx context.Context, id string) (*gtsmodel.FilterStatus, error) {
	return getFilterEntryByID[*statusEntry, gtsmodel.FilterStatus](ctx, f, id)
}

func (f *filterDB) GetFilterStatusesForFilterID(ctx context.Context, filterID string) ([]*gtsmodel.FilterStatus, error) {
	return getFilterEntries[*statusEntry, gtsmodel.FilterStatus](ctx, f, "filter_id", filterID)
}

func (f *filterDB) GetFilterStatusesForAccountID(ctx context.Context, accountID string) ([]*gtsmodel.FilterStatus, error) {
	return getFilterEntries[*statusEntry, gtsmodel.FilterStatus](ctx, f, "account_id", accountID)
}

func (f *filterDB) PutFilterStatus(ctx context.Context, filterStatus *gtsmodel.FilterStatus) error {
	return putFilterEntry[*statusEntry](ctx, f, filterStatus)
}

func (f *filterDB) UpdateFilterStatus(ctx context.Context, filterStatus *gtsmodel.FilterStatus, columns ...string) error {
	return updateFilterEntry[*statusEntry](ctx, f, filterStatus, columns)
}

func (f *filterDB) DeleteFilterStatusByID(ctx context.Context, id string) error {
	return deleteFilterEntryByID[*statusEntry, gtsmodel.FilterStatus](ctx, f, id)
}

//endregion
