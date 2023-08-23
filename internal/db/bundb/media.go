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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type mediaDB struct {
	db    *DB
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
	attachments := make([]*gtsmodel.MediaAttachment, 0, len(ids))

	for _, id := range ids {
		// Attempt fetch from DB
		attachment, err := m.GetAttachmentByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting attachment %q: %v", id, err)
			continue
		}

		// Append attachment
		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

func (m *mediaDB) getAttachment(ctx context.Context, lookup string, dbQuery func(*gtsmodel.MediaAttachment) error, keyParts ...any) (*gtsmodel.MediaAttachment, error) {
	return m.state.Caches.GTS.Media().Load(lookup, func() (*gtsmodel.MediaAttachment, error) {
		var attachment gtsmodel.MediaAttachment

		// Not cached! Perform database query
		if err := dbQuery(&attachment); err != nil {
			return nil, err
		}

		return &attachment, nil
	}, keyParts...)
}

func (m *mediaDB) PutAttachment(ctx context.Context, media *gtsmodel.MediaAttachment) error {
	return m.state.Caches.GTS.Media().Store(media, func() error {
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

	return m.state.Caches.GTS.Media().Store(media, func() error {
		_, err := m.db.NewUpdate().
			Model(media).
			Where("? = ?", bun.Ident("media_attachment.id"), media.ID).
			Column(columns...).
			Exec(ctx)
		return err
	})
}

func (m *mediaDB) DeleteAttachment(ctx context.Context, id string) error {
	// Load media into cache before attempting a delete,
	// as we need it cached in order to trigger the invalidate
	// callback. This in turn invalidates others.
	media, err := m.GetAttachmentByID(gtscontext.SetBarebones(ctx), id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// not an issue.
			err = nil
		}
		return err
	}

	// On return, ensure that media with ID is invalidated.
	defer m.state.Caches.GTS.Media().Invalidate("ID", id)

	// Delete media attachment in new transaction.
	err = m.db.RunInTx(ctx, func(tx bun.Tx) error {
		if media.AccountID != "" {
			var account gtsmodel.Account

			// Get related account model.
			if _, err := tx.NewSelect().
				Model(&account).
				Where("? = ?", bun.Ident("id"), media.AccountID).
				Exec(ctx); err != nil && !errors.Is(err, db.ErrNoEntries) {
				return gtserror.Newf("error selecting account: %w", err)
			}

			var set func(*bun.UpdateQuery) *bun.UpdateQuery

			switch {
			case *media.Avatar && account.AvatarMediaAttachmentID == id:
				set = func(q *bun.UpdateQuery) *bun.UpdateQuery {
					return q.Set("? = NULL", bun.Ident("avatar_media_attachment_id"))
				}
			case *media.Header && account.HeaderMediaAttachmentID == id:
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

		if media.StatusID != "" {
			var status gtsmodel.Status

			// Get related status model.
			if _, err := tx.NewSelect().
				Model(&status).
				Where("? = ?", bun.Ident("id"), media.StatusID).
				Exec(ctx); err != nil && !errors.Is(err, db.ErrNoEntries) {
				return gtserror.Newf("error selecting status: %w", err)
			}

			if updatedIDs := dropID(status.AttachmentIDs, id); // nocollapse
			len(updatedIDs) != len(status.AttachmentIDs) {
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

		// Finally delete this media.
		if _, err := tx.NewDelete().
			Table("media_attachments").
			Where("? = ?", bun.Ident("id"), id).
			Exec(ctx); err != nil {
			return gtserror.Newf("error deleting media: %w", err)
		}

		return nil
	})

	return err
}

func (m *mediaDB) GetAttachments(ctx context.Context, maxID string, limit int) ([]*gtsmodel.MediaAttachment, error) {
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

func (m *mediaDB) GetRemoteAttachments(ctx context.Context, maxID string, limit int) ([]*gtsmodel.MediaAttachment, error) {
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
