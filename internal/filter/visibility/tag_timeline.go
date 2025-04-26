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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

// StatusHomeTimelineable checks if given status should be included
// on requester's tag timeline, primarily relying on status visibility
// to requester and the AP visibility setting.
func (f *Filter) StatusTagTimelineable(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (bool, error) {
	if status.CreatedAt.After(time.Now().Add(24 * time.Hour)) {
		// Statuses made over 1 day in the future we don't show...
		log.Warnf(ctx, "status >24hrs in the future: %+v", status)
		return false, nil
	}

	// Don't show boosts on tag timeline.
	if status.BoostOfID != "" {
		return false, nil
	}

	// Check whether status is visible to requesting account.
	visible, err := f.StatusVisible(ctx, requester, status)
	if err != nil {
		return false, err
	}

	if !visible {
		log.Trace(ctx, "status not visible to timeline requester")
		return false, nil
	}

	// Looks good!
	return true, nil
}
