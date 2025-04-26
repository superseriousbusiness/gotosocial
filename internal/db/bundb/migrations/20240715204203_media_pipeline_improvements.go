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

	old_gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20240715204203_media_pipeline_improvements"
	new_gtsmodel "code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"

	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		log.Info(
			ctx,
			"doing media pipeline improvements; "+
				"this may take a while if your database has lots of media attachments, don't interrupt it!",
		)
		if err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			if _, err := tx.NewAddColumn().
				Table("media_attachments").
				ColumnExpr("? INTEGER NOT NULL DEFAULT ?", bun.Ident("type_new"), 0).
				Exec(ctx); err != nil {
				return err
			}

			for old, new := range map[old_gtsmodel.FileType]new_gtsmodel.FileType{
				old_gtsmodel.FileTypeAudio:   new_gtsmodel.FileTypeAudio,
				old_gtsmodel.FileTypeImage:   new_gtsmodel.FileTypeImage,
				old_gtsmodel.FileTypeGifv:    new_gtsmodel.FileTypeImage,
				old_gtsmodel.FileTypeVideo:   new_gtsmodel.FileTypeVideo,
				old_gtsmodel.FileTypeUnknown: new_gtsmodel.FileTypeUnknown,
			} {
				if _, err := tx.NewUpdate().
					Table("media_attachments").
					Where("? = ?", bun.Ident("type"), old).
					Set("? = ?", bun.Ident("type_new"), new).
					Exec(ctx); err != nil {
					return err
				}
			}

			if _, err := tx.NewDropColumn().
				Table("media_attachments").
				ColumnExpr("?", bun.Ident("type")).
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.NewRaw(
				"ALTER TABLE ? RENAME COLUMN ? TO ?",
				bun.Ident("media_attachments"),
				bun.Ident("type_new"),
				bun.Ident("type"),
			).Exec(ctx); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}

		// Zero-out attachment data
		// for "unknown" non-locally
		// stored media attachments.
		if _, err := db.NewUpdate().
			Table("media_attachments").
			Where("? = ?", bun.Ident("type"), new_gtsmodel.FileTypeUnknown).
			Set("? = ?", bun.Ident("url"), "").
			Set("? = ?", bun.Ident("file_path"), "").
			Set("? = ?", bun.Ident("file_content_type"), "").
			Set("? = ?", bun.Ident("file_file_size"), 0).
			Set("? = ?", bun.Ident("thumbnail_path"), "").
			Set("? = ?", bun.Ident("thumbnail_content_type"), "").
			Set("? = ?", bun.Ident("thumbnail_file_size"), 0).
			Set("? = ?", bun.Ident("thumbnail_url"), "").
			Exec(ctx); err != nil {
			return err
		}

		// Zero-out emoji data for
		// non-locally stored emoji.
		if _, err := db.NewUpdate().
			Table("emojis").
			WhereOr("? = ?", bun.Ident("image_url"), "").
			WhereOr("? = ?", bun.Ident("image_path"), "").
			Set("? = ?", bun.Ident("image_path"), "").
			Set("? = ?", bun.Ident("image_url"), "").
			Set("? = ?", bun.Ident("image_file_size"), 0).
			Set("? = ?", bun.Ident("image_content_type"), "").
			Set("? = ?", bun.Ident("image_static_path"), "").
			Set("? = ?", bun.Ident("image_static_url"), "").
			Set("? = ?", bun.Ident("image_static_file_size"), 0).
			Set("? = ?", bun.Ident("image_static_content_type"), "").
			Exec(ctx); err != nil {
			return err
		}

		return nil
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
