/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package media

import (
	"context"
	"errors"

	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (m *manager) PruneAllMeta(ctx context.Context) (int, error) {
	var (
		totalPruned int
		maxID       string
	)

	for {
		// select "selectPruneLimit" headers / avatars at a time for pruning
		attachments, err := m.db.GetAvatarsAndHeaders(ctx, maxID, selectPruneLimit)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return totalPruned, err
		} else if len(attachments) == 0 {
			break
		}

		// use the id of the last attachment in the slice as the next 'maxID' value
		log.Tracef("PruneAllMeta: got %d attachments with maxID < %s", len(attachments), maxID)
		maxID = attachments[len(attachments)-1].ID

		// prune each attachment that meets one of the following criteria:
		// - has no owning account in the database
		// - is a header but isn't the owning account's current header
		// - is an avatar but isn't the owning account's current avatar
		for _, attachment := range attachments {
			if attachment.Account == nil ||
				(*attachment.Header && attachment.ID != attachment.Account.HeaderMediaAttachmentID) ||
				(*attachment.Avatar && attachment.ID != attachment.Account.AvatarMediaAttachmentID) {
				if err := m.pruneOneAvatarOrHeader(ctx, attachment); err != nil {
					return totalPruned, err
				}
				totalPruned++
			}
		}
	}

	log.Infof("PruneAllMeta: finished pruning avatars + headers: pruned %d entries", totalPruned)
	return totalPruned, nil
}

func (m *manager) pruneOneAvatarOrHeader(ctx context.Context, attachment *gtsmodel.MediaAttachment) error {
	if attachment.File.Path != "" {
		// delete the full size attachment from storage
		log.Tracef("pruneOneAvatarOrHeader: deleting %s", attachment.File.Path)
		if err := m.storage.Delete(ctx, attachment.File.Path); err != nil && err != storage.ErrNotFound {
			return err
		}
	}

	if attachment.Thumbnail.Path != "" {
		// delete the thumbnail from storage
		log.Tracef("pruneOneAvatarOrHeader: deleting %s", attachment.Thumbnail.Path)
		if err := m.storage.Delete(ctx, attachment.Thumbnail.Path); err != nil && err != storage.ErrNotFound {
			return err
		}
	}

	// delete the attachment entry completely
	return m.db.DeleteByID(ctx, attachment.ID, &gtsmodel.MediaAttachment{})
}
