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

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) notifyStatus(status *gtsmodel.Status) error {
	return nil
}

func (p *processor) notifyFollow(follow *gtsmodel.Follow) error {
	return nil
}

func (p *processor) notifyFave(fave *gtsmodel.StatusFave) error {

	notif := &gtsmodel.Notification{
		NotificationType: gtsmodel.NotificationFave,
		TargetAccountID:  fave.TargetAccountID,
		OriginAccountID:  fave.AccountID,
		StatusID:         fave.StatusID,
	}

	if err := p.db.Put(notif); err != nil {
		return fmt.Errorf("notifyFave: error putting fave in database: %s", err)
	}

	return nil
}
