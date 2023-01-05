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
	"fmt"
	"time"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (f *filter) StatusPublictimelineable(ctx context.Context, targetStatus *gtsmodel.Status, timelineOwnerAccount *gtsmodel.Account) (bool, error) {
	l := log.WithFields(kv.Fields{{"statusID", targetStatus.ID}}...)

	// don't timeline statuses more than 5 min in the future
	maxID, err := id.NewULIDFromTime(time.Now().Add(5 * time.Minute))
	if err != nil {
		return false, err
	}

	if targetStatus.ID > maxID {
		l.Debug("status not hometimelineable because it's from more than 5 minutes in the future")
		return false, nil
	}

	// Don't timeline boosted statuses
	if targetStatus.BoostOfID != "" {
		return false, nil
	}

	// Don't timeline a reply
	if targetStatus.InReplyToURI != "" || targetStatus.InReplyToID != "" || targetStatus.InReplyToAccountID != "" {
		return false, nil
	}

	// status owner should always be able to see their own status in their timeline so we can return early if this is the case
	if timelineOwnerAccount != nil && targetStatus.AccountID == timelineOwnerAccount.ID {
		return true, nil
	}

	v, err := f.StatusVisible(ctx, targetStatus, timelineOwnerAccount)
	if err != nil {
		return false, fmt.Errorf("StatusPublictimelineable: error checking visibility of status with id %s: %s", targetStatus.ID, err)
	}

	if !v {
		l.Debug("status is not publicTimelineable because it's not visible to the requester")
		return false, nil
	}

	return true, nil
}
