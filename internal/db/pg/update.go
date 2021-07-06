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
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

func (ps *postgresService) Upsert(i interface{}, conflictColumn string) error {
	if _, err := ps.conn.Model(i).OnConflict(fmt.Sprintf("(%s) DO UPDATE", conflictColumn)).Insert(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) UpdateByID(id string, i interface{}) error {
	if _, err := ps.conn.Model(i).Where("id = ?", id).OnConflict("(id) DO UPDATE").Insert(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) UpdateOneByID(id string, key string, value interface{}, i interface{}) error {
	_, err := ps.conn.Model(i).Set("? = ?", pg.Safe(key), value).Where("id = ?", id).Update()
	return err
}

func (ps *postgresService) UpdateWhere(where []db.Where, key string, value interface{}, i interface{}) error {
	q := ps.conn.Model(i)

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
