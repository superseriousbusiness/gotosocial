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

	"database/sql"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/uptrace/bun"
)

// processErrorResponse parses the given error and returns an appropriate DBError.
func processErrorResponse(err error) db.Error {
	switch err {
	case nil:
		return nil
	case sql.ErrNoRows:
		return db.ErrNoEntries
	default:
		return err
	}
}

func exists(ctx context.Context, q *bun.SelectQuery) (bool, db.Error) {
	count, err := q.Count(ctx)

	exists := count != 0

	err = processErrorResponse(err)

	if err != nil {
		if err == db.ErrNoEntries {
			return false, nil
		}
		return false, err
	}

	return exists, nil
}

func notExists(ctx context.Context, q *bun.SelectQuery) (bool, db.Error) {
	count, err := q.Count(ctx)

	notExists := count == 0

	err = processErrorResponse(err)

	if err != nil {
		if err == db.ErrNoEntries {
			return true, nil
		}
		return false, err
	}

	return notExists, nil
}
