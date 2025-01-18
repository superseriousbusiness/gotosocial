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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

type filterDB struct {
	db    *bun.DB
	state *state.State
}

func (f *filterDB) GetFilterByID(ctx context.Context, id string) (*gtsmodel.Filter, error) {
	filter, err := f.state.Caches.DB.Filter.LoadOne(
		"ID",
		func() (*gtsmodel.Filter, error) {
			var filter gtsmodel.Filter
			err := f.db.
				NewSelect().
				Model(&filter).
				Where("? = ?", bun.Ident("id"), id).
				Scan(ctx)
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
	filters, err := f.state.Caches.DB.Filter.LoadIDs("ID",
		filterIDs,
		func(uncached []string) ([]*gtsmodel.Filter, error) {
			filters := make([]*gtsmodel.Filter, 0, len(uncached))
			if err := f.db.
				NewSelect().
				Model(&filters).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}
			return filters, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Put the filter structs in the same order as the filter IDs.
	xslices.OrderBy(filters, filterIDs, func(filter *gtsmodel.Filter) string { return filter.ID })

	if gtscontext.Barebones(ctx) {
		return filters, nil
	}

	// Populate the filters. Remove any that we can't populate from the return slice.
	errs := gtserror.NewMultiError(len(filters))
	filters = slices.DeleteFunc(filters, func(filter *gtsmodel.Filter) bool {
		if err := f.populateFilter(ctx, filter); err != nil {
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
			errs.Appendf("error populating filter statuses: %w", err)
		}
		for i := range filter.Statuses {
			filter.Statuses[i].Filter = filter
		}
	}

	return errs.Combine()
}

func (f *filterDB) PutFilter(ctx context.Context, filter *gtsmodel.Filter) error {
	// Pre-compile filter keyword regular expressions.
	for _, filterKeyword := range filter.Keywords {
		if err := filterKeyword.Compile(); err != nil {
			return gtserror.Newf("error compiling filter keyword regex: %w", err)
		}
	}

	// Update database.
	if err := f.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.
			NewInsert().
			Model(filter).
			Exec(ctx); err != nil {
			return err
		}

		if len(filter.Keywords) > 0 {
			if _, err := tx.
				NewInsert().
				Model(&filter.Keywords).
				Exec(ctx); err != nil {
				return err
			}
		}

		if len(filter.Statuses) > 0 {
			if _, err := tx.
				NewInsert().
				Model(&filter.Statuses).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	// Update cache.
	f.state.Caches.DB.Filter.Put(filter)
	f.state.Caches.DB.FilterKeyword.Put(filter.Keywords...)
	f.state.Caches.DB.FilterStatus.Put(filter.Statuses...)

	return nil
}

func (f *filterDB) UpdateFilter(
	ctx context.Context,
	filter *gtsmodel.Filter,
	filterColumns []string,
	filterKeywordColumns [][]string,
	deleteFilterKeywordIDs []string,
	deleteFilterStatusIDs []string,
) error {
	if len(filter.Keywords) != len(filterKeywordColumns) {
		return errors.New("number of filter keywords must match number of lists of filter keyword columns")
	}

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
	for i := range filterKeywordColumns {
		if len(filterKeywordColumns[i]) > 0 {
			filterKeywordColumns[i] = append(filterKeywordColumns[i], "updated_at")
		}
	}

	// Pre-compile filter keyword regular expressions.
	for _, filterKeyword := range filter.Keywords {
		if err := filterKeyword.Compile(); err != nil {
			return gtserror.Newf("error compiling filter keyword regex: %w", err)
		}
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

		for i, filterKeyword := range filter.Keywords {
			if _, err := NewUpsert(tx).
				Model(filterKeyword).
				Constraint("id").
				Column(filterKeywordColumns[i]...).
				Exec(ctx); err != nil {
				return err
			}
		}

		if len(filter.Statuses) > 0 {
			if _, err := tx.
				NewInsert().
				Ignore().
				Model(&filter.Statuses).
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
	f.state.Caches.DB.Filter.Put(filter)
	f.state.Caches.DB.FilterKeyword.Put(filter.Keywords...)
	f.state.Caches.DB.FilterStatus.Put(filter.Statuses...)
	// TODO: (Vyr) replace with cache multi-invalidate call
	for _, id := range deleteFilterKeywordIDs {
		f.state.Caches.DB.FilterKeyword.Invalidate("ID", id)
	}
	for _, id := range deleteFilterStatusIDs {
		f.state.Caches.DB.FilterStatus.Invalidate("ID", id)
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
	f.state.Caches.DB.Filter.Invalidate("ID", id)

	// Invalidate all keywords and statuses for this filter.
	f.state.Caches.DB.FilterKeyword.Invalidate("FilterID", id)
	f.state.Caches.DB.FilterStatus.Invalidate("FilterID", id)

	return nil
}
