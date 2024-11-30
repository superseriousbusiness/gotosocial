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
	"errors"
	"fmt"
	"net/url"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *Processor) NotificationsGet(
	ctx context.Context,
	authed *oauth.Auth,
	page *paging.Page,
	types []gtsmodel.NotificationType,
	excludeTypes []gtsmodel.NotificationType,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	notifs, err := p.state.DB.GetAccountNotifications(
		ctx,
		authed.Account.ID,
		page,
		types,
		excludeTypes,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("NotificationsGet: db error getting notifications: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(notifs)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	filters, err := p.state.DB.GetFiltersForAccountID(ctx, authed.Account.ID)
	if err != nil {
		err = gtserror.Newf("couldn't retrieve filters for account %s: %w", authed.Account.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	mutes, err := p.state.DB.GetAccountMutes(gtscontext.SetBarebones(ctx), authed.Account.ID, nil)
	if err != nil {
		err = gtserror.Newf("couldn't retrieve mutes for account %s: %w", authed.Account.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	compiledMutes := usermute.NewCompiledUserMuteList(mutes)

	var (
		items = make([]interface{}, 0, count)

		// Get the lowest and highest
		// ID values, used for paging.
		lo = notifs[count-1].ID
		hi = notifs[0].ID
	)

	for _, n := range notifs {
		visible, err := p.notifVisible(ctx, n, authed.Account)
		if err != nil {
			log.Debugf(ctx, "skipping notification %s because of an error checking notification visibility: %v", n.ID, err)
			continue
		}

		if !visible {
			continue
		}

		item, err := p.converter.NotificationToAPINotification(ctx, n, filters, compiledMutes)
		if err != nil {
			if !errors.Is(err, status.ErrHideStatus) {
				log.Debugf(ctx, "skipping notification %s because it couldn't be converted to its api representation: %s", n.ID, err)
			}
			continue
		}

		items = append(items, item)
	}

	// Build type query string.
	query := make(url.Values)
	for _, typ := range types {
		query.Add("types[]", typ.String())
	}
	for _, typ := range excludeTypes {
		query.Add("exclude_types[]", typ.String())
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/notifications",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: query,
	}), nil
}

func (p *Processor) NotificationGet(ctx context.Context, account *gtsmodel.Account, targetNotifID string) (*apimodel.Notification, gtserror.WithCode) {
	notif, err := p.state.DB.GetNotificationByID(ctx, targetNotifID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}

		// Real error.
		return nil, gtserror.NewErrorInternalError(err)
	}

	if notifTargetAccountID := notif.TargetAccountID; notifTargetAccountID != account.ID {
		err = fmt.Errorf("account %s does not have permission to view notification belong to account %s", account.ID, notifTargetAccountID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	filters, err := p.state.DB.GetFiltersForAccountID(ctx, account.ID)
	if err != nil {
		err = gtserror.Newf("couldn't retrieve filters for account %s: %w", account.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	mutes, err := p.state.DB.GetAccountMutes(gtscontext.SetBarebones(ctx), account.ID, nil)
	if err != nil {
		err = gtserror.Newf("couldn't retrieve mutes for account %s: %w", account.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	compiledMutes := usermute.NewCompiledUserMuteList(mutes)

	apiNotif, err := p.converter.NotificationToAPINotification(ctx, notif, filters, compiledMutes)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}

		// Real error.
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiNotif, nil
}

func (p *Processor) NotificationsClear(ctx context.Context, authed *oauth.Auth) gtserror.WithCode {
	// Delete all notifications of all types that target the authorized account.
	if err := p.state.DB.DeleteNotifications(ctx, nil, authed.Account.ID, ""); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}

func (p *Processor) notifVisible(
	ctx context.Context,
	n *gtsmodel.Notification,
	acct *gtsmodel.Account,
) (bool, error) {
	// If account is set, ensure it's
	// visible to notif target.
	if n.OriginAccount != nil {
		// If this is a new local account sign-up,
		// skip normal visibility checking because
		// origin account won't be confirmed yet.
		if n.NotificationType == gtsmodel.NotificationAdminSignup {
			return true, nil
		}

		visible, err := p.visFilter.AccountVisible(ctx, acct, n.OriginAccount)
		if err != nil {
			return false, err
		}

		if !visible {
			return false, nil
		}
	}

	// If status is set, ensure it's
	// visible to notif target.
	if n.Status != nil {
		visible, err := p.visFilter.StatusVisible(ctx, acct, n.Status)
		if err != nil {
			return false, err
		}

		if !visible {
			return false, nil
		}
	}

	return true, nil
}
