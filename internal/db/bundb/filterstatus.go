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
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

func (f *filterDB) GetFilterStatusByID(ctx context.Context, id string) (*gtsmodel.FilterStatus, error) {
	filterStatus, err := f.state.Caches.DB.FilterStatus.LoadOne(
		"ID",
		func() (*gtsmodel.FilterStatus, error) {
			var filterStatus gtsmodel.FilterStatus
			err := f.db.
				NewSelect().
				Model(&filterStatus).
				Where("? = ?", bun.Ident("id"), id).
				Scan(ctx)
			return &filterStatus, err
		},
		id,
	)
	if err != nil {
		return nil, err
	}

	if !gtscontext.Barebones(ctx) {
		err = f.populateFilterStatus(ctx, filterStatus)
		if err != nil {
			return nil, err
		}
	}

	return filterStatus, nil
}

func (f *filterDB) populateFilterStatus(ctx context.Context, filterStatus *gtsmodel.FilterStatus) error {
	if filterStatus.Filter == nil {
		// Filter is not set, fetch from the cache or database.
		filter, err := f.state.DB.GetFilterByID(
			// Don't populate the filter with all of its keywords and statuses or we'll just end up back here.
			gtscontext.SetBarebones(ctx),
			filterStatus.FilterID,
		)
		if err != nil {
			return err
		}
		filterStatus.Filter = filter
	}

	return nil
}

func (f *filterDB) GetFilterStatusesForFilterID(ctx context.Context, filterID string) ([]*gtsmodel.FilterStatus, error) {
	return f.getFilterStatuses(ctx, "filter_id", filterID)
}

func (f *filterDB) GetFilterStatusesForAccountID(ctx context.Context, accountID string) ([]*gtsmodel.FilterStatus, error) {
	return f.getFilterStatuses(ctx, "account_id", accountID)
}

func (f *filterDB) getFilterStatuses(ctx context.Context, idColumn string, id string) ([]*gtsmodel.FilterStatus, error) {
	var filterStatusIDs []string
	if err := f.db.
		NewSelect().
		Model((*gtsmodel.FilterStatus)(nil)).
		Column("id").
		Where("? = ?", bun.Ident(idColumn), id).
		Scan(ctx, &filterStatusIDs); err != nil {
		return nil, err
	}
	if len(filterStatusIDs) == 0 {
		return nil, nil
	}

	// Get each filter status by ID from the cache or DB.
	filterStatuses, err := f.state.Caches.DB.FilterStatus.LoadIDs("ID",
		filterStatusIDs,
		func(uncached []string) ([]*gtsmodel.FilterStatus, error) {
			filterStatuses := make([]*gtsmodel.FilterStatus, 0, len(uncached))
			if err := f.db.
				NewSelect().
				Model(&filterStatuses).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}
			return filterStatuses, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Put the filter status structs in the same order as the filter status IDs.
	util.OrderBy(filterStatuses, filterStatusIDs, func(filterStatus *gtsmodel.FilterStatus) string {
		return filterStatus.ID
	})

	if gtscontext.Barebones(ctx) {
		return filterStatuses, nil
	}

	// Populate the filter statuses. Remove any that we can't populate from the return slice.
	errs := gtserror.NewMultiError(len(filterStatuses))
	filterStatuses = slices.DeleteFunc(filterStatuses, func(filterStatus *gtsmodel.FilterStatus) bool {
		if err := f.populateFilterStatus(ctx, filterStatus); err != nil {
			errs.Appendf(
				"error populating filter status %s: %w",
				filterStatus.ID,
				err,
			)
			return true
		}
		return false
	})

	return filterStatuses, errs.Combine()
}

func (f *filterDB) PutFilterStatus(ctx context.Context, filterStatus *gtsmodel.FilterStatus) error {
	return f.state.Caches.DB.FilterStatus.Store(filterStatus, func() error {
		_, err := f.db.
			NewInsert().
			Model(filterStatus).
			Exec(ctx)
		return err
	})
}

func (f *filterDB) UpdateFilterStatus(ctx context.Context, filterStatus *gtsmodel.FilterStatus, columns ...string) error {
	filterStatus.UpdatedAt = time.Now()
	if len(columns) > 0 {
		columns = append(columns, "updated_at")
	}

	return f.state.Caches.DB.FilterStatus.Store(filterStatus, func() error {
		_, err := f.db.
			NewUpdate().
			Model(filterStatus).
			Where("? = ?", bun.Ident("id"), filterStatus.ID).
			Column(columns...).
			Exec(ctx)
		return err
	})
}

func (f *filterDB) DeleteFilterStatusByID(ctx context.Context, id string) error {
	if _, err := f.db.
		NewDelete().
		Model((*gtsmodel.FilterStatus)(nil)).
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil {
		return err
	}

	f.state.Caches.DB.FilterStatus.Invalidate("ID", id)

	return nil
}
