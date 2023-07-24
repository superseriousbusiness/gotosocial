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

package bundb

import (
	"context"
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type notificationDB struct {
	db    *WrappedDB
	state *state.State
}

func (n *notificationDB) GetNotificationByID(ctx context.Context, id string) (*gtsmodel.Notification, error) {
	return n.state.Caches.GTS.Notification().Load("ID", func() (*gtsmodel.Notification, error) {
		var notif gtsmodel.Notification

		q := n.db.NewSelect().
			Model(&notif).
			Where("? = ?", bun.Ident("notification.id"), id)
		if err := q.Scan(ctx); err != nil {
			return nil, n.db.ProcessError(err)
		}

		return &notif, nil
	}, id)
}

func (n *notificationDB) GetNotification(
	ctx context.Context,
	notificationType gtsmodel.NotificationType,
	targetAccountID string,
	originAccountID string,
	statusID string,
) (*gtsmodel.Notification, error) {
	return n.state.Caches.GTS.Notification().Load("NotificationType.TargetAccountID.OriginAccountID.StatusID", func() (*gtsmodel.Notification, error) {
		var notif gtsmodel.Notification

		q := n.db.NewSelect().
			Model(&notif).
			Where("? = ?", bun.Ident("notification_type"), notificationType).
			Where("? = ?", bun.Ident("target_account_id"), targetAccountID).
			Where("? = ?", bun.Ident("origin_account_id"), originAccountID).
			Where("? = ?", bun.Ident("status_id"), statusID)

		if err := q.Scan(ctx); err != nil {
			return nil, n.db.ProcessError(err)
		}

		return &notif, nil
	}, notificationType, targetAccountID, originAccountID, statusID)
}

func (n *notificationDB) GetAccountNotifications(
	ctx context.Context,
	accountID string,
	maxID string,
	sinceID string,
	minID string,
	limit int,
	excludeTypes []string,
) ([]*gtsmodel.Notification, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	var (
		notifIDs    = make([]string, 0, limit)
		frontToBack = true
	)

	q := n.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("notifications"), bun.Ident("notification")).
		Column("notification.id")

	if maxID == "" {
		maxID = id.Highest
	}

	// Return only notifs LOWER (ie., older) than maxID.
	q = q.Where("? < ?", bun.Ident("notification.id"), maxID)

	if sinceID != "" {
		// Return only notifs HIGHER (ie., newer) than sinceID.
		q = q.Where("? > ?", bun.Ident("notification.id"), sinceID)
	}

	if minID != "" {
		// Return only notifs HIGHER (ie., newer) than minID.
		q = q.Where("? > ?", bun.Ident("notification.id"), minID)

		frontToBack = false // page up
	}

	for _, excludeType := range excludeTypes {
		// Filter out unwanted notif types.
		q = q.Where("? != ?", bun.Ident("notification.notification_type"), excludeType)
	}

	// Return only notifs for this account.
	q = q.Where("? = ?", bun.Ident("notification.target_account_id"), accountID)

	if limit > 0 {
		q = q.Limit(limit)
	}

	if frontToBack {
		// Page down.
		q = q.Order("notification.id DESC")
	} else {
		// Page up.
		q = q.Order("notification.id ASC")
	}

	if err := q.Scan(ctx, &notifIDs); err != nil {
		return nil, n.db.ProcessError(err)
	}

	if len(notifIDs) == 0 {
		return nil, nil
	}

	// If we're paging up, we still want notifications
	// to be sorted by ID desc, so reverse ids slice.
	// https://zchee.github.io/golang-wiki/SliceTricks/#reversing
	if !frontToBack {
		for l, r := 0, len(notifIDs)-1; l < r; l, r = l+1, r-1 {
			notifIDs[l], notifIDs[r] = notifIDs[r], notifIDs[l]
		}
	}

	notifs := make([]*gtsmodel.Notification, 0, len(notifIDs))
	for _, id := range notifIDs {
		// Attempt fetch from DB
		notif, err := n.GetNotificationByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error fetching notification %q: %v", id, err)
			continue
		}

		// Append notification
		notifs = append(notifs, notif)
	}

	return notifs, nil
}

func (n *notificationDB) PutNotification(ctx context.Context, notif *gtsmodel.Notification) error {
	return n.state.Caches.GTS.Notification().Store(notif, func() error {
		_, err := n.db.NewInsert().Model(notif).Exec(ctx)
		return n.db.ProcessError(err)
	})
}

func (n *notificationDB) DeleteNotificationByID(ctx context.Context, id string) error {
	defer n.state.Caches.GTS.Notification().Invalidate("ID", id)

	// Load notif into cache before attempting a delete,
	// as we need it cached in order to trigger the invalidate
	// callback. This in turn invalidates others.
	_, err := n.GetNotificationByID(gtscontext.SetBarebones(ctx), id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// not an issue.
			err = nil
		}
		return err
	}

	// Finally delete notif from DB.
	_, err = n.db.NewDelete().
		TableExpr("? AS ?", bun.Ident("notifications"), bun.Ident("notification")).
		Where("? = ?", bun.Ident("notification.id"), id).
		Exec(ctx)
	return n.db.ProcessError(err)
}

func (n *notificationDB) DeleteNotifications(ctx context.Context, types []string, targetAccountID string, originAccountID string) error {
	if targetAccountID == "" && originAccountID == "" {
		return errors.New("DeleteNotifications: one of targetAccountID or originAccountID must be set")
	}

	var notifIDs []string

	q := n.db.
		NewSelect().
		Column("id").
		Table("notifications")

	if len(types) > 0 {
		q = q.Where("? IN (?)", bun.Ident("notification_type"), bun.In(types))
	}

	if targetAccountID != "" {
		q = q.Where("? = ?", bun.Ident("target_account_id"), targetAccountID)
	}

	if originAccountID != "" {
		q = q.Where("? = ?", bun.Ident("origin_account_id"), originAccountID)
	}

	if _, err := q.Exec(ctx, &notifIDs); err != nil {
		return n.db.ProcessError(err)
	}

	defer func() {
		// Invalidate all IDs on return.
		for _, id := range notifIDs {
			n.state.Caches.GTS.Notification().Invalidate("ID", id)
		}
	}()

	// Load all notif into cache, this *really* isn't great
	// but it is the only way we can ensure we invalidate all
	// related caches correctly (e.g. visibility).
	for _, id := range notifIDs {
		_, err := n.GetNotificationByID(ctx, id)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return err
		}
	}

	// Finally delete all from DB.
	_, err := n.db.NewDelete().
		Table("notifications").
		Where("? IN (?)", bun.Ident("id"), bun.In(notifIDs)).
		Exec(ctx)
	return n.db.ProcessError(err)
}

func (n *notificationDB) DeleteNotificationsForStatus(ctx context.Context, statusID string) error {
	var notifIDs []string

	q := n.db.
		NewSelect().
		Column("id").
		Table("notifications").
		Where("? = ?", bun.Ident("status_id"), statusID)

	if _, err := q.Exec(ctx, &notifIDs); err != nil {
		return n.db.ProcessError(err)
	}

	defer func() {
		// Invalidate all IDs on return.
		for _, id := range notifIDs {
			n.state.Caches.GTS.Notification().Invalidate("ID", id)
		}
	}()

	// Load all notif into cache, this *really* isn't great
	// but it is the only way we can ensure we invalidate all
	// related caches correctly (e.g. visibility).
	for _, id := range notifIDs {
		_, err := n.GetNotificationByID(ctx, id)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return err
		}
	}

	// Finally delete all from DB.
	_, err := n.db.NewDelete().
		Table("notifications").
		Where("? IN (?)", bun.Ident("id"), bun.In(notifIDs)).
		Exec(ctx)
	return n.db.ProcessError(err)
}
