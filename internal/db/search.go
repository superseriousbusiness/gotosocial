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

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type Search interface {
	// SearchForAccounts uses the given query text to search for accounts that accountID follows.
	SearchForAccounts(ctx context.Context, accountID string, query string, maxID string, minID string, limit int, following bool, offset int) ([]*gtsmodel.Account, error)

	// SearchForStatuses uses the given query text to search for statuses created by accountID.
	SearchForStatuses(ctx context.Context, accountID string, query string, maxID string, minID string, limit int, offset int) ([]*gtsmodel.Status, error)
}
