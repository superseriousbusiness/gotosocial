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

package bundb

import (
	"context"

	"codeberg.org/gruf/go-cache/v2"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type notificationDB struct {
	conn  *DBConn
	cache cache.Cache[string, *gtsmodel.Notification]
}

func (n *notificationDB) newNotificationQ(i interface{}) *bun.SelectQuery {
	return n.conn.
		NewSelect().
		Model(i).
		Relation("OriginAccount").
		Relation("TargetAccount").
		Relation("Status")
}

func (n *notificationDB) GetNotification(ctx context.Context, id string) (*gtsmodel.Notification, db.Error) {
	if notification, ok := n.cache.Get(id); ok {
		return notification, nil
	}
	notif := &gtsmodel.Notification{}
	err := n.getNotificationDB(ctx, id, notif)
	if err != nil {
		return nil, err
	}
	return notif, nil
}

func (n *notificationDB) GetNotifications(ctx context.Context, accountID string, limit int, maxID string, sinceID string) ([]*gtsmodel.Notification, db.Error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make a guess for slice size
	notifications := make([]*gtsmodel.Notification, 0, limit)

	q := n.conn.
		NewSelect().
		Model(&notifications).
		Column("id")

	if maxID != "" {
		q = q.Where("id < ?", maxID)
	}

	if sinceID != "" {
		q = q.Where("id > ?", sinceID)
	}

	q = q.
		Where("target_account_id = ?", accountID).
		Order("id DESC")

	if limit != 0 {
		q = q.Limit(limit)
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, n.conn.ProcessError(err)
	}

	// now we have the IDs, select the notifs one by one
	// reason for this is that for each notif, we can instead get it from our cache if it's cached
	for i, notif := range notifications {
		// Attempt fetch from DB
		notif, err := n.GetNotification(ctx, notif.ID)
		if err != nil {
			return nil, err
		}

		// Set notification
		notifications[i] = notif
	}

	return notifications, nil
}

func (n *notificationDB) getNotificationDB(ctx context.Context, id string, dst *gtsmodel.Notification) error {
	q := n.newNotificationQ(dst).WherePK()

	if err := q.Scan(ctx); err != nil {
		return n.conn.ProcessError(err)
	}

	n.cache.Set(id, dst)
	return nil
}
