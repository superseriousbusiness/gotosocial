/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package message

import (
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) notifyStatus(status *gtsmodel.Status) error {
	return nil
}

func (p *processor) notifyFollowRequest(followRequest *gtsmodel.FollowRequest, receivingAccount *gtsmodel.Account) error {
	// return if this isn't a local account
	if receivingAccount.Domain != "" {
		return nil
	}

	notif := &gtsmodel.Notification{
		NotificationType: gtsmodel.NotificationFollowRequest,
		TargetAccountID:  followRequest.TargetAccountID,
		OriginAccountID:  followRequest.AccountID,
	}

	if err := p.db.Put(notif); err != nil {
		return fmt.Errorf("notifyFollowRequest: error putting notification in database: %s", err)
	}

	return nil
}

func (p *processor) notifyFollow(follow *gtsmodel.Follow, receivingAccount *gtsmodel.Account) error {
	// return if this isn't a local account
	if receivingAccount.Domain != "" {
		return nil
	}

	// first remove the follow request notification
	if err := p.db.DeleteWhere([]db.Where{
		{Key: "notification_type", Value: gtsmodel.NotificationFollowRequest},
		{Key: "target_account_id", Value: follow.TargetAccountID},
		{Key: "origin_account_id", Value: follow.AccountID},
	}, &gtsmodel.Notification{}); err != nil {
		return fmt.Errorf("notifyFollow: error removing old follow request notification from database: %s", err)
	}

	// now create the new follow notification
	notif := &gtsmodel.Notification{
		NotificationType: gtsmodel.NotificationFollow,
		TargetAccountID:  follow.TargetAccountID,
		OriginAccountID:  follow.AccountID,
	}
	if err := p.db.Put(notif); err != nil {
		return fmt.Errorf("notifyFollow: error putting notification in database: %s", err)
	}

	return nil
}

func (p *processor) notifyFave(fave *gtsmodel.StatusFave, receivingAccount *gtsmodel.Account) error {
	// return if this isn't a local account
	if receivingAccount.Domain != "" {
		return nil
	}

	notif := &gtsmodel.Notification{
		NotificationType: gtsmodel.NotificationFave,
		TargetAccountID:  fave.TargetAccountID,
		OriginAccountID:  fave.AccountID,
		StatusID:         fave.StatusID,
	}

	if err := p.db.Put(notif); err != nil {
		return fmt.Errorf("notifyFave: error putting notification in database: %s", err)
	}

	return nil
}
