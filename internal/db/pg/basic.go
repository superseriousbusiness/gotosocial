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
	"errors"
	"fmt"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/uptrace/bun"
)

type basicDB struct {
	config *config.Config
	conn   *bun.DB
	log    *logrus.Logger
	cancel context.CancelFunc
}

func (b *basicDB) Put(i interface{}) db.Error {
	_, err := b.conn.Model(i).Insert(i)
	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return db.ErrAlreadyExists
	}
	return err
}

func (b *basicDB) GetByID(id string, i interface{}) db.Error {
	if err := b.conn.Model(i).Where("id = ?", id).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries
		}
		return err

	}
	return nil
}

func (b *basicDB) GetWhere(where []db.Where, i interface{}) db.Error {
	if len(where) == 0 {
		return errors.New("no queries provided")
	}

	q := b.conn.Model(i)
	for _, w := range where {

		if w.Value == nil {
			q = q.Where("? IS NULL", pg.Ident(w.Key))
		} else {
			if w.CaseInsensitive {
				q = q.Where("LOWER(?) = LOWER(?)", pg.Safe(w.Key), w.Value)
			} else {
				q = q.Where("? = ?", pg.Safe(w.Key), w.Value)
			}
		}
	}

	if err := q.Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries
		}
		return err
	}
	return nil
}

func (b *basicDB) GetAll(i interface{}) db.Error {
	if err := b.conn.Model(i).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries
		}
		return err
	}
	return nil
}

func (b *basicDB) DeleteByID(id string, i interface{}) db.Error {
	if _, err := b.conn.Model(i).Where("id = ?", id).Delete(); err != nil {
		// if there are no rows *anyway* then that's fine
		// just return err if there's an actual error
		if err != pg.ErrNoRows {
			return err
		}
	}
	return nil
}

func (b *basicDB) DeleteWhere(where []db.Where, i interface{}) db.Error {
	if len(where) == 0 {
		return errors.New("no queries provided")
	}

	q := b.conn.Model(i)
	for _, w := range where {
		q = q.Where("? = ?", pg.Safe(w.Key), w.Value)
	}

	if _, err := q.Delete(); err != nil {
		// if there are no rows *anyway* then that's fine
		// just return err if there's an actual error
		if err != pg.ErrNoRows {
			return err
		}
	}
	return nil
}

func (b *basicDB) Upsert(i interface{}, conflictColumn string) db.Error {
	if _, err := b.conn.Model(i).OnConflict(fmt.Sprintf("(%s) DO UPDATE", conflictColumn)).Insert(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries
		}
		return err
	}
	return nil
}

func (b *basicDB) UpdateByID(id string, i interface{}) db.Error {
	if _, err := b.conn.Model(i).Where("id = ?", id).OnConflict("(id) DO UPDATE").Insert(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries
		}
		return err
	}
	return nil
}

func (b *basicDB) UpdateOneByID(id string, key string, value interface{}, i interface{}) db.Error {
	_, err := b.conn.Model(i).Set("? = ?", pg.Safe(key), value).Where("id = ?", id).Update()
	return err
}

func (b *basicDB) UpdateWhere(where []db.Where, key string, value interface{}, i interface{}) db.Error {
	q := b.conn.Model(i)

	for _, w := range where {
		if w.Value == nil {
			q = q.Where("? IS NULL", pg.Ident(w.Key))
		} else {
			if w.CaseInsensitive {
				q = q.Where("LOWER(?) = LOWER(?)", pg.Safe(w.Key), w.Value)
			} else {
				q = q.Where("? = ?", pg.Safe(w.Key), w.Value)
			}
		}
	}

	q = q.Set("? = ?", pg.Safe(key), value)

	_, err := q.Update()

	return err
}

func (b *basicDB) CreateTable(i interface{}) db.Error {
	return b.conn.Model(i).CreateTable(&orm.CreateTableOptions{
		IfNotExists: true,
	})
}

func (b *basicDB) DropTable(i interface{}) db.Error {
	return b.conn.Model(i).DropTable(&orm.DropTableOptions{
		IfExists: true,
	})
}

func (b *basicDB) RegisterTable(i interface{}) db.Error {
	orm.RegisterTable(i)
	return nil
}

func (b *basicDB) IsHealthy(ctx context.Context) db.Error {
	return b.conn.Ping(ctx)
}

func (b *basicDB) Stop(ctx context.Context) db.Error {
	b.log.Info("closing db connection")
	if err := b.conn.Close(); err != nil {
		// only cancel if there's a problem closing the db
		b.cancel()
		return err
	}
	return nil
}
