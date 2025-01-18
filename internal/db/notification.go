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

package db

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// Notification contains functions for creating and getting notifications.
type Notification interface {
	// GetAccountNotifications returns a slice of notifications that pertain to the given accountID.
	//
	// Returned notifications will be ordered ID descending (ie., highest/newest to lowest/oldest).
	// If types is empty, *all* notification types will be included.
	GetAccountNotifications(ctx context.Context, accountID string, page *paging.Page, types []gtsmodel.NotificationType, excludeTypes []gtsmodel.NotificationType) ([]*gtsmodel.Notification, error)

	// GetNotificationByID returns one notification according to its id.
	GetNotificationByID(ctx context.Context, id string) (*gtsmodel.Notification, error)

	// GetNotificationsByIDs returns a slice of notifications of the the provided IDs.
	GetNotificationsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Notification, error)

	// GetNotification gets one notification according to the provided parameters, if it exists.
	// Since not all notifications are about a status, statusID can be an empty string.
	GetNotification(ctx context.Context, notificationType gtsmodel.NotificationType, targetAccountID string, originAccountID string, statusID string) (*gtsmodel.Notification, error)

	// PopulateNotification ensures that the notification's struct fields are populated.
	PopulateNotification(ctx context.Context, notif *gtsmodel.Notification) error

	// PutNotification will insert the given notification into the database.
	PutNotification(ctx context.Context, notif *gtsmodel.Notification) error

	// DeleteNotificationByID deletes one notification according to its id,
	// and removes that notification from the in-memory cache.
	DeleteNotificationByID(ctx context.Context, id string) error

	// DeleteNotifications mass deletes notifications targeting targetAccountID
	// and/or originating from originAccountID.
	//
	// If targetAccountID is set and originAccountID isn't, all notifications
	// that target the given account will be deleted.
	//
	// If originAccountID is set and targetAccountID isn't, all notifications
	// originating from the given account will be deleted.
	//
	// If both are set, then notifications that target targetAccountID and
	// originate from originAccountID will be deleted.
	//
	// At least one parameter must not be an empty string.
	DeleteNotifications(ctx context.Context, types []gtsmodel.NotificationType, targetAccountID string, originAccountID string) error

	// DeleteNotificationsForStatus deletes all notifications that relate to
	// the given statusID. This function is useful when a status has been deleted,
	// and so notifications relating to that status must also be deleted.
	DeleteNotificationsForStatus(ctx context.Context, statusID string) error
}
