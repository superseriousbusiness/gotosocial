/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type mediaDB struct {
	config *config.Config
	conn   *DBConn
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
		Where("media_attachment.id = ?", id)

	err := q.Scan(ctx)
	if err != nil {
		return nil, m.conn.ProcessError(err)
	}
	return attachment, nil
}
