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

// Filter contains methods for creating, reading, updating,
// and deleting filters and their keyword and status entries.
type Filter interface {

	// GetFilterByID gets one filter with the given id.
	GetFilterByID(ctx context.Context, id string) (*gtsmodel.Filter, error)

	// GetFiltersByAccountID gets all filters owned by the given accountID.
	GetFiltersByAccountID(ctx context.Context, accountID string) ([]*gtsmodel.Filter, error)

	// PutFilter puts a new filter in the database, adding any attached keywords or statuses.
	// It uses a transaction to ensure no partial updates.
	PutFilter(ctx context.Context, filter *gtsmodel.Filter) error

	// UpdateFilter ...
	UpdateFilter(ctx context.Context, filter *gtsmodel.Filter, cols ...string) error

	// DeleteFilter deletes the given filter and all associated FilterKeyword{}
	// and FilterStatus{} models from the database in a single transaction.
	DeleteFilter(ctx context.Context, filter *gtsmodel.Filter) error

	// GetFilterKeywordByID gets one filter keyword with the given ID.
	GetFilterKeywordByID(ctx context.Context, id string) (*gtsmodel.FilterKeyword, error)

	// GetFilterKeywordsByIDs ...
	GetFilterKeywordsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.FilterKeyword, error)

	// PutFilterKeyword inserts a single filter keyword into the database.
	PutFilterKeyword(ctx context.Context, filterKeyword *gtsmodel.FilterKeyword) error

	// UpdateFilterKeyword updates the given filter keyword.
	UpdateFilterKeyword(ctx context.Context, filterKeyword *gtsmodel.FilterKeyword, cols ...string) error

	// DeleteFilterKeywordsByIDs deletes filter keywords with the given ids.
	DeleteFilterKeywordsByIDs(ctx context.Context, ids ...string) error

	// GetFilterStatusByID gets one filter status with the given ID.
	GetFilterStatusByID(ctx context.Context, id string) (*gtsmodel.FilterStatus, error)

	// GetFilterStatusesByIDs ...
	GetFilterStatusesByIDs(ctx context.Context, ids []string) ([]*gtsmodel.FilterStatus, error)

	// PutFilterStatus inserts a single filter status into the database.
	PutFilterStatus(ctx context.Context, filterStatus *gtsmodel.FilterStatus) error

	// DeleteFilterStatusesByIDs deletes filter statuses with the given ids.
	DeleteFilterStatusesByIDs(ctx context.Context, ids ...string) error
}
