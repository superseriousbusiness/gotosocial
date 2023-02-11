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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Media contains functions related to creating/getting/removing media attachments.
type Media interface {
	// GetAttachmentByID gets a single attachment by its ID
	GetAttachmentByID(ctx context.Context, id string) (*gtsmodel.MediaAttachment, Error)

	// GetRemoteOlderThan gets limit n remote media attachments (including avatars and headers) older than the given
	// olderThan time. These will be returned in order of attachment.created_at descending (newest to oldest in other words).
	//
	// The selected media attachments will be those with both a URL and a RemoteURL filled in.
	// In other words, media attachments that originated remotely, and that we currently have cached locally.
	GetRemoteOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]*gtsmodel.MediaAttachment, Error)

	// CountRemoteOlderThan is like GetRemoteOlderThan, except instead of getting limit n attachments,
	// it just counts how many remote attachments in the database (including avatars and headers) meet
	// the olderThan criteria.
	CountRemoteOlderThan(ctx context.Context, olderThan time.Time) (int, Error)

	// GetAvatarsAndHeaders fetches limit n avatars and headers with an id < maxID. These headers
	// and avis may be in use or not; the caller should check this if it's important.
	GetAvatarsAndHeaders(ctx context.Context, maxID string, limit int) ([]*gtsmodel.MediaAttachment, Error)

	// GetLocalUnattachedOlderThan fetches limit n local media attachments (including avatars and headers), older than
	// the given time, which aren't header or avatars, and aren't attached to a status. In other words, attachments which were
	// uploaded but never used for whatever reason, or attachments that were attached to a status which was subsequently deleted.
	//
	// These will be returned in order of attachment.created_at descending (newest to oldest in other words).
	GetLocalUnattachedOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]*gtsmodel.MediaAttachment, Error)

	// CountLocalUnattachedOlderThan is like GetLocalUnattachedOlderThan, except instead of getting limit n attachments,
	// it just counts how many local attachments in the database meet the olderThan criteria.
	CountLocalUnattachedOlderThan(ctx context.Context, olderThan time.Time) (int, Error)
}
