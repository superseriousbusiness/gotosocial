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

package bundb

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

type mediaDB struct {
	db    *bun.DB
	state *state.State
}

func (m *mediaDB) GetAttachmentByID(ctx context.Context, id string) (*gtsmodel.MediaAttachment, error) {
	return m.getAttachment(
		ctx,
		"ID",
		func(attachment *gtsmodel.MediaAttachment) error {
			return m.db.NewSelect().
				Model(attachment).
				Where("? = ?", bun.Ident("media_attachment.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (m *mediaDB) GetAttachmentsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.MediaAttachment, error) {
	// Load all media IDs via cache loader callbacks.
	media, err := m.state.Caches.DB.Media.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.MediaAttachment, error) {
			// Preallocate expected length of uncached media attachments.
			media := make([]*gtsmodel.MediaAttachment, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := m.db.NewSelect().
				Model(&media).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return media, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the media by their
	// IDs to ensure in correct order.
	getID := func(m *gtsmodel.MediaAttachment) string { return m.ID }
	util.OrderBy(media, ids, getID)

	return media, nil
}

func (m *mediaDB) getAttachment(ctx context.Context, lookup string, dbQuery func(*gtsmodel.MediaAttachment) error, keyParts ...any) (*gtsmodel.MediaAttachment, error) {
	return m.state.Caches.DB.Media.LoadOne(lookup, func() (*gtsmodel.MediaAttachment, error) {
		var attachment gtsmodel.MediaAttachment

		// Not cached! Perform database query
		if err := dbQuery(&attachment); err != nil {
			return nil, err
		}

		return &attachment, nil
	}, keyParts...)
}

func (m *mediaDB) PutAttachment(ctx context.Context, media *gtsmodel.MediaAttachment) error {
	return m.state.Caches.DB.Media.Store(media, func() error {
		_, err := m.db.NewInsert().Model(media).Exec(ctx)
		return err
	})
}

func (m *mediaDB) UpdateAttachment(ctx context.Context, media *gtsmodel.MediaAttachment, columns ...string) error {
	media.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	return m.state.Caches.DB.Media.Store(media, func() error {
		_, err := m.db.NewUpdate().
			Model(media).
			Where("? = ?", bun.Ident("media_attachment.id"), media.ID).
			Column(columns...).
			Exec(ctx)
		return err
	})
}

func (m *mediaDB) DeleteAttachment(ctx context.Context, id string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted gtsmodel.MediaAttachment
	deleted.ID = id

	// Delete media attachment and update related models in new transaction.
	err := m.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

		// Initially, delete the media model,
		// returning the required fields we need.
		if _, err := tx.NewDelete().
			Model(&deleted).
			Where("? = ?", bun.Ident("id"), id).
			Returning("?, ?, ?, ?",
				bun.Ident("account_id"),
				bun.Ident("status_id"),
				bun.Ident("avatar"),
				bun.Ident("header"),
			).
			Exec(ctx); err != nil {
			return gtserror.Newf("error deleting media: %w", err)
		}

		// If media was attached to account,
		// we need to remove link from account.
		if deleted.AccountID != "" {
			var account gtsmodel.Account

			// Get related account model.
			if _, err := tx.NewSelect().
				Model(&account).
				Where("? = ?", bun.Ident("id"), deleted.AccountID).
				Exec(ctx); err != nil && !errors.Is(err, db.ErrNoEntries) {
				return gtserror.Newf("error selecting account: %w", err)
			}

			var set func(*bun.UpdateQuery) *bun.UpdateQuery

			switch {
			case *deleted.Avatar && account.AvatarMediaAttachmentID == id:
				set = func(q *bun.UpdateQuery) *bun.UpdateQuery {
					return q.Set("? = NULL", bun.Ident("avatar_media_attachment_id"))
				}
			case *deleted.Header && account.HeaderMediaAttachmentID == id:
				set = func(q *bun.UpdateQuery) *bun.UpdateQuery {
					return q.Set("? = NULL", bun.Ident("header_media_attachment_id"))
				}
			}

			if set != nil {
				// Note: this handles not found.
				//
				// Update the account model.
				q := tx.NewUpdate().
					Table("accounts").
					Where("? = ?", bun.Ident("id"), account.ID)
				if _, err := set(q).Exec(ctx); err != nil {
					return gtserror.Newf("error updating account: %w", err)
				}
			}
		}

		// If media was attached to a status,
		// we need to remove link from status.
		if deleted.StatusID != "" {
			var status gtsmodel.Status

			// Get related status model.
			if _, err := tx.NewSelect().
				Model(&status).
				Where("? = ?", bun.Ident("id"), deleted.StatusID).
				Exec(ctx); err != nil && !errors.Is(err, db.ErrNoEntries) {
				return gtserror.Newf("error selecting status: %w", err)
			}

			// Delete all instances of this deleted media ID from status attachments.
			updatedIDs := slices.DeleteFunc(status.AttachmentIDs, func(s string) bool {
				return s == id
			})

			if len(updatedIDs) != len(status.AttachmentIDs) {
				// Note: this handles not found.
				//
				// Attachments changed, update the status.
				if _, err := tx.NewUpdate().
					Table("statuses").
					Where("? = ?", bun.Ident("id"), status.ID).
					Set("? = ?", bun.Ident("attachment_ids"), updatedIDs).
					Exec(ctx); err != nil {
					return gtserror.Newf("error updating status: %w", err)
				}
			}
		}

		return nil
	})

	// Invalidate cached media with ID, manually
	// call invalidate hook in case not in cache.
	m.state.Caches.DB.Media.Invalidate("ID", id)
	m.state.Caches.OnInvalidateMedia(&deleted)

	return err
}

func (m *mediaDB) GetAttachments(ctx context.Context, page *paging.Page) ([]*gtsmodel.MediaAttachment, error) {
	maxID := page.GetMax()
	limit := page.GetLimit()

	attachmentIDs := make([]string, 0, limit)

	q := m.db.NewSelect().
		Table("media_attachments").
		Column("id").
		Order("id DESC")

	if maxID != "" {
		q = q.Where("id < ?", maxID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &attachmentIDs); err != nil {
		return nil, err
	}

	return m.GetAttachmentsByIDs(ctx, attachmentIDs)
}

func (m *mediaDB) GetRemoteAttachments(ctx context.Context, page *paging.Page) ([]*gtsmodel.MediaAttachment, error) {
	maxID := page.GetMax()
	limit := page.GetLimit()

	attachmentIDs := make([]string, 0, limit)

	q := m.db.NewSelect().
		Table("media_attachments").
		Column("id").
		Where("remote_url IS NOT NULL").
		Order("id DESC")

	if maxID != "" {
		q = q.Where("id < ?", maxID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &attachmentIDs); err != nil {
		return nil, err
	}

	return m.GetAttachmentsByIDs(ctx, attachmentIDs)
}

func (m *mediaDB) GetCachedAttachmentsOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]*gtsmodel.MediaAttachment, error) {
	attachmentIDs := make([]string, 0, limit)

	q := m.db.
		NewSelect().
		Table("media_attachments").
		Column("id").
		Where("cached = true").
		Where("remote_url IS NOT NULL").
		Where("created_at < ?", olderThan).
		Order("created_at DESC")

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &attachmentIDs); err != nil {
		return nil, err
	}

	return m.GetAttachmentsByIDs(ctx, attachmentIDs)
}
