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
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type notificationDB struct {
	conn  *DBConn
	cache cache.Cache[string, *gtsmodel.Notification]
}

func (n *notificationDB) GetNotification(ctx context.Context, id string) (*gtsmodel.Notification, db.Error) {
	if notification, ok := n.cache.Get(id); ok {
		return notification, nil
	}

	dst := gtsmodel.Notification{ID: id}

	q := n.conn.NewSelect().
		Model(&dst).
		Relation("OriginAccount").
		Relation("TargetAccount").
		Relation("Status").
		WherePK()

	if err := q.Scan(ctx); err != nil {
		return nil, n.conn.ProcessError(err)
	}

	copy := dst
	n.cache.Set(id, &copy)

	return &dst, nil
}

func (n *notificationDB) GetNotifications(ctx context.Context, accountID string, limit int, maxID string, sinceID string) ([]*gtsmodel.Notification, db.Error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make a guess for slice size
	notifIDs := make([]string, 0, limit)

	q := n.conn.
		NewSelect().
		Table("notifications").
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

	if err := q.Scan(ctx, &notifIDs); err != nil {
		return nil, n.conn.ProcessError(err)
	}

	notifs := make([]*gtsmodel.Notification, 0, limit)

	// now we have the IDs, select the notifs one by one
	// reason for this is that for each notif, we can instead get it from our cache if it's cached
	for _, id := range notifIDs {
		// Attempt fetch from DB
		notif, err := n.GetNotification(ctx, id)
		if err != nil {
			logrus.Errorf("GetNotifications: error getting notification %q: %v", id, err)
			continue
		}

		// Append notification
		notifs = append(notifs, notif)
	}

	return notifs, nil
}
