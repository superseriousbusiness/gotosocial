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
	"crypto/rand"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

type sessionDB struct {
	conn *DBConn
}

func (s *sessionDB) GetSession(ctx context.Context) (*gtsmodel.RouterSession, db.Error) {
	rss := make([]*gtsmodel.RouterSession, 0, 1)

	// get the first router session in the db or...
	if err := s.conn.
		NewSelect().
		Model(&rss).
		Limit(1).
		Order("router_session.id DESC").
		Scan(ctx); err != nil {
		return nil, s.conn.ProcessError(err)
	}

	// ... create a new one
	if len(rss) == 0 {
		return s.createSession(ctx)
	}

	return rss[0], nil
}

func (s *sessionDB) createSession(ctx context.Context) (*gtsmodel.RouterSession, db.Error) {
	auth := make([]byte, 32)
	crypt := make([]byte, 32)

	if _, err := rand.Read(auth); err != nil {
		return nil, err
	}
	if _, err := rand.Read(crypt); err != nil {
		return nil, err
	}

	rs := &gtsmodel.RouterSession{
		ID:    id.NewULID(),
		Auth:  auth,
		Crypt: crypt,
	}

	if _, err := s.conn.
		NewInsert().
		Model(rs).
		Exec(ctx); err != nil {
		return nil, s.conn.ProcessError(err)
	}

	return rs, nil
}
