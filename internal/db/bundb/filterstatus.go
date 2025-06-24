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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

func (f *filterDB) GetFilterStatusByID(ctx context.Context, id string) (*gtsmodel.FilterStatus, error) {
	return f.state.Caches.DB.FilterStatus.LoadOne(
		"ID",
		func() (*gtsmodel.FilterStatus, error) {
			var filterStatus gtsmodel.FilterStatus
			if err := f.db.
				NewSelect().
				Model(&filterStatus).
				Where("? = ?", bun.Ident("id"), id).
				Scan(ctx); err != nil {
				return nil, err
			}
			return &filterStatus, nil
		},
		id,
	)
}

func (f *filterDB) GetFilterStatusesByIDs(ctx context.Context, ids []string) ([]*gtsmodel.FilterStatus, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	// Get each filter status by ID from the cache or DB.
	filterStatuses, err := f.state.Caches.DB.FilterStatus.LoadIDs("ID",
		ids,
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
	xslices.OrderBy(filterStatuses, ids, func(filterStatus *gtsmodel.FilterStatus) string {
		return filterStatus.ID
	})

	return filterStatuses, nil
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

func (f *filterDB) DeleteFilterStatusesByIDs(ctx context.Context, ids ...string) error {
	if _, err := f.db.
		NewDelete().
		Model((*gtsmodel.FilterStatus)(nil)).
		Where("? IN (?)", bun.Ident("id"), bun.In(ids)).
		Exec(ctx); err != nil {
		return err
	}
	f.state.Caches.DB.FilterStatus.InvalidateIDs("ID", ids)
	return nil
}
