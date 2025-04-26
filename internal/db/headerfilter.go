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
	"net/http"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

type HeaderFilter interface {
	// AllowHeaderRegularMatch performs an headerfilter.Filter.RegularMatch() on cached allow header filters.
	// (Note: the actual matching code can be found under ./internal/headerfilter/ ).
	AllowHeaderRegularMatch(ctx context.Context, hdr http.Header) (string, string, error)

	// AllowHeaderInverseMatch performs an headerfilter.Filter.InverseMatch() on cached allow header filters.
	// (Note: the actual matching code can be found under ./internal/headerfilter/ ).
	AllowHeaderInverseMatch(ctx context.Context, hdr http.Header) (string, string, error)

	// BlockHeaderRegularMatch performs an headerfilter.Filter.RegularMatch() on cached block header filters.
	// (Note: the actual matching code can be found under ./internal/headerfilter/ ).
	BlockHeaderRegularMatch(ctx context.Context, hdr http.Header) (string, string, error)

	// BlockHeaderInverseMatch performs an headerfilter.Filter.InverseMatch() on cached block header filters.
	// (Note: the actual matching code can be found under ./internal/headerfilter/ ).
	BlockHeaderInverseMatch(ctx context.Context, hdr http.Header) (string, string, error)

	// GetAllowHeaderFilter fetches the allow header filter with ID from the database.
	GetAllowHeaderFilter(ctx context.Context, id string) (*gtsmodel.HeaderFilter, error)

	// GetBlockHeaderFilter fetches the block header filter with ID from the database.
	GetBlockHeaderFilter(ctx context.Context, id string) (*gtsmodel.HeaderFilter, error)

	// GetAllowHeaderFilters fetches all allow header filters from the database.
	GetAllowHeaderFilters(ctx context.Context) ([]*gtsmodel.HeaderFilter, error)

	// GetBlockHeaderFilters fetches all block header filters from the database.
	GetBlockHeaderFilters(ctx context.Context) ([]*gtsmodel.HeaderFilter, error)

	// PutAllowHeaderFilter inserts the given allow header filter into the database.
	PutAllowHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter) error

	// PutBlockHeaderFilter inserts the given block header filter into the database.
	PutBlockHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter) error

	// UpdateAllowHeaderFilter updates the given allow header filter in the database, only updating given columns if provided.
	UpdateAllowHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter, cols ...string) error

	// UpdateBlockHeaderFilter updates the given block header filter in the database, only updating given columns if provided.
	UpdateBlockHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter, cols ...string) error

	// DeleteAllowHeaderFilter deletes the allow header filter with ID from the database.
	DeleteAllowHeaderFilter(ctx context.Context, id string) error

	// DeleteBlockHeaderFilter deletes the block header filter with ID from the database.
	DeleteBlockHeaderFilter(ctx context.Context, id string) error
}
