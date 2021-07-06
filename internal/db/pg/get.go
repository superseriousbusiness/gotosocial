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
	"errors"

	"github.com/go-pg/pg/v10"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

func (ps *postgresService) GetByID(id string, i interface{}) error {
	if err := ps.conn.Model(i).Where("id = ?", id).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err

	}
	return nil
}

func (ps *postgresService) GetWhere(where []db.Where, i interface{}) error {
	if len(where) == 0 {
		return errors.New("no queries provided")
	}

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

	if err := q.Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetAll(i interface{}) error {
	if err := ps.conn.Model(i).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}
