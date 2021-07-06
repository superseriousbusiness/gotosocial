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

func (ps *postgresService) DeleteByID(id string, i interface{}) error {
	if _, err := ps.conn.Model(i).Where("id = ?", id).Delete(); err != nil {
		// if there are no rows *anyway* then that's fine
		// just return err if there's an actual error
		if err != pg.ErrNoRows {
			return err
		}
	}
	return nil
}

func (ps *postgresService) DeleteWhere(where []db.Where, i interface{}) error {
	if len(where) == 0 {
		return errors.New("no queries provided")
	}

	q := ps.conn.Model(i)
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
