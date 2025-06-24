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

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

func (f *filterDB) GetFilterKeywordByID(ctx context.Context, id string) (*gtsmodel.FilterKeyword, error) {
	return f.state.Caches.DB.FilterKeyword.LoadOne(
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
}

func (f *filterDB) GetFilterKeywordsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.FilterKeyword, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	// Get each filter keyword by ID from the cache or DB.
	filterKeywords, err := f.state.Caches.DB.FilterKeyword.LoadIDs("ID",
		ids,
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
	xslices.OrderBy(filterKeywords, ids, func(filterKeyword *gtsmodel.FilterKeyword) string {
		return filterKeyword.ID
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

func (f *filterDB) UpdateFilterKeyword(ctx context.Context, filterKeyword *gtsmodel.FilterKeyword, cols ...string) error {
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
			Column(cols...).
			Exec(ctx)
		return err
	})
}

func (f *filterDB) DeleteFilterKeywordsByIDs(ctx context.Context, ids ...string) error {
	if _, err := f.db.
		NewDelete().
		Model((*gtsmodel.FilterKeyword)(nil)).
		Where("? IN (?)", bun.Ident("id"), bun.In(ids)).
		Exec(ctx); err != nil {
		return err
	}
	f.state.Caches.DB.FilterKeyword.InvalidateIDs("ID", ids)
	return nil
}
