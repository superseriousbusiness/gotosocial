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
	// ClearNotifications deletes every notification that pertain to the given accountID.
	ClearNotifications(ctx context.Context, accountID string) Error
}
