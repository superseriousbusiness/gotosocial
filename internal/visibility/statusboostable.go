/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package visibility

import (
	"context"
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (f *filter) StatusBoostable(ctx context.Context, targetStatus *gtsmodel.Status, requestingAccount *gtsmodel.Account) (bool, error) {
	// if the status isn't visible, it certainly isn't boostable
	visible, err := f.StatusVisible(ctx, targetStatus, requestingAccount)
	if err != nil {
		return false, fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err)
	}
	if !visible {
		return false, errors.New("status is not visible")
	}

	// direct messages are never boostable, even if they're visible
	if targetStatus.Visibility == gtsmodel.VisibilityDirect {
		log.Trace("status is not boostable because it is a DM")
		return false, nil
	}

	// the original account should always be able to boost its own non-DM statuses
	if requestingAccount.ID == targetStatus.Account.ID {
		log.Trace("status is boostable because author is booster")
		return true, nil
	}

	// if status is followers-only and not the author's, it is not boostable
	if targetStatus.Visibility == gtsmodel.VisibilityFollowersOnly {
		log.Trace("status not boostable because it is followers-only")
		return false, nil
	}

	// otherwise, status is as boostable as it says it is
	log.Trace("defaulting to status.boostable value")
	return *targetStatus.Boostable, nil
}
