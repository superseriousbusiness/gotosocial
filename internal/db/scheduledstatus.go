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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

type ScheduledStatus interface {
	// GetAllScheduledStatuses returns all pending scheduled statuses.
	GetAllScheduledStatuses(ctx context.Context) ([]*gtsmodel.ScheduledStatus, error)

	// GetScheduledStatusByID gets one scheduled status with the given id.
	GetScheduledStatusByID(ctx context.Context, id string) (*gtsmodel.ScheduledStatus, error)

	// GetScheduledStatusesForAcct returns statuses scheduled by the given account.
	GetScheduledStatusesForAcct(
		ctx context.Context,
		acctID string,
		page *paging.Page,
	) ([]*gtsmodel.ScheduledStatus, error)

	// PutScheduledStatus puts the given scheduled status in the database.
	PutScheduledStatus(ctx context.Context, status *gtsmodel.ScheduledStatus) error

	// DeleteScheduledStatusByID deletes one scheduled status from the database.
	DeleteScheduledStatusByID(ctx context.Context, id string) error

	// DeleteScheduledStatusByID deletes all scheduled statuses from an account from the database.
	DeleteScheduledStatusesByAccountID(ctx context.Context, accountID string) error

	// DeleteScheduledStatusesByApplicationID deletes all scheduled statuses posted from the given application from the database.
	DeleteScheduledStatusesByApplicationID(ctx context.Context, applicationID string) error

	// UpdateScheduledStatusScheduledDate updates `scheduled_at` param for the given scheduled status in the database.
	UpdateScheduledStatusScheduledDate(ctx context.Context, scheduledStatus *gtsmodel.ScheduledStatus, scheduledAt *time.Time) error

	// GetScheduledStatusesCountForAcct returns the number of pending statuses scheduled by the given account, optionally for a specific day.
	GetScheduledStatusesCountForAcct(ctx context.Context, acctID string, scheduledAt *time.Time) (int, error)
}
