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
	conn  *DBConn
	state *state.State
}

func (m *mediaDB) GetAttachmentByID(ctx context.Context, id string) (*gtsmodel.MediaAttachment, db.Error) {
	return m.getAttachment(
		ctx,
		"ID",
		func(attachment *gtsmodel.MediaAttachment) error {
			return m.conn.NewSelect().
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

func (m *mediaDB) getAttachment(ctx context.Context, lookup string, dbQuery func(*gtsmodel.MediaAttachment) error, keyParts ...any) (*gtsmodel.MediaAttachment, db.Error) {
	return m.state.Caches.GTS.Media().Load(lookup, func() (*gtsmodel.MediaAttachment, error) {
		var attachment gtsmodel.MediaAttachment

		// Not cached! Perform database query
		if err := dbQuery(&attachment); err != nil {
			return nil, m.conn.ProcessError(err)
		}

		return &attachment, nil
	}, keyParts...)
}

func (m *mediaDB) PutAttachment(ctx context.Context, media *gtsmodel.MediaAttachment) error {
	return m.state.Caches.GTS.Media().Store(media, func() error {
		_, err := m.conn.NewInsert().Model(media).Exec(ctx)
		return m.conn.ProcessError(err)
	})
}

func (m *mediaDB) UpdateAttachment(ctx context.Context, media *gtsmodel.MediaAttachment, columns ...string) error {
	media.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	return m.state.Caches.GTS.Media().Store(media, func() error {
		_, err := m.conn.NewUpdate().
			Model(media).
			Where("? = ?", bun.Ident("media_attachment.id"), media.ID).
			Column(columns...).
			Exec(ctx)
		return m.conn.ProcessError(err)
	})
}

func (m *mediaDB) DeleteAttachment(ctx context.Context, id string) error {
	defer m.state.Caches.GTS.Media().Invalidate("ID", id)

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

	var (
		invalidateAccount bool
		invalidateStatus  bool
	)

	// Delete media attachment in new transaction.
	err = m.conn.RunInTx(ctx, func(tx bun.Tx) error {
		// Attempt to delete this media.
		if _, err := m.conn.NewDelete().
			Table("media_attachments").
			Where("? = ?", bun.Ident("id"), id).
			Exec(ctx); err != nil {
			return gtserror.Newf("error deleting media: %w", err)
		}

		if media.AccountID != "" {
			var account gtsmodel.Account

			// Get related account model.
			if _, err := m.conn.NewSelect().
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
				q := m.conn.NewUpdate().
					Table("accounts").
					Where("? = ?", bun.Ident("id"), account.ID)
				if _, err := set(q).Exec(ctx); err != nil {
					return gtserror.Newf("error updating account: %w", err)
				}

				// Mark as needing invalidate.
				invalidateAccount = true
			}
		}

		if media.StatusID != "" {
			var status gtsmodel.Status

			// Get related status model.
			if _, err := m.conn.NewSelect().
				Model(&status).
				Where("? = ?", bun.Ident("id"), media.StatusID).
				Exec(ctx); err != nil && !errors.Is(err, db.ErrNoEntries) {
				return gtserror.Newf("error selecting status: %w", err)
			}

			// Get length of attachments beforehand.
			before := len(status.AttachmentIDs)

			for i := 0; i < len(status.AttachmentIDs); {
				if status.AttachmentIDs[i] == id {
					// Remove this reference to deleted attachment ID.
					copy(status.AttachmentIDs[i:], status.AttachmentIDs[i+1:])
					status.AttachmentIDs = status.AttachmentIDs[:len(status.AttachmentIDs)-1]
					continue
				}
				i++
			}

			if before != len(status.AttachmentIDs) {
				// Note: this accounts for status not found.
				//
				// Attachments changed, update the status.
				if _, err := m.conn.NewUpdate().
					Table("statuses").
					Where("? = ?", bun.Ident("id"), status.ID).
					Set("? = ?", bun.Ident("attachment_ids"), status.AttachmentIDs).
					Exec(ctx); err != nil {
					return gtserror.Newf("error updating status: %w", err)
				}

				// Mark as needing invalidate.
				invalidateStatus = true
			}
		}

		return nil
	})

	if invalidateAccount {
		// The account for given ID will have been updated in transaction.
		m.state.Caches.GTS.Account().Invalidate("ID", media.AccountID)
	}

	if invalidateStatus {
		// The status for given ID will have been updated in transaction.
		m.state.Caches.GTS.Status().Invalidate("ID", media.StatusID)
	}

	return m.conn.ProcessError(err)
}

func (m *mediaDB) GetRemoteOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]*gtsmodel.MediaAttachment, db.Error) {
	attachmentIDs := []string{}

	q := m.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("media_attachments"), bun.Ident("media_attachment")).
		Column("media_attachment.id").
		Where("? = ?", bun.Ident("media_attachment.cached"), true).
		Where("? < ?", bun.Ident("media_attachment.created_at"), olderThan).
		WhereGroup(" AND ", whereNotEmptyAndNotNull("media_attachment.remote_url")).
		Order("media_attachment.created_at DESC")

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &attachmentIDs); err != nil {
		return nil, m.conn.ProcessError(err)
	}

	return m.GetAttachmentsByIDs(ctx, attachmentIDs)
}

func (m *mediaDB) CountRemoteOlderThan(ctx context.Context, olderThan time.Time) (int, db.Error) {
	q := m.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("media_attachments"), bun.Ident("media_attachment")).
		Column("media_attachment.id").
		Where("? = ?", bun.Ident("media_attachment.cached"), true).
		Where("? < ?", bun.Ident("media_attachment.created_at"), olderThan).
		WhereGroup(" AND ", whereNotEmptyAndNotNull("media_attachment.remote_url"))

	count, err := q.Count(ctx)
	if err != nil {
		return 0, m.conn.ProcessError(err)
	}

	return count, nil
}

func (m *mediaDB) GetAttachments(ctx context.Context, maxID string, limit int) ([]*gtsmodel.MediaAttachment, error) {
	attachmentIDs := []string{}

	q := m.conn.NewSelect().
		TableExpr("? AS ?", bun.Ident("media_attachments"), bun.Ident("media_attachment")).
		Column("media_attachment.id").
		Order("media_attachment.id DESC")

	if maxID != "" {
		q = q.Where("? < ?", bun.Ident("media_attachment.id"), maxID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &attachmentIDs); err != nil {
		return nil, m.conn.ProcessError(err)
	}

	return m.GetAttachmentsByIDs(ctx, attachmentIDs)
}

func (m *mediaDB) GetAvatarsAndHeaders(ctx context.Context, maxID string, limit int) ([]*gtsmodel.MediaAttachment, db.Error) {
	attachmentIDs := []string{}

	q := m.conn.NewSelect().
		TableExpr("? AS ?", bun.Ident("media_attachments"), bun.Ident("media_attachment")).
		Column("media_attachment.id").
		WhereGroup(" AND ", func(innerQ *bun.SelectQuery) *bun.SelectQuery {
			return innerQ.
				WhereOr("? = ?", bun.Ident("media_attachment.avatar"), true).
				WhereOr("? = ?", bun.Ident("media_attachment.header"), true)
		}).
		Order("media_attachment.id DESC")

	if maxID != "" {
		q = q.Where("? < ?", bun.Ident("media_attachment.id"), maxID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &attachmentIDs); err != nil {
		return nil, m.conn.ProcessError(err)
	}

	return m.GetAttachmentsByIDs(ctx, attachmentIDs)
}

func (m *mediaDB) GetLocalUnattachedOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]*gtsmodel.MediaAttachment, db.Error) {
	attachmentIDs := []string{}

	q := m.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("media_attachments"), bun.Ident("media_attachment")).
		Column("media_attachment.id").
		Where("? = ?", bun.Ident("media_attachment.cached"), true).
		Where("? = ?", bun.Ident("media_attachment.avatar"), false).
		Where("? = ?", bun.Ident("media_attachment.header"), false).
		Where("? < ?", bun.Ident("media_attachment.created_at"), olderThan).
		Where("? IS NULL", bun.Ident("media_attachment.remote_url")).
		Where("? IS NULL", bun.Ident("media_attachment.status_id")).
		Order("media_attachment.created_at DESC")

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &attachmentIDs); err != nil {
		return nil, m.conn.ProcessError(err)
	}

	return m.GetAttachmentsByIDs(ctx, attachmentIDs)
}

func (m *mediaDB) CountLocalUnattachedOlderThan(ctx context.Context, olderThan time.Time) (int, db.Error) {
	q := m.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("media_attachments"), bun.Ident("media_attachment")).
		Column("media_attachment.id").
		Where("? = ?", bun.Ident("media_attachment.cached"), true).
		Where("? = ?", bun.Ident("media_attachment.avatar"), false).
		Where("? = ?", bun.Ident("media_attachment.header"), false).
		Where("? < ?", bun.Ident("media_attachment.created_at"), olderThan).
		Where("? IS NULL", bun.Ident("media_attachment.remote_url")).
		Where("? IS NULL", bun.Ident("media_attachment.status_id"))

	count, err := q.Count(ctx)
	if err != nil {
		return 0, m.conn.ProcessError(err)
	}

	return count, nil
}
