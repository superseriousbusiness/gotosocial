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

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type HeaderFilter interface {
	// HeaderAllow performs a positive match of given http headers against stored allow header filters.
	// (Note: the actual matching code can be found under ./internal/headerfilter/ ).
	HeaderAllow(ctx context.Context, hdr http.Header) (bool, error)

	// HeaderBlock performs a negative match of given http headers against stored block header filters.
	// (Note: the actual matching code can be found under ./internal/headerfilter/ ).
	HeaderBlock(ctx context.Context, hdr http.Header) (bool, error)

	// PutAllowHeaderFilter ...
	PutAllowHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilterAllow) error

	// PutBlockHeaderFilter ...
	PutBlockHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilterAllow) error

	// UpdateAllowHeaderFilter ...
	UpdateAllowHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilterAllow, cols ...string) error

	// UpdateBlockHeaderFilter ...
	UpdateBlockHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilterAllow, cols ...string) error

	// DeleteAllowHeaderFilter ...
	DeleteAllowHeaderFilter(ctx context.Context, id string) error

	// DeleteBlockHeaderFilter ...
	DeleteBlockHeaderFilter(ctx context.Context, id string) error
}
