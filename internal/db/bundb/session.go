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
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

type sessionDB struct {
	config *config.Config
	conn   *DBConn
}

func (s *sessionDB) GetSession(ctx context.Context) (*gtsmodel.RouterSession, db.Error) {
	rss := make([]*gtsmodel.RouterSession, 0, 1)

	_, err := s.conn.
		NewSelect().
		Model(&rss).
		Limit(1).
		Order("id DESC").
		Exec(ctx)
	if err != nil {
		return nil, s.conn.ProcessError(err)
	}

	if len(rss) <= 0 {
		// no session created yet, so make one
		return s.createSession(ctx)
	}

	if len(rss) != 1 {
		// we asked for 1 so we should get 1
		return nil, errors.New("more than 1 router session was returned")
	}

	// return the one session found
	rs := rss[0]
	return rs, nil
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
	if err != nil {
		return nil, s.conn.ProcessError(err)
	}
	return rs, nil
}
