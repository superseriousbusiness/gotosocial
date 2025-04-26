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

type Search interface {
	// SearchForAccounts uses the given query text to search for accounts that accountID follows.
	SearchForAccounts(ctx context.Context, accountID string, query string, maxID string, minID string, limit int, following bool, offset int) ([]*gtsmodel.Account, error)

	// SearchForStatuses uses the given query text to search for statuses created by requestingAccountID, or in reply to requestingAccountID.
	// If fromAccountID is used, the results are restricted to statuses created by fromAccountID.
	SearchForStatuses(ctx context.Context, requestingAccountID string, query string, fromAccountID string, maxID string, minID string, limit int, offset int) ([]*gtsmodel.Status, error)

	// SearchForTags searches for tags that start with the given query text (case insensitive).
	SearchForTags(ctx context.Context, query string, maxID string, minID string, limit int, offset int) ([]*gtsmodel.Tag, error)
}
