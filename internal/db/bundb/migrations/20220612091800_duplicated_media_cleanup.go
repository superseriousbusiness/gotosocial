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

package migrations

import (
	"context"
	"fmt"
	"path"

	"codeberg.org/gruf/go-store/kv"
	"codeberg.org/gruf/go-store/storage"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

func init() {
	deleteAttachment := func(ctx context.Context, l *logrus.Entry, a *gtsmodel.MediaAttachment, s *kv.KVStore, tx bun.Tx) {
		if err := s.Delete(a.File.Path); err != nil && err != storage.ErrNotFound {
			l.Errorf("error removing file %s: %s", a.File.Path, err)
		}
		l.Debugf("deleted %s", a.File.Path)

		if err := s.Delete(a.Thumbnail.Path); err != nil && err != storage.ErrNotFound {
			l.Errorf("error removing file %s: %s", a.Thumbnail.Path, err)
		}
		l.Debugf("deleted %s", a.Thumbnail.Path)

		if _, err := tx.NewDelete().
			Model(a).
			WherePK().
			Exec(ctx); err != nil {
			l.Errorf("error deleting attachment with id %s: %s", a.ID, err)
		}
	}

	up := func(ctx context.Context, db *bun.DB) error {
		l := logrus.WithField("migration", "20220612091800_duplicated_media_cleanup")

		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			storageBasePath := config.GetStorageLocalBasePath()
			s, err := kv.OpenFile(storageBasePath, &storage.DiskConfig{
				LockFile: path.Join(storageBasePath, "store.lock"),
			})
			if err != nil {
				return fmt.Errorf("error creating storage backend: %s", err)
			}
			defer s.Close()

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

				statusID := dupedAttachments[0].StatusID
				if statusID == "" {
					l.Debugf("%s not associated with a status, moving on", dupedRemoteURL.RemoteURL)
					continue
				}

				// step 3: get the status that these attachments are supposedly associated with
				status := &gtsmodel.Status{}
				if err := tx.NewSelect().
					Model(status).
					Where("id = ?", statusID).
					Scan(ctx); err != nil {
					l.Errorf("error selecting status with id %s: %s", statusID, err)
					continue
				}

				// step 4: for each attachment, check if it's actually the one that the status is currently set to use, and delete if not
				for _, a := range dupedAttachments {
					var current bool
					for _, attachmentID := range status.AttachmentIDs {
						if attachmentID == a.ID {
							current = true
						}
					}

					if current {
						l.Debugf("attachment with id %s is the correct current attachment, leaving it alone!", a.ID)
						continue
					}

					deleteAttachment(ctx, l, a, s, tx)
				}
			}
			return nil
		})
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			return nil
		})
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
