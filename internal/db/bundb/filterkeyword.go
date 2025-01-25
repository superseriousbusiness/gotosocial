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

	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

func (f *filterDB) GetFilterKeywordByID(ctx context.Context, id string) (*gtsmodel.FilterKeyword, error) {
	filterKeyword, err := f.state.Caches.DB.FilterKeyword.LoadOne(
		"ID",
		func() (*gtsmodel.FilterKeyword, error) {
			var filterKeyword gtsmodel.FilterKeyword

			// Scan from DB.
			if err := f.db.
				NewSelect().
				Model(&filterKeyword).
				Where("? = ?", bun.Ident("id"), id).
				Scan(ctx); err != nil {
				return nil, err
			}

			// Pre-compile filter keyword regular expression.
			if err := filterKeyword.Compile(); err != nil {
				return nil, gtserror.Newf("error compiling filter keyword regex: %w", err)
			}

			return &filterKeyword, nil
		},
		id,
	)
	if err != nil {
		return nil, err
	}

	if !gtscontext.Barebones(ctx) {
		err = f.populateFilterKeyword(ctx, filterKeyword)
		if err != nil {
			return nil, err
		}
	}

	return filterKeyword, nil
}

func (f *filterDB) populateFilterKeyword(ctx context.Context, filterKeyword *gtsmodel.FilterKeyword) (err error) {
	if filterKeyword.Filter == nil {
		// Filter is not set, fetch from the cache or database.
		filterKeyword.Filter, err = f.state.DB.GetFilterByID(

			// Don't populate the filter with all of its keywords
			// and statuses or we'll just end up back here.
			gtscontext.SetBarebones(ctx),
			filterKeyword.FilterID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *filterDB) GetFilterKeywordsForFilterID(ctx context.Context, filterID string) ([]*gtsmodel.FilterKeyword, error) {
	return f.getFilterKeywords(ctx, "filter_id", filterID)
}

func (f *filterDB) GetFilterKeywordsForAccountID(ctx context.Context, accountID string) ([]*gtsmodel.FilterKeyword, error) {
	return f.getFilterKeywords(ctx, "account_id", accountID)
}

func (f *filterDB) getFilterKeywords(ctx context.Context, idColumn string, id string) ([]*gtsmodel.FilterKeyword, error) {
	var filterKeywordIDs []string

	if err := f.db.
		NewSelect().
		Model((*gtsmodel.FilterKeyword)(nil)).
		Column("id").
		Where("? = ?", bun.Ident(idColumn), id).
		Scan(ctx, &filterKeywordIDs); err != nil {
		return nil, err
	}

	if len(filterKeywordIDs) == 0 {
		return nil, nil
	}

	// Get each filter keyword by ID from the cache or DB.
	filterKeywords, err := f.state.Caches.DB.FilterKeyword.LoadIDs("ID",
		filterKeywordIDs,
		func(uncached []string) ([]*gtsmodel.FilterKeyword, error) {
			filterKeywords := make([]*gtsmodel.FilterKeyword, 0, len(uncached))

			if err := f.db.
				NewSelect().
				Model(&filterKeywords).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			// Compile all the keyword regular expressions.
			filterKeywords = slices.DeleteFunc(filterKeywords, func(filterKeyword *gtsmodel.FilterKeyword) bool {
				if err := filterKeyword.Compile(); err != nil {
					log.Errorf(ctx, "error compiling filter keyword regex: %v", err)
					return true
				}
				return false
			})

			return filterKeywords, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Put the filter keyword structs in the same order as the filter keyword IDs.
	xslices.OrderBy(filterKeywords, filterKeywordIDs, func(filterKeyword *gtsmodel.FilterKeyword) string {
		return filterKeyword.ID
	})

	if gtscontext.Barebones(ctx) {
		return filterKeywords, nil
	}

	// Populate the filter keywords. Remove any that we can't populate from the return slice.
	filterKeywords = slices.DeleteFunc(filterKeywords, func(filterKeyword *gtsmodel.FilterKeyword) bool {
		if err := f.populateFilterKeyword(ctx, filterKeyword); err != nil {
			log.Errorf(ctx, "error populating filter keyword: %v", err)
			return true
		}
		return false
	})

	return filterKeywords, nil
}

func (f *filterDB) PutFilterKeyword(ctx context.Context, filterKeyword *gtsmodel.FilterKeyword) error {
	if filterKeyword.Regexp == nil {
		// Ensure regexp is compiled
		// before attempted caching.
		err := filterKeyword.Compile()
		if err != nil {
			return gtserror.Newf("error compiling filter keyword regex: %w", err)
		}
	}
	return f.state.Caches.DB.FilterKeyword.Store(filterKeyword, func() error {
		_, err := f.db.
			NewInsert().
			Model(filterKeyword).
			Exec(ctx)
		return err
	})
}

func (f *filterDB) UpdateFilterKeyword(ctx context.Context, filterKeyword *gtsmodel.FilterKeyword, columns ...string) error {
	filterKeyword.UpdatedAt = time.Now()
	if len(columns) > 0 {
		columns = append(columns, "updated_at")
	}
	if filterKeyword.Regexp == nil {
		// Ensure regexp is compiled
		// before attempted caching.
		err := filterKeyword.Compile()
		if err != nil {
			return gtserror.Newf("error compiling filter keyword regex: %w", err)
		}
	}
	return f.state.Caches.DB.FilterKeyword.Store(filterKeyword, func() error {
		_, err := f.db.
			NewUpdate().
			Model(filterKeyword).
			Where("? = ?", bun.Ident("id"), filterKeyword.ID).
			Column(columns...).
			Exec(ctx)
		return err
	})
}

func (f *filterDB) DeleteFilterKeywordByID(ctx context.Context, id string) error {
	if _, err := f.db.
		NewDelete().
		Model((*gtsmodel.FilterKeyword)(nil)).
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil {
		return err
	}

	f.state.Caches.DB.FilterKeyword.Invalidate("ID", id)

	return nil
}
