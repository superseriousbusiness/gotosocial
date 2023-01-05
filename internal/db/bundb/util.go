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
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/uptrace/bun"
)

// whereEmptyOrNull is a convenience function to return a bun WhereGroup that specifies
// that the given column should be EITHER an empty string OR null.
//
// Use it as follows:
//
//	q = q.WhereGroup(" AND ", whereEmptyOrNull("whatever_column"))
func whereEmptyOrNull(column string) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.
			WhereOr("? IS NULL", bun.Ident(column)).
			WhereOr("? = ''", bun.Ident(column))
	}
}

// whereNotEmptyAndNotNull is a convenience function to return a bun WhereGroup that specifies
// that the given column should be NEITHER an empty string NOR null.
//
// Use it as follows:
//
//	q = q.WhereGroup(" AND ", whereNotEmptyAndNotNull("whatever_column"))
func whereNotEmptyAndNotNull(column string) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.
			Where("? IS NOT NULL", bun.Ident(column)).
			Where("? != ''", bun.Ident(column))
	}
}

// updateWhere parses []db.Where and adds it to the given update query.
func updateWhere(q *bun.UpdateQuery, where []db.Where) {
	for _, w := range where {
		query, args := parseWhere(w)
		q = q.Where(query, args...)
	}
}

// selectWhere parses []db.Where and adds it to the given select query.
func selectWhere(q *bun.SelectQuery, where []db.Where) {
	for _, w := range where {
		query, args := parseWhere(w)
		q = q.Where(query, args...)
	}
}

// deleteWhere parses []db.Where and adds it to the given where query.
func deleteWhere(q *bun.DeleteQuery, where []db.Where) {
	for _, w := range where {
		query, args := parseWhere(w)
		q = q.Where(query, args...)
	}
}

// parseWhere looks through the options on a single db.Where entry, and
// returns the appropriate query string and arguments.
func parseWhere(w db.Where) (query string, args []interface{}) {
	if w.Not {
		if w.Value == nil {
			query = "? IS NOT NULL"
			args = []interface{}{bun.Ident(w.Key)}
			return
		}

		query = "? != ?"
		args = []interface{}{bun.Ident(w.Key), w.Value}
		return
	}

	if w.Value == nil {
		query = "? IS NULL"
		args = []interface{}{bun.Ident(w.Key)}
		return
	}

	query = "? = ?"
	args = []interface{}{bun.Ident(w.Key), w.Value}
	return
}
