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
	"github.com/uptrace/bun"
)

type mediaDB struct {
	conn *DBConn
}

func (m *mediaDB) newMediaQ(i interface{}) *bun.SelectQuery {
	return m.conn.
		NewSelect().
		Model(i).
		Relation("Account")
}

func (m *mediaDB) GetAttachmentByID(ctx context.Context, id string) (*gtsmodel.MediaAttachment, db.Error) {
	attachment := &gtsmodel.MediaAttachment{}

	q := m.newMediaQ(attachment).
		Where("? = ?", bun.Ident("media_attachment.id"), id)

	if err := q.Scan(ctx); err != nil {
		return nil, m.conn.ProcessError(err)
	}
	return attachment, nil
}

func (m *mediaDB) GetRemoteOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]*gtsmodel.MediaAttachment, db.Error) {
	attachments := []*gtsmodel.MediaAttachment{}

	q := m.conn.
		NewSelect().
		Model(&attachments).
		Where("? = ?", bun.Ident("media_attachment.cached"), true).
		Where("? < ?", bun.Ident("media_attachment.created_at"), olderThan).
		WhereGroup(" AND ", whereNotEmptyAndNotNull("media_attachment.remote_url")).
		Order("media_attachment.created_at DESC")

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, m.conn.ProcessError(err)
	}

	return attachments, nil
}

func (m *mediaDB) GetAvatarsAndHeaders(ctx context.Context, maxID string, limit int) ([]*gtsmodel.MediaAttachment, db.Error) {
	attachments := []*gtsmodel.MediaAttachment{}

	q := m.newMediaQ(&attachments).
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

	if err := q.Scan(ctx); err != nil {
		return nil, m.conn.ProcessError(err)
	}

	return attachments, nil
}

func (m *mediaDB) GetLocalUnattachedOlderThan(ctx context.Context, olderThan time.Time, maxID string, limit int) ([]*gtsmodel.MediaAttachment, db.Error) {
	attachments := []*gtsmodel.MediaAttachment{}

	q := m.newMediaQ(&attachments).
		Where("? = ?", bun.Ident("media_attachment.cached"), true).
		Where("? = ?", bun.Ident("media_attachment.avatar"), false).
		Where("? = ?", bun.Ident("media_attachment.header"), false).
		Where("? < ?", bun.Ident("media_attachment.created_at"), olderThan).
		Where("? IS NULL", bun.Ident("media_attachment.remote_url")).
		Where("? IS NULL", bun.Ident("media_attachment.status_id"))

	if maxID != "" {
		q = q.Where("? < ?", bun.Ident("media_attachment.id"), maxID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, m.conn.ProcessError(err)
	}

	return attachments, nil
}
