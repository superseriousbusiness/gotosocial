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
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/uptrace/bun"
)

type basicDB struct {
	config *config.Config
	conn   *DBConn
}

func (b *basicDB) Put(ctx context.Context, i interface{}) db.Error {
	_, err := b.conn.NewInsert().Model(i).Exec(ctx)
	return b.conn.ProcessError(err)
}

func (b *basicDB) GetByID(ctx context.Context, id string, i interface{}) db.Error {
	q := b.conn.
		NewSelect().
		Model(i).
		Where("id = ?", id)

	err := q.Scan(ctx)
	return b.conn.ProcessError(err)
}

func (b *basicDB) GetWhere(ctx context.Context, where []db.Where, i interface{}) db.Error {
	if len(where) == 0 {
		return errors.New("no queries provided")
	}

	q := b.conn.NewSelect().Model(i)
	for _, w := range where {
		if w.Value == nil {
			q = q.Where("? IS NULL", bun.Ident(w.Key))
		} else {
			if w.CaseInsensitive {
				q = q.Where("LOWER(?) = LOWER(?)", bun.Safe(w.Key), w.Value)
			} else {
				q = q.Where("? = ?", bun.Safe(w.Key), w.Value)
			}
		}
	}

	err := q.Scan(ctx)
	return b.conn.ProcessError(err)
}

func (b *basicDB) GetAll(ctx context.Context, i interface{}) db.Error {
	q := b.conn.
		NewSelect().
		Model(i)

	err := q.Scan(ctx)
	return b.conn.ProcessError(err)
}

func (b *basicDB) DeleteByID(ctx context.Context, id string, i interface{}) db.Error {
	q := b.conn.
		NewDelete().
		Model(i).
		Where("id = ?", id)

	_, err := q.Exec(ctx)
	return b.conn.ProcessError(err)
}

func (b *basicDB) DeleteWhere(ctx context.Context, where []db.Where, i interface{}) db.Error {
	if len(where) == 0 {
		return errors.New("no queries provided")
	}

	q := b.conn.
		NewDelete().
		Model(i)

	for _, w := range where {
		q = q.Where("? = ?", bun.Safe(w.Key), w.Value)
	}

	_, err := q.Exec(ctx)
	return b.conn.ProcessError(err)
}

func (b *basicDB) UpdateByID(ctx context.Context, id string, i interface{}) db.Error {
	q := b.conn.
		NewUpdate().
		Model(i).
		WherePK()

	_, err := q.Exec(ctx)
	return b.conn.ProcessError(err)
}

func (b *basicDB) UpdateOneByID(ctx context.Context, id string, key string, value interface{}, i interface{}) db.Error {
	q := b.conn.NewUpdate().
		Model(i).
		Set("? = ?", bun.Safe(key), value).
		WherePK()

	_, err := q.Exec(ctx)
	return b.conn.ProcessError(err)
}

func (b *basicDB) UpdateWhere(ctx context.Context, where []db.Where, key string, value interface{}, i interface{}) db.Error {
	q := b.conn.NewUpdate().Model(i)

	for _, w := range where {
		if w.Value == nil {
			q = q.Where("? IS NULL", bun.Ident(w.Key))
		} else {
			if w.CaseInsensitive {
				q = q.Where("LOWER(?) = LOWER(?)", bun.Safe(w.Key), w.Value)
			} else {
				q = q.Where("? = ?", bun.Safe(w.Key), w.Value)
			}
		}
	}

	q = q.Set("? = ?", bun.Safe(key), value)

	_, err := q.Exec(ctx)
	return b.conn.ProcessError(err)
}

func (b *basicDB) CreateTable(ctx context.Context, i interface{}) db.Error {
	_, err := b.conn.NewCreateTable().Model(i).IfNotExists().Exec(ctx)
	return err
}

func (b *basicDB) DropTable(ctx context.Context, i interface{}) db.Error {
	_, err := b.conn.NewDropTable().Model(i).IfExists().Exec(ctx)
	return b.conn.ProcessError(err)
}

func (b *basicDB) IsHealthy(ctx context.Context) db.Error {
	return b.conn.Ping()
}

func (b *basicDB) Stop(ctx context.Context) db.Error {
	b.conn.log.Info("closing db connection")
	return b.conn.Close()
}
