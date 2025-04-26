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

package db

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Filter contains methods for creating, reading, updating, and deleting filters and their keyword and status entries.
type Filter interface {
	//<editor-fold desc="Filter methods">

	// GetFilterByID gets one filter with the given id.
	GetFilterByID(ctx context.Context, id string) (*gtsmodel.Filter, error)

	// GetFiltersForAccountID gets all filters owned by the given accountID.
	GetFiltersForAccountID(ctx context.Context, accountID string) ([]*gtsmodel.Filter, error)

	// PutFilter puts a new filter in the database, adding any attached keywords or statuses.
	// It uses a transaction to ensure no partial updates.
	PutFilter(ctx context.Context, filter *gtsmodel.Filter) error

	// UpdateFilter updates the given filter,
	// upserts any attached keywords and inserts any new statuses (existing statuses cannot be updated),
	// and deletes indicated filter keywords and statuses by ID.
	// It uses a transaction to ensure no partial updates.
	// The column lists are optional; if not specified, all columns will be updated.
	// The filter keyword columns list is *per keyword*.
	// To update all keyword columns, provide a list where every element is an empty list.
	UpdateFilter(
		ctx context.Context,
		filter *gtsmodel.Filter,
		filterColumns []string,
		filterKeywordColumns [][]string,
		deleteFilterKeywordIDs []string,
		deleteFilterStatusIDs []string,
	) error

	// DeleteFilterByID deletes one filter with the given ID.
	// It uses a transaction to ensure no partial updates.
	DeleteFilterByID(ctx context.Context, id string) error

	//</editor-fold>

	//<editor-fold desc="Filter keyword methods">

	// GetFilterKeywordByID gets one filter keyword with the given ID.
	GetFilterKeywordByID(ctx context.Context, id string) (*gtsmodel.FilterKeyword, error)

	// GetFilterKeywordsForFilterID gets filter keywords from the given filterID.
	GetFilterKeywordsForFilterID(ctx context.Context, filterID string) ([]*gtsmodel.FilterKeyword, error)

	// GetFilterKeywordsForAccountID gets filter keywords from the given accountID.
	GetFilterKeywordsForAccountID(ctx context.Context, accountID string) ([]*gtsmodel.FilterKeyword, error)

	// PutFilterKeyword inserts a single filter keyword into the database.
	PutFilterKeyword(ctx context.Context, filterKeyword *gtsmodel.FilterKeyword) error

	// UpdateFilterKeyword updates the given filter keyword.
	// Columns is optional, if not specified all will be updated.
	UpdateFilterKeyword(ctx context.Context, filterKeyword *gtsmodel.FilterKeyword, columns ...string) error

	// DeleteFilterKeywordByID deletes one filter keyword with the given id.
	DeleteFilterKeywordByID(ctx context.Context, id string) error

	//</editor-fold>

	//<editor-fold desc="Filter status methods">

	// GetFilterStatusByID gets one filter status with the given ID.
	GetFilterStatusByID(ctx context.Context, id string) (*gtsmodel.FilterStatus, error)

	// GetFilterStatusesForFilterID gets filter statuses from the given filterID.
	GetFilterStatusesForFilterID(ctx context.Context, filterID string) ([]*gtsmodel.FilterStatus, error)

	// GetFilterStatusesForAccountID gets filter keywords from the given accountID.
	GetFilterStatusesForAccountID(ctx context.Context, accountID string) ([]*gtsmodel.FilterStatus, error)

	// PutFilterStatus inserts a single filter status into the database.
	PutFilterStatus(ctx context.Context, filterStatus *gtsmodel.FilterStatus) error

	// DeleteFilterStatusByID deletes one filter status with the given id.
	DeleteFilterStatusByID(ctx context.Context, id string) error

	//</editor-fold>
}
