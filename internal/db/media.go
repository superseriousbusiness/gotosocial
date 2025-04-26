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

	// GetAttachments fetches media attachments up to a given max ID, and at most limit.
	GetAttachments(ctx context.Context, page *paging.Page) ([]*gtsmodel.MediaAttachment, error)

	// GetRemoteAttachments fetches media attachments with a non-empty domain, up to a given max ID, and at most limit.
	GetRemoteAttachments(ctx context.Context, page *paging.Page) ([]*gtsmodel.MediaAttachment, error)

	// GetCachedAttachmentsOlderThan gets limit n remote attachments (including avatars and headers) older than
	// the given time. These will be returned in order of attachment.created_at descending (i.e. newest to oldest).
	GetCachedAttachmentsOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]*gtsmodel.MediaAttachment, error)
}
