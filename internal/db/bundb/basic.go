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

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/uptrace/bun"
)

type basicDB struct {
	db *DB
}

func (b *basicDB) Put(ctx context.Context, i interface{}) error {
	_, err := b.db.NewInsert().Model(i).Exec(ctx)
	return err
}

func (b *basicDB) GetByID(ctx context.Context, id string, i interface{}) error {
	q := b.db.
		NewSelect().
		Model(i).
		Where("id = ?", id)

	err := q.Scan(ctx)
	return err
}

func (b *basicDB) GetWhere(ctx context.Context, where []db.Where, i interface{}) error {
	if len(where) == 0 {
		return errors.New("no queries provided")
	}

	q := b.db.NewSelect().Model(i)

	selectWhere(q, where)

	err := q.Scan(ctx)
	return err
}

func (b *basicDB) GetAll(ctx context.Context, i interface{}) error {
	q := b.db.
		NewSelect().
		Model(i)

	err := q.Scan(ctx)
	return err
}

func (b *basicDB) DeleteByID(ctx context.Context, id string, i interface{}) error {
	q := b.db.
		NewDelete().
		Model(i).
		Where("id = ?", id)

	_, err := q.Exec(ctx)
	return err
}

func (b *basicDB) DeleteWhere(ctx context.Context, where []db.Where, i interface{}) error {
	if len(where) == 0 {
		return errors.New("no queries provided")
	}

	q := b.db.
		NewDelete().
		Model(i)

	deleteWhere(q, where)

	_, err := q.Exec(ctx)
	return err
}

func (b *basicDB) UpdateByID(ctx context.Context, i interface{}, id string, columns ...string) error {
	q := b.db.
		NewUpdate().
		Model(i).
		Column(columns...).
		Where("? = ?", bun.Ident("id"), id)

	_, err := q.Exec(ctx)
	return err
}

func (b *basicDB) UpdateWhere(ctx context.Context, where []db.Where, key string, value interface{}, i interface{}) error {
	q := b.db.NewUpdate().Model(i)

	updateWhere(q, where)

	q = q.Set("? = ?", bun.Ident(key), value)

	_, err := q.Exec(ctx)
	return err
}

func (b *basicDB) CreateTable(ctx context.Context, i interface{}) error {
	_, err := b.db.NewCreateTable().Model(i).IfNotExists().Exec(ctx)
	return err
}

func (b *basicDB) CreateAllTables(ctx context.Context) error {
	models := []interface{}{
		&gtsmodel.Account{},
		&gtsmodel.Application{},
		&gtsmodel.Block{},
		&gtsmodel.DomainBlock{},
		&gtsmodel.EmailDomainBlock{},
		&gtsmodel.Follow{},
		&gtsmodel.FollowRequest{},
		&gtsmodel.MediaAttachment{},
		&gtsmodel.Mention{},
		&gtsmodel.Status{},
		&gtsmodel.StatusToEmoji{},
		&gtsmodel.StatusFave{},
		&gtsmodel.StatusBookmark{},
		&gtsmodel.StatusMute{},
		&gtsmodel.Tag{},
		&gtsmodel.User{},
		&gtsmodel.Emoji{},
		&gtsmodel.Instance{},
		&gtsmodel.Notification{},
		&gtsmodel.RouterSession{},
		&gtsmodel.Token{},
		&gtsmodel.Client{},
	}
	for _, i := range models {
		if err := b.CreateTable(ctx, i); err != nil {
			return err
		}
	}
	return nil
}

func (b *basicDB) DropTable(ctx context.Context, i interface{}) error {
	_, err := b.db.NewDropTable().Model(i).IfExists().Exec(ctx)
	return err
}

func (b *basicDB) IsHealthy(ctx context.Context) error {
	return b.db.PingContext(ctx)
}

func (b *basicDB) Close() error {
	log.Info(nil, "closing db connection")
	return b.db.Close()
}
