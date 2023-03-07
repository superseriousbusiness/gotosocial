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

package processing

import (
	"context"
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *Processor) NotificationsGet(ctx context.Context, authed *oauth.Auth, excludeTypes []string, limit int, maxID string, sinceID string) (*apimodel.PageableResponse, gtserror.WithCode) {
	notifs, err := p.state.DB.GetNotifications(ctx, authed.Account.ID, excludeTypes, limit, maxID, sinceID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(notifs)

	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	items := make([]interface{}, 0, count)
	nextMaxIDValue := ""
	prevMinIDValue := ""
	for i, n := range notifs {
		item, err := p.tc.NotificationToAPINotification(ctx, n)
		if err != nil {
			log.Debugf(ctx, "got an error converting a notification to api, will skip it: %s", err)
			continue
		}

		if i == count-1 {
			nextMaxIDValue = item.GetID()
		}

		if i == 0 {
			prevMinIDValue = item.GetID()
		}

		items = append(items, item)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "api/v1/notifications",
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDKey:   "since_id",
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}

func (p *Processor) NotificationsClear(ctx context.Context, authed *oauth.Auth) gtserror.WithCode {
	// Delete all notifications that target the authorized account.
	if err := p.state.DB.DeleteNotifications(ctx, authed.Account.ID, "", ""); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}
