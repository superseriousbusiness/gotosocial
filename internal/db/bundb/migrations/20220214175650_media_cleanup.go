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
	"time"

	previousgtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20211113114307_init"
	newgtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20220214175650_media_cleanup"
	"github.com/uptrace/bun"
)

func init() {
	const batchSize = 100
	up := func(ctx context.Context, db *bun.DB) error {
		// we need to migrate media attachments into a new table
		// see section 6 here: https://www.sqlite.org/lang_altertable.html

		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// create the new media attachments table
			if _, err := tx.
				NewCreateTable().
				ModelTableExpr("new_media_attachments").
				Model(&newgtsmodel.MediaAttachment{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			offset := time.Now()
			// migrate existing media attachments into new table
		migrateLoop:
			for {
				oldAttachments := []*previousgtsmodel.MediaAttachment{}
				err := tx.
					NewSelect().
					Model(&oldAttachments).
					// subtract a millisecond from the offset just to make sure we're not getting double entries (this happens sometimes)
					Where("media_attachment.created_at < ?", offset.Add(-1*time.Millisecond)).
					Order("media_attachment.created_at DESC").
					Limit(batchSize).
					Scan(ctx)
				if err != nil && err != sql.ErrNoRows {
					// there's been a real error
					return err
				}

				if err == sql.ErrNoRows || len(oldAttachments) == 0 {
					// we're finished migrating
					break migrateLoop
				}

				// update the offset to the createdAt time of the oldest media attachment in the slice
				offset = oldAttachments[len(oldAttachments)-1].CreatedAt

				// for every old attachment, we need to make a new attachment out of it by taking the same values
				newAttachments := []*newgtsmodel.MediaAttachment{}
				for _, old := range oldAttachments {
					new := &newgtsmodel.MediaAttachment{
						ID:        old.ID,
						CreatedAt: old.CreatedAt,
						UpdatedAt: old.UpdatedAt,
						StatusID:  old.StatusID,
						URL:       old.URL,
						RemoteURL: old.RemoteURL,
						Type:      newgtsmodel.FileType(old.Type),
						FileMeta: newgtsmodel.FileMeta{
							Original: newgtsmodel.Original{
								Width:  old.FileMeta.Original.Width,
								Height: old.FileMeta.Original.Height,
								Size:   old.FileMeta.Original.Size,
								Aspect: old.FileMeta.Original.Aspect,
							},
							Small: newgtsmodel.Small{
								Width:  old.FileMeta.Small.Width,
								Height: old.FileMeta.Small.Height,
								Size:   old.FileMeta.Small.Size,
								Aspect: old.FileMeta.Small.Aspect,
							},
							Focus: newgtsmodel.Focus{
								X: old.FileMeta.Focus.X,
								Y: old.FileMeta.Focus.Y,
							},
						},
						AccountID:         old.AccountID,
						Description:       old.Description,
						ScheduledStatusID: old.ScheduledStatusID,
						Blurhash:          old.Blurhash,
						Processing:        newgtsmodel.ProcessingStatus(old.Processing),
						File: newgtsmodel.File{
							Path:        old.File.Path,
							ContentType: old.File.ContentType,
							FileSize:    old.File.FileSize,
							UpdatedAt:   old.File.UpdatedAt,
						},
						Thumbnail: newgtsmodel.Thumbnail{
							Path:        old.Thumbnail.Path,
							ContentType: old.Thumbnail.ContentType,
							FileSize:    old.Thumbnail.FileSize,
							UpdatedAt:   old.Thumbnail.UpdatedAt,
							URL:         old.Thumbnail.URL,
							RemoteURL:   old.Thumbnail.RemoteURL,
						},
						Avatar: old.Avatar,
						Header: old.Header,
						Cached: true,
					}
					newAttachments = append(newAttachments, new)
				}

				// insert this batch of new attachments, and then continue the loop
				if _, err := tx.
					NewInsert().
					Model(&newAttachments).
					ModelTableExpr("new_media_attachments").
					Exec(ctx); err != nil {
					return err
				}
			}

			// we have all the data we need from the old table, so we can safely drop it now
			if _, err := tx.NewDropTable().Model(&previousgtsmodel.MediaAttachment{}).Exec(ctx); err != nil {
				return err
			}

			// rename the new table to the same name as the old table was
			if _, err := tx.ExecContext(ctx, "ALTER TABLE new_media_attachments RENAME TO media_attachments;"); err != nil {
				return err
			}

			// add an index to the new table
			if _, err := tx.
				NewCreateIndex().
				Model(&newgtsmodel.MediaAttachment{}).
				Index("media_attachments_id_idx").
				Column("id").
				Exec(ctx); err != nil {
				return err
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
