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

package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"codeberg.org/gruf/go-storage"
	"codeberg.org/gruf/go-storage/disk"
	"github.com/uptrace/bun"
)

func init() {
	deleteAttachment := func(ctx context.Context, l log.Entry, a *gtsmodel.MediaAttachment, s storage.Storage, tx bun.Tx) {
		if err := s.Remove(ctx, a.File.Path); err != nil && !errors.Is(err, storage.ErrNotFound) {
			l.Errorf("error removing file %s: %s", a.File.Path, err)
		} else {
			l.Debugf("deleted %s", a.File.Path)
		}

		if err := s.Remove(ctx, a.Thumbnail.Path); err != nil && !errors.Is(err, storage.ErrNotFound) {
			l.Errorf("error removing file %s: %s", a.Thumbnail.Path, err)
		} else {
			l.Debugf("deleted %s", a.Thumbnail.Path)
		}

		if _, err := tx.NewDelete().
			TableExpr("? AS ?", bun.Ident("media_attachments"), bun.Ident("media_attachment")).
			Where("? = ?", bun.Ident("media_attachment.id"), a.ID).
			Exec(ctx); err != nil {
			l.Errorf("error deleting attachment with id %s: %s", a.ID, err)
		} else {
			l.Debugf("deleted attachment with id %s", a.ID)
		}
	}

	up := func(ctx context.Context, db *bun.DB) error {
		l := log.WithField("migration", "20220612091800_duplicated_media_cleanup")

		if config.GetStorageBackend() != "local" {
			// this migration only affects versions which only supported local storage
			return nil
		}

		storageBasePath := config.GetStorageLocalBasePath()
		if storageBasePath == "" {
			return fmt.Errorf("%s must be set to do storage migration", config.StorageLocalBasePathFlag)
		}

		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			s, err := disk.Open(storageBasePath, nil)
			if err != nil {
				return fmt.Errorf("error creating storage backend: %s", err)
			}

			// step 1. select all media attachment remote URLs that have duplicates
			var dupes int
			dupedRemoteURLs := []*gtsmodel.MediaAttachment{}
			if err := tx.NewSelect().
				Model(&dupedRemoteURLs).
				ColumnExpr("remote_url", "count(*)").
				Where("remote_url IS NOT NULL").
				Group("remote_url").
				Having("count(*) > 1").
				Scan(ctx); err != nil {
				return err
			}
			dupes = len(dupedRemoteURLs)
			l.Infof("found %d attachments with duplicate remote URLs", dupes)

			for i, dupedRemoteURL := range dupedRemoteURLs {
				if i%10 == 0 {
					l.Infof("cleaning %d of %d", i, dupes)
				}

				// step 2: select all media attachments associated with this url
				dupedAttachments := []*gtsmodel.MediaAttachment{}
				if err := tx.NewSelect().
					Model(&dupedAttachments).
					Where("remote_url = ?", dupedRemoteURL.RemoteURL).
					Scan(ctx); err != nil {
					l.Errorf("error running same attachments query: %s", err)
					continue
				}
				l.Debugf("found %d duplicates of attachment with remote url %s", len(dupedAttachments), dupedRemoteURL.RemoteURL)

				var statusID string
			statusIDLoop:
				for _, dupe := range dupedAttachments {
					if dupe.StatusID != "" {
						statusID = dupe.StatusID
						break statusIDLoop
					}
				}

				if statusID == "" {
					l.Debugf("%s not associated with a status, moving on", dupedRemoteURL.RemoteURL)
					continue
				}
				l.Debugf("%s is associated with status %s", dupedRemoteURL.RemoteURL, statusID)

				// step 3: get the status that these attachments are supposedly associated with, bail if we can't get it
				status := &gtsmodel.Status{}
				if err := tx.NewSelect().
					Model(status).
					Where("id = ?", statusID).
					Scan(ctx); err != nil {
					if err != sql.ErrNoRows {
						l.Errorf("error selecting status with id %s: %s", statusID, err)
					}
					continue
				}

				// step 4: for each attachment, check if it's actually one that the status is currently set to use, and delete if not
				for _, dupe := range dupedAttachments {
					var currentlyUsed bool
				currentlyUsedLoop:
					for _, attachmentID := range status.AttachmentIDs {
						if attachmentID == dupe.ID {
							currentlyUsed = true
							break currentlyUsedLoop
						}
					}

					if currentlyUsed {
						l.Debugf("attachment with id %s is a correct current attachment, leaving it alone!", dupe.ID)
						continue
					}

					deleteAttachment(ctx, l, dupe, s, tx)
				}
			}
			return nil
		})
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return nil
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
