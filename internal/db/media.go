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

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Media contains functions related to creating/getting/removing media attachments.
type Media interface {
	// GetAttachmentByID gets a single attachment by its ID.
	GetAttachmentByID(ctx context.Context, id string) (*gtsmodel.MediaAttachment, error)

	// GetAttachmentsByIDs fetches a list of media attachments for given IDs.
	GetAttachmentsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.MediaAttachment, error)

	// PutAttachment inserts the given attachment into the database.
	PutAttachment(ctx context.Context, media *gtsmodel.MediaAttachment) error

	// UpdateAttachment will update the given attachment in the database.
	UpdateAttachment(ctx context.Context, media *gtsmodel.MediaAttachment, columns ...string) error

	// DeleteAttachment deletes the attachment with given ID from the database.
	DeleteAttachment(ctx context.Context, id string) error

	// GetAttachments ...
	GetAttachments(ctx context.Context, maxID string, limit int) ([]*gtsmodel.MediaAttachment, error)

	// GetRemoteAttachments ...
	GetRemoteAttachments(ctx context.Context, maxID string, limit int) ([]*gtsmodel.MediaAttachment, error)

	// GetCachedAttachmentsOlderThan gets limit n remote attachments (including avatars and headers) older than
	// the given time. These will be returned in order of attachment.created_at descending (i.e. newest to oldest).
	GetCachedAttachmentsOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]*gtsmodel.MediaAttachment, error)

	// CountRemoteOlderThan is like GetRemoteOlderThan, except instead of getting limit n attachments,
	// it just counts how many remote attachments in the database (including avatars and headers) meet
	// the olderThan criteria.
	CountRemoteOlderThan(ctx context.Context, olderThan time.Time) (int, error)

	// GetAvatarsAndHeaders fetches limit n avatars and headers with an id < maxID. These headers
	// and avis may be in use or not; the caller should check this if it's important.
	GetAvatarsAndHeaders(ctx context.Context, maxID string, limit int) ([]*gtsmodel.MediaAttachment, error)

	// GetLocalUnattachedOlderThan fetches limit n local media attachments (including avatars and headers), older than
	// the given time, which aren't header or avatars, and aren't attached to a status. In other words, attachments which were
	// uploaded but never used for whatever reason, or attachments that were attached to a status which was subsequently deleted.
	//
	// These will be returned in order of attachment.created_at descending (newest to oldest in other words).
	GetLocalUnattachedOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]*gtsmodel.MediaAttachment, error)

	// CountLocalUnattachedOlderThan is like GetLocalUnattachedOlderThan, except instead of getting limit n attachments,
	// it just counts how many local attachments in the database meet the olderThan criteria.
	CountLocalUnattachedOlderThan(ctx context.Context, olderThan time.Time) (int, error)
}
