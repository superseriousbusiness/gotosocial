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
	// HeaderMatchPositive performs a positive match of given http headers against stored positive type header filters.
	// (Note: the actual matching code can be found under ./internal/headerfilter/ ).
	HeaderMatchPositive(ctx context.Context, hdr http.Header) (bool, error)

	// HeaderMatchNegative performs a negative match of given http headers against stored negative type header filters.
	// (Note: the actual matching code can be found under ./internal/headerfilter/ ).
	HeaderMatchNegative(ctx context.Context, hdr http.Header) (bool, error)

	// PutHeaderFilter inserts the given header filter into the database.
	PutHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter) error
}
