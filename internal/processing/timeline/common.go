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

package timeline

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/timeline"
)

// SkipInsert returns a function that satisifes SkipInsertFunction.
func SkipInsert() timeline.SkipInsertFunction {
	// Gap to allow between a status or boost of status,
	// and reinsertion of a new boost of that status.
	// This is useful to avoid a heavily boosted status
	// showing up way too often in a user's timeline.
	const boostReinsertionDepth = 50

	return func(
		ctx context.Context,
		newItemID string,
		newItemAccountID string,
		newItemBoostOfID string,
		newItemBoostOfAccountID string,
		nextItemID string,
		nextItemAccountID string,
		nextItemBoostOfID string,
		nextItemBoostOfAccountID string,
		depth int,
	) (bool, error) {
		if newItemID == nextItemID {
			// Don't insert duplicates.
			return true, nil
		}

		if newItemBoostOfID != "" {
			if newItemBoostOfID == nextItemBoostOfID &&
				depth < boostReinsertionDepth {
				// Don't insert boosts of items
				// we've seen boosted recently.
				return true, nil
			}

			if newItemBoostOfID == nextItemID &&
				depth < boostReinsertionDepth {
				// Don't insert boosts of items when
				// we've seen the original recently.
				return true, nil
			}
		}

		// Proceed with insertion
		// (that's what she said!).
		return false, nil
	}
}
