/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *filter) StatusHometimelineable(ctx context.Context, targetStatus *gtsmodel.Status, timelineOwnerAccount *gtsmodel.Account) (bool, error) {
	l := logrus.WithFields(logrus.Fields{
		"func":     "StatusHometimelineable",
		"statusID": targetStatus.ID,
	})

	// status owner should always be able to see their own status in their timeline so we can return early if this is the case
	if timelineOwnerAccount != nil && targetStatus.AccountID == timelineOwnerAccount.ID {
		return true, nil
	}

	v, err := f.StatusVisible(ctx, targetStatus, timelineOwnerAccount)
	if err != nil {
		return false, fmt.Errorf("StatusHometimelineable: error checking visibility of status with id %s: %s", targetStatus.ID, err)
	}

	if !v {
		l.Debug("status is not hometimelineable because it's not visible to the requester")
		return false, nil
	}

	for _, m := range targetStatus.Mentions {
		if m.TargetAccountID == timelineOwnerAccount.ID {
			// if we're mentioned we should be able to see the post
			return true, nil
		}
	}

	// Don't timeline a status whose parent hasn't been dereferenced yet or can't be dereferenced.
	// If we have the reply to URI but don't have an ID for the replied-to account or the replied-to status in our database, we haven't dereferenced it yet.
	if targetStatus.InReplyToURI != "" && (targetStatus.InReplyToID == "" || targetStatus.InReplyToAccountID == "") {
		return false, nil
	}

	// if a status replies to an ID we know in the database, we need to make sure we also follow the replied-to status owner account
	if targetStatus.InReplyToID != "" {
		// pin the reply to status on to this status if it hasn't been done already
		if targetStatus.InReplyTo == nil {
			rs, err := f.db.GetStatusByID(ctx, targetStatus.InReplyToID)
			if err != nil {
				return false, fmt.Errorf("StatusHometimelineable: error getting replied to status with id %s: %s", targetStatus.InReplyToID, err)
			}
			targetStatus.InReplyTo = rs
		}

		// pin the reply to account on to this status if it hasn't been done already
		if targetStatus.InReplyToAccount == nil {
			ra, err := f.db.GetAccountByID(ctx, targetStatus.InReplyToAccountID)
			if err != nil {
				return false, fmt.Errorf("StatusHometimelineable: error getting replied to account with id %s: %s", targetStatus.InReplyToAccountID, err)
			}
			targetStatus.InReplyToAccount = ra
		}

		// if it's a reply to the timelineOwnerAccount, we don't need to check if the timelineOwnerAccount follows itself, just return true, they can see it
		if targetStatus.AccountID == timelineOwnerAccount.ID {
			return true, nil
		}

		// the replied-to account != timelineOwnerAccount, so make sure the timelineOwnerAccount follows the replied-to account
		follows, err := f.db.IsFollowing(ctx, timelineOwnerAccount, targetStatus.InReplyToAccount)
		if err != nil {
			return false, fmt.Errorf("StatusHometimelineable: error checking follow from account %s to account %s: %s", timelineOwnerAccount.ID, targetStatus.InReplyToAccountID, err)
		}

		// we don't want to timeline a reply to a status whose owner isn't followed by the requesting account
		if !follows {
			return false, nil
		}
	}

	return true, nil
}
