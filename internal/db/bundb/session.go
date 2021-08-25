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
	"crypto/rand"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/uptrace/bun"
)

type sessionDB struct {
	config *config.Config
	conn   *bun.DB
	log    *logrus.Logger
}

func (s *sessionDB) GetSession(ctx context.Context) (*gtsmodel.RouterSession, db.Error) {
	rs := new(gtsmodel.RouterSession)

	q := s.conn.
		NewSelect().
		Model(rs).
		Limit(1)

	_, err := q.Exec(ctx)

	err = processErrorResponse(err)

	return rs, err
}

func (s *sessionDB) CreateSession(ctx context.Context) (*gtsmodel.RouterSession, db.Error) {
	auth := make([]byte, 32)
	crypt := make([]byte, 32)

	if _, err := rand.Read(auth); err != nil {
		return nil, err
	}
	if _, err := rand.Read(crypt); err != nil {
		return nil, err
	}

	rid, err := id.NewULID()
	if err != nil {
		return nil, err
	}

	rs := &gtsmodel.RouterSession{
		ID:    rid,
		Auth:  auth,
		Crypt: crypt,
	}

	q := s.conn.
		NewInsert().
		Model(rs)

	_, err = q.Exec(ctx)

	err = processErrorResponse(err)

	return rs, err
}
