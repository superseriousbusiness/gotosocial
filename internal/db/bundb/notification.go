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

package bundb

import (
	"context"

	"github.com/ReneKroon/ttlcache"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type notificationDB struct {
	config *config.Config
	conn   *DBConn
	cache  *ttlcache.Cache
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
	if notification, cached := n.getNotificationCache(id); cached {
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
		Column("id").
		Where("target_account_id = ?", accountID).
		Order("id DESC")

	if maxID != "" {
		q = q.Where("id < ?", maxID)
	}

	if sinceID != "" {
		q = q.Where("id > ?", sinceID)
	}

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
		// Check cache for notification
		nn, cached := n.getNotificationCache(notif.ID)
		if cached {
			notifications[i] = nn
			continue
		}

		// Check DB for notification
		err := n.getNotificationDB(ctx, notif.ID, notif)
		if err != nil {
			return nil, err
		}
	}

	return notifications, nil
}

func (n *notificationDB) getNotificationCache(id string) (*gtsmodel.Notification, bool) {
	v, ok := n.cache.Get(id)
	if !ok {
		return nil, false
	}
	return v.(*gtsmodel.Notification), true
}

func (n *notificationDB) putNotificationCache(notif *gtsmodel.Notification) {
	n.cache.Set(notif.ID, notif)
}

func (n *notificationDB) getNotificationDB(ctx context.Context, id string, dst *gtsmodel.Notification) error {
	q := n.newNotificationQ(dst).
		Where("notification.id = ?", id)

	err := q.Scan(ctx)
	if err != nil {
		return n.conn.ProcessError(err)
	}

	n.putNotificationCache(dst)
	return nil
}
