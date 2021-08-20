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

package pg

import (
	"context"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type notificationDB struct {
	config *config.Config
	conn   *pg.DB
	log    *logrus.Logger
	cancel context.CancelFunc
	cache  cache.Cache
}

func (n *notificationDB) cacheNotification(id string, notification *gtsmodel.Notification) {
	if n.cache == nil {
		n.cache = cache.New()
	}

	if err := n.cache.Store(id, notification); err != nil {
		n.log.Panicf("notificationDB: error storing in cache: %s", err)
	}
}

func (n *notificationDB) notificationCached(id string) (*gtsmodel.Notification, bool) {
	if n.cache == nil {
		n.cache = cache.New()
		return nil, false
	}

	nI, err := n.cache.Fetch(id)
	if err != nil || nI == nil {
		return nil, false
	}

	notification, ok := nI.(*gtsmodel.Notification)
	if !ok {
		n.log.Panicf("notificationDB: cached interface with key %s was not a notification", id)
	}

	return notification, true
}

func (n *notificationDB) newNotificationQ(i interface{}) *orm.Query {
	return n.conn.Model(i).
		Relation("OriginAccount").
		Relation("TargetAccount").
		Relation("Status")
}

func (n *notificationDB) GetNotification(id string) (*gtsmodel.Notification, db.Error) {
	if notification, cached := n.notificationCached(id); cached {
		return notification, nil
	}

	notification := &gtsmodel.Notification{}

	q := n.newNotificationQ(notification).
		Where("notification.id = ?", id)

	err := processErrorResponse(q.Select())

	if err == nil && notification != nil {
		n.cacheNotification(id, notification)
	}

	return notification, err
}

func (n *notificationDB) GetNotifications(accountID string, limit int, maxID string, sinceID string) ([]*gtsmodel.Notification, db.Error) {
	// begin by selecting just the IDs
	notifIDs := []*gtsmodel.Notification{}
	q := n.conn.
		Model(&notifIDs).
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

	err := processErrorResponse(q.Select())
	if err != nil {
		return nil, err
	}

	// now we have the IDs, select the notifs one by one
	// reason for this is that for each notif, we can instead get it from our cache if it's cached
	notifications := []*gtsmodel.Notification{}
	for _, notifID := range notifIDs {
		notif, err := n.GetNotification(notifID.ID)
		errP := processErrorResponse(err)
		if errP != nil {
			return nil, errP
		}
		notifications = append(notifications, notif)
	}

	return notifications, nil
}
