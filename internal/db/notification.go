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

package db

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Notification contains functions for creating and getting notifications.
type Notification interface {
	// GetNotifications returns a slice of notifications that pertain to the given accountID.
	//
	// Returned notifications will be ordered ID descending (ie., highest/newest to lowest/oldest).
	GetNotifications(ctx context.Context, accountID string, excludeTypes []string, limit int, maxID string, sinceID string) ([]*gtsmodel.Notification, Error)

	// GetNotification returns one notification according to its id.
	GetNotification(ctx context.Context, id string) (*gtsmodel.Notification, Error)

	// DeleteNotification deletes one notification according to its id,
	// and removes that notification from the in-memory cache.
	DeleteNotification(ctx context.Context, id string) Error

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
	DeleteNotifications(ctx context.Context, targetAccountID string, originAccountID string) Error

	// DeleteNotificationsForStatus deletes all notifications that relate to
	// the given statusID. This function is useful when a status has been deleted,
	// and so notifications relating to that status must also be deleted.
	DeleteNotificationsForStatus(ctx context.Context, statusID string) Error
}
