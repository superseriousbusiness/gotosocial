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

package bundb

import (
	"context"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type mediaDB struct {
	conn  *DBConn
	state *state.State
}

func (m *mediaDB) newMediaQ(i *gtsmodel.MediaAttachment) *bun.SelectQuery {
	return m.conn.
		NewSelect().
		Model(i)
}

func (m *mediaDB) GetAttachmentByID(ctx context.Context, id string) (*gtsmodel.MediaAttachment, db.Error) {
	return m.getAttachment(
		ctx,
		"ID",
		func(attachment *gtsmodel.MediaAttachment) error {
			return m.newMediaQ(attachment).Where("? = ?", bun.Ident("media_attachment.id"), id).Scan(ctx)
		},
		id,
	)
}

func (m *mediaDB) getAttachments(ctx context.Context, ids []string) ([]*gtsmodel.MediaAttachment, db.Error) {
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

	return m.getAttachments(ctx, attachmentIDs)
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

	return m.getAttachments(ctx, attachmentIDs)
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

	return m.getAttachments(ctx, attachmentIDs)
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

func (m *mediaDB) getAttachment(ctx context.Context, lookup string, dbQuery func(*gtsmodel.MediaAttachment) error, keyParts ...any) (*gtsmodel.MediaAttachment, db.Error) {
	// Fetch attachment from database
	// todo: cache this lookup
	attachment := new(gtsmodel.MediaAttachment)

	if err := dbQuery(attachment); err != nil {
		return nil, m.conn.ProcessError(err)
	}

	return attachment, nil
}
