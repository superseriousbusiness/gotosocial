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

package processing

import (
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) NotificationsGet(authed *oauth.Auth, limit int, maxID string, sinceID string) ([]*apimodel.Notification, ErrorWithCode) {
	l := p.log.WithField("func", "NotificationsGet")

	notifs, err := p.db.GetNotificationsForAccount(authed.Account.ID, limit, maxID, sinceID)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	mastoNotifs := []*apimodel.Notification{}
	for _, n := range notifs {
		mastoNotif, err := p.tc.NotificationToMasto(n)
		if err != nil {
			l.Debugf("got an error converting a notification to masto, will skip it: %s", err)
			continue
		}
		mastoNotifs = append(mastoNotifs, mastoNotif)
	}

	return mastoNotifs, nil
}
