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

package pg

import (
	"context"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type mediaDB struct {
	config *config.Config
	conn   *pg.DB
	log    *logrus.Logger
	cancel context.CancelFunc
}

func (m *mediaDB) newMediaQ(i interface{}) *orm.Query {
	return m.conn.Model(i).
		Relation("Account")
}

func (m *mediaDB) GetAttachmentByID(id string) (*gtsmodel.MediaAttachment, db.Error) {
	attachment := &gtsmodel.MediaAttachment{}

	q := m.newMediaQ(attachment).
		Where("media_attachment.id = ?", id)

	err := processErrorResponse(q.Select())

	return attachment, err
}
