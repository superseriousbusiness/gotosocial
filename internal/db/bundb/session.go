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
	"crypto/rand"
	"io"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"github.com/uptrace/bun"
)

type sessionDB struct {
	db *bun.DB
}

func (s *sessionDB) GetSession(ctx context.Context) (*gtsmodel.RouterSession, error) {
	rss := make([]*gtsmodel.RouterSession, 0, 1)

	// get the first router session in the db or...
	if err := s.db.
		NewSelect().
		Model(&rss).
		Limit(1).
		Order("router_session.id DESC").
		Scan(ctx); err != nil {
		return nil, err
	}

	// ... create a new one
	if len(rss) == 0 {
		return s.createSession(ctx)
	}

	return rss[0], nil
}

func (s *sessionDB) createSession(ctx context.Context) (*gtsmodel.RouterSession, error) {
	buf := make([]byte, 64)
	auth := buf[:32]
	crypt := buf[32:64]

	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return nil, err
	}

	rs := &gtsmodel.RouterSession{
		ID:    id.NewULID(),
		Auth:  auth,
		Crypt: crypt,
	}

	if _, err := s.db.
		NewInsert().
		Model(rs).
		Exec(ctx); err != nil {
		return nil, err
	}

	return rs, nil
}
