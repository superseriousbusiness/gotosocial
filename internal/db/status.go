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
)

// Status contains functions for getting statuses, creating statuses, and checking various other fields on statuses.
type Status interface {
	// GetStatusByID returns one status from the database, with no rel fields populated, only their linking ID / URIs
	GetStatusByID(ctx context.Context, id string) (*gtsmodel.Status, error)

	// GetStatusByURI returns one status from the database, with no rel fields populated, only their linking ID / URIs
	GetStatusByURI(ctx context.Context, uri string) (*gtsmodel.Status, error)

	// GetStatusByURL returns one status from the database, with no rel fields populated, only their linking ID / URIs
	GetStatusByURL(ctx context.Context, uri string) (*gtsmodel.Status, error)

	// GetStatusBoost ...
	GetStatusBoost(ctx context.Context, boostOfID string, byAccountID string) (*gtsmodel.Status, error)

	// PopulateStatus ensures that all sub-models of a status are populated (e.g. mentions, attachments, etc).
	PopulateStatus(ctx context.Context, status *gtsmodel.Status) error

	// PutStatus stores one status in the database.
	PutStatus(ctx context.Context, status *gtsmodel.Status) error

	// UpdateStatus updates one status in the database.
	UpdateStatus(ctx context.Context, status *gtsmodel.Status, columns ...string) error

	// DeleteStatusByID deletes one status from the database.
	DeleteStatusByID(ctx context.Context, id string) error

	// GetStatuses gets a slice of statuses corresponding to the given status IDs.
	GetStatusesByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Status, error)

	// GetStatusesUsingEmoji fetches all status models using emoji with given ID stored in their 'emojis' column.
	GetStatusesUsingEmoji(ctx context.Context, emojiID string) ([]*gtsmodel.Status, error)

	// GetStatusReplies ...
	GetStatusReplies(ctx context.Context, statusID string) ([]*gtsmodel.Status, error)

	// CountStatusReplies ...
	CountStatusReplies(ctx context.Context, statusID string) (int, error)

	// GetStatusBoosts ...
	GetStatusBoosts(ctx context.Context, statusID string) ([]*gtsmodel.Status, error)

	// CountStatusBoosts ...
	CountStatusBoosts(ctx context.Context, statusID string) (int, error)

	// IsStatusBoostedBy ...
	IsStatusBoostedBy(ctx context.Context, statusID string, accountID string) (bool, error)

	// GetStatusParents gets the parent statuses of a given status.
	//
	// If onlyDirect is true, only the immediate parent will be returned.
	GetStatusParents(ctx context.Context, status *gtsmodel.Status, onlyDirect bool) ([]*gtsmodel.Status, error)

	// GetStatusChildren gets the child statuses of a given status.
	//
	// If onlyDirect is true, only the immediate children will be returned.
	GetStatusChildren(ctx context.Context, status *gtsmodel.Status, onlyDirect bool, minID string) ([]*gtsmodel.Status, error)

	// IsStatusMutedBy checks if a given status has been muted by a given account ID
	IsStatusMutedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, error)

	// IsStatusBookmarkedBy checks if a given status has been bookmarked by a given account ID
	IsStatusBookmarkedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, error)
}
