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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Status contains functions for getting statuses, creating statuses, and checking various other fields on statuses.
type Status interface {
	// GetStatusByID fetches the status from the database with matching id column.
	GetStatusByID(ctx context.Context, id string) (*gtsmodel.Status, error)

	// GetStatusByURI fetches the status from the database with matching uri column.
	GetStatusByURI(ctx context.Context, uri string) (*gtsmodel.Status, error)

	// GetStatusByURL fetches the status from the database with matching url column.
	GetStatusByURL(ctx context.Context, uri string) (*gtsmodel.Status, error)

	// GetStatusByPollID fetches the status from the database with matching poll_id column.
	GetStatusByPollID(ctx context.Context, pollID string) (*gtsmodel.Status, error)

	// GetStatusBoost fetches the status whose boost_of_id column refers to boostOfID, authored by given account ID.
	GetStatusBoost(ctx context.Context, boostOfID string, byAccountID string) (*gtsmodel.Status, error)

	// PopulateStatus ensures that all sub-models of a status are populated (e.g. mentions, attachments, etc).
	// Except for edits, to fetch these please call PopulateStatusEdits() .
	PopulateStatus(ctx context.Context, status *gtsmodel.Status) error

	// PopulateStatusEdits ensures that status' edits are fully popualted.
	PopulateStatusEdits(ctx context.Context, status *gtsmodel.Status) error

	// PutStatus stores one status in the database, this also handles status threading.
	PutStatus(ctx context.Context, status *gtsmodel.Status) error

	// UpdateStatus updates one status in the database.
	UpdateStatus(ctx context.Context, status *gtsmodel.Status, columns ...string) error

	// DeleteStatusByID deletes one status from the database.
	DeleteStatusByID(ctx context.Context, id string) error

	// GetStatuses gets a slice of statuses corresponding to the given status IDs.
	GetStatusesByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Status, error)

	// GetStatusesUsingEmoji fetches all status models using emoji with given ID stored in their 'emojis' column.
	GetStatusesUsingEmoji(ctx context.Context, emojiID string) ([]*gtsmodel.Status, error)

	// GetStatusReplies returns the *direct* (i.e. in_reply_to_id column) replies to this status ID, ordered DESC by ID.
	GetStatusReplies(ctx context.Context, statusID string) ([]*gtsmodel.Status, error)

	// CountStatusReplies returns the number of stored *direct* (i.e. in_reply_to_id column) replies to this status ID.
	CountStatusReplies(ctx context.Context, statusID string) (int, error)

	// GetStatusBoosts returns all statuses whose boost_of_id column refer to given status ID.
	GetStatusBoosts(ctx context.Context, statusID string) ([]*gtsmodel.Status, error)

	// CountStatusBoosts returns the number of stored boosts for status ID.
	CountStatusBoosts(ctx context.Context, statusID string) (int, error)

	// IsStatusBoostedBy checks whether the given status ID is boosted by account ID.
	IsStatusBoostedBy(ctx context.Context, statusID string, accountID string) (bool, error)

	// GetStatusParents gets the parent statuses of a given status.
	GetStatusParents(ctx context.Context, status *gtsmodel.Status) ([]*gtsmodel.Status, error)

	// GetStatusChildren gets the child statuses of a given status.
	GetStatusChildren(ctx context.Context, statusID string) ([]*gtsmodel.Status, error)

	// MaxDirectStatusID returns the newest ID across all DM statuses.
	// Returns the empty string with no error if there are no DM statuses yet.
	// It is used only by the conversation advanced migration.
	MaxDirectStatusID(ctx context.Context) (string, error)

	// GetDirectStatusIDsBatch returns up to count DM status IDs strictly greater than minID
	// and less than or equal to maxIDInclusive. Note that this is different from most of our paging,
	// which uses a maxID and returns IDs strictly less than that, because it's called with the result of
	// MaxDirectStatusID, and expects to eventually return the status with that ID.
	// It is used only by the conversation advanced migration.
	GetDirectStatusIDsBatch(ctx context.Context, minID string, maxIDInclusive string, count int) ([]string, error)

	// GetStatusInteractions gets all abstract "interactions" of a status (likes, replies, boosts).
	// If localOnly is true, will return only interactions performed by accounts on this instance.
	// Aside from that, interactions are not filtered or deduplicated, it's up to the caller to do that.
	GetStatusInteractions(ctx context.Context, statusID string, localOnly bool) ([]gtsmodel.Interaction, error)

	// GetStatusByEditID gets one status corresponding to the given edit ID.
	GetStatusByEditID(ctx context.Context, editID string) (*gtsmodel.Status, error)
}
