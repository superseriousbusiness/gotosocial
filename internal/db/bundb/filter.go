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

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
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

func (f *filterDB) GetFiltersByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Filter, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	// Get each filter by ID from the cache or DB.
	filters, err := f.state.Caches.DB.Filter.LoadIDs("ID",
		ids,
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
	xslices.OrderBy(filters, ids, func(filter *gtsmodel.Filter) string { return filter.ID })

	if gtscontext.Barebones(ctx) {
		return filters, nil
	}

	var errs gtserror.MultiError

	// Populate the filters. Remove any that we can't populate from the return slice.
	filters = slices.DeleteFunc(filters, func(filter *gtsmodel.Filter) bool {
		if err := f.populateFilter(ctx, filter); err != nil {
			errs.Appendf("error populating filter %s: %w", filter.ID, err)
			return true
		}
		return false
	})

	return filters, errs.Combine()
}

func (f *filterDB) GetFilterIDsByAccountID(ctx context.Context, accountID string) ([]string, error) {
	return f.state.Caches.DB.FilterIDs.Load(accountID, func() ([]string, error) {
		var filterIDs []string

		if err := f.db.
			NewSelect().
			Model((*gtsmodel.Filter)(nil)).
			Column("id").
			Where("? = ?", bun.Ident("account_id"), accountID).
			Scan(ctx, &filterIDs); err != nil {
			return nil, err
		}

		return filterIDs, nil
	})
}

func (f *filterDB) GetFiltersByAccountID(ctx context.Context, accountID string) ([]*gtsmodel.Filter, error) {
	filterIDs, err := f.GetFilterIDsByAccountID(ctx, accountID)
	if err != nil {
		return nil, gtserror.Newf("error getting filter ids: %w", err)
	}
	return f.GetFiltersByIDs(ctx, filterIDs)
}

func (f *filterDB) populateFilter(ctx context.Context, filter *gtsmodel.Filter) error {
	var err error
	var errs gtserror.MultiError

	if !filter.KeywordsPopulated() {
		// Filter keywords are not set, fetch from the database.
		filter.Keywords, err = f.GetFilterKeywordsByIDs(ctx, filter.KeywordIDs)
		if err != nil {
			errs.Appendf("error populating filter keywords: %w", err)
		}
	}

	if !filter.StatusesPopulated() {
		// Filter statuses are not set, fetch from the database.
		filter.Statuses, err = f.GetFilterStatusesByIDs(ctx, filter.StatusIDs)
		if err != nil {
			errs.Appendf("error populating filter statuses: %w", err)
		}
	}

	return errs.Combine()
}

func (f *filterDB) PutFilter(ctx context.Context, filter *gtsmodel.Filter) error {
	return f.state.Caches.DB.Filter.Store(filter, func() error {
		_, err := f.db.NewInsert().Model(filter).Exec(ctx)
		return err
	})
}

func (f *filterDB) UpdateFilter(ctx context.Context, filter *gtsmodel.Filter, cols ...string) error {
	return f.state.Caches.DB.Filter.Store(filter, func() error {
		_, err := f.db.NewUpdate().
			Model(filter).
			Where("? = ?", bun.Ident("id"), filter.ID).
			Column(cols...).
			Exec(ctx)
		return err
	})
}

func (f *filterDB) DeleteFilter(ctx context.Context, filter *gtsmodel.Filter) error {
	if err := f.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Delete all keywords both known
		// by filter, and possible stragglers,
		// storing IDs in filter.KeywordIDs.
		if _, err := tx.
			NewDelete().
			Model((*gtsmodel.FilterKeyword)(nil)).
			Where("? = ?", bun.Ident("filter_id"), filter.ID).
			Returning("?", bun.Ident("id")).
			Exec(ctx, &filter.KeywordIDs); err != nil &&
			!errors.Is(err, db.ErrNoEntries) {
			return err
		}

		// Delete all statuses both known
		// by filter, and possible stragglers.
		// storing IDs in filter.StatusIDs.
		if _, err := tx.
			NewDelete().
			Model((*gtsmodel.FilterStatus)(nil)).
			Where("? = ?", bun.Ident("filter_id"), filter.ID).
			Returning("?", bun.Ident("id")).
			Exec(ctx, &filter.StatusIDs); err != nil &&
			!errors.Is(err, db.ErrNoEntries) {
			return err
		}

		// Delete filter itself.
		_, err := tx.
			NewDelete().
			Model((*gtsmodel.Filter)(nil)).
			Where("? = ?", bun.Ident("id"), filter.ID).
			Exec(ctx)
		return err
	}); err != nil {
		return err
	}

	// Invalidate the filter itself, and
	// call invalidate hook in-case not cached.
	f.state.Caches.DB.Filter.Invalidate("ID", filter.ID)
	f.state.Caches.OnInvalidateFilter(filter)

	return nil
}
