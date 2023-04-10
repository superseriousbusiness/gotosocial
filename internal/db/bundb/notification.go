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
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type notificationDB struct {
	conn  *DBConn
	state *state.State
}

func (n *notificationDB) GetNotificationByID(ctx context.Context, id string) (*gtsmodel.Notification, db.Error) {
	return n.state.Caches.GTS.Notification().Load("ID", func() (*gtsmodel.Notification, error) {
		var notif gtsmodel.Notification

		q := n.conn.NewSelect().
			Model(&notif).
			Where("? = ?", bun.Ident("notification.id"), id)
		if err := q.Scan(ctx); err != nil {
			return nil, n.conn.ProcessError(err)
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
) (*gtsmodel.Notification, db.Error) {
	return n.state.Caches.GTS.Notification().Load("NotificationType.TargetAccountID.OriginAccountID.StatusID", func() (*gtsmodel.Notification, error) {
		var notif gtsmodel.Notification

		q := n.conn.NewSelect().
			Model(&notif).
			Where("? = ?", bun.Ident("notification_type"), notificationType).
			Where("? = ?", bun.Ident("target_account_id"), targetAccountID).
			Where("? = ?", bun.Ident("origin_account_id"), originAccountID).
			Where("? = ?", bun.Ident("status_id"), statusID)

		if err := q.Scan(ctx); err != nil {
			return nil, n.conn.ProcessError(err)
		}

		return &notif, nil
	}, notificationType, targetAccountID, originAccountID, statusID)
}

func (n *notificationDB) GetAccountNotifications(ctx context.Context, accountID string, excludeTypes []string, limit int, maxID string, sinceID string) ([]*gtsmodel.Notification, db.Error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make a guess for slice size
	notifIDs := make([]string, 0, limit)

	q := n.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("notifications"), bun.Ident("notification")).
		Column("notification.id")

	if maxID != "" {
		q = q.Where("? < ?", bun.Ident("notification.id"), maxID)
	}

	if sinceID != "" {
		q = q.Where("? > ?", bun.Ident("notification.id"), sinceID)
	}

	for _, excludeType := range excludeTypes {
		q = q.Where("? != ?", bun.Ident("notification.notification_type"), excludeType)
	}

	q = q.
		Where("? = ?", bun.Ident("notification.target_account_id"), accountID).
		Order("notification.id DESC")

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &notifIDs); err != nil {
		return nil, n.conn.ProcessError(err)
	}

	notifs := make([]*gtsmodel.Notification, 0, limit)

	// now we have the IDs, select the notifs one by one
	// reason for this is that for each notif, we can instead get it from our cache if it's cached
	for _, id := range notifIDs {
		// Attempt fetch from DB
		notif, err := n.GetNotificationByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting notification %q: %v", id, err)
			continue
		}

		// Append notification
		notifs = append(notifs, notif)
	}

	return notifs, nil
}

func (n *notificationDB) PutNotification(ctx context.Context, notif *gtsmodel.Notification) error {
	return n.state.Caches.GTS.Notification().Store(notif, func() error {
		_, err := n.conn.NewInsert().Model(notif).Exec(ctx)
		return n.conn.ProcessError(err)
	})
}

func (n *notificationDB) DeleteNotificationByID(ctx context.Context, id string) db.Error {
	if _, err := n.conn.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("notifications"), bun.Ident("notification")).
		Where("? = ?", bun.Ident("notification.id"), id).
		Exec(ctx); err != nil {
		return n.conn.ProcessError(err)
	}

	n.state.Caches.GTS.Notification().Invalidate("ID", id)
	return nil
}

func (n *notificationDB) DeleteNotifications(ctx context.Context, types []string, targetAccountID string, originAccountID string) db.Error {
	if targetAccountID == "" && originAccountID == "" {
		return errors.New("DeleteNotifications: one of targetAccountID or originAccountID must be set")
	}

	// Capture notification IDs in a RETURNING statement.
	var ids []string

	q := n.conn.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("notifications"), bun.Ident("notification")).
		Returning("?", bun.Ident("id"))

	if len(types) > 0 {
		q = q.Where("? IN (?)", bun.Ident("notification.notification_type"), bun.In(types))
	}

	if targetAccountID != "" {
		q = q.Where("? = ?", bun.Ident("notification.target_account_id"), targetAccountID)
	}

	if originAccountID != "" {
		q = q.Where("? = ?", bun.Ident("notification.origin_account_id"), originAccountID)
	}

	if _, err := q.Exec(ctx, &ids); err != nil {
		return n.conn.ProcessError(err)
	}

	// Invalidate each returned ID.
	for _, id := range ids {
		n.state.Caches.GTS.Notification().Invalidate("ID", id)
	}

	return nil
}

func (n *notificationDB) DeleteNotificationsForStatus(ctx context.Context, statusID string) db.Error {
	// Capture notification IDs in a RETURNING statement.
	var ids []string

	q := n.conn.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("notifications"), bun.Ident("notification")).
		Where("? = ?", bun.Ident("notification.status_id"), statusID).
		Returning("?", bun.Ident("id"))

	if _, err := q.Exec(ctx, &ids); err != nil {
		return n.conn.ProcessError(err)
	}

	// Invalidate each returned ID.
	for _, id := range ids {
		n.state.Caches.GTS.Notification().Invalidate("ID", id)
	}

	return nil
}
