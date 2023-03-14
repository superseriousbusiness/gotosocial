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

package visibility

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// StatusBoostable checks if given status is boostable by requester, checking boolean status visibility to requester and ultimately the AP status visibility setting.
func (f *Filter) StatusBoostable(ctx context.Context, requester *gtsmodel.Account, status *gtsmodel.Status) (bool, error) {
	if status.Visibility == gtsmodel.VisibilityDirect {
		log.Trace(ctx, "direct statuses are not boostable")
		return false, nil
	}

	// Check whether status is visible to requesting account.
	visible, err := f.StatusVisible(ctx, requester, status)
	if err != nil {
		return false, err
	}

	if !visible {
		log.Trace(ctx, "status not visible to requesting account")
		return false, nil
	}

	if requester.ID == status.AccountID {
		// Status author can always boost non-directs.
		return true, nil
	}

	if status.Visibility == gtsmodel.VisibilityFollowersOnly ||
		status.Visibility == gtsmodel.VisibilityMutualsOnly {
		log.Trace(ctx, "unauthored %s status not boostable", status.Visibility)
		return false, nil
	}

	if !*status.Boostable {
		log.Trace(ctx, "status marked not boostable")
		return false, nil
	}

	return true, nil
}
