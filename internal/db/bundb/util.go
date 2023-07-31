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
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/uptrace/bun"
)

// likeEscaper is a thread-safe string replacer which escapes
// common SQLite + Postgres `LIKE` wildcard chars using the
// escape character `\`. Initialized as a var in this package
// so it can be reused.
var likeEscaper = strings.NewReplacer(
	`\`, `\\`, // Escape char.
	`%`, `\%`, // Zero or more char.
	`_`, `\_`, // Exactly one char.
)

// whereLike appends a WHERE clause to the
// given SelectQuery, which searches for
// matches of `search` in the given subQuery
// using LIKE.
func whereLike(
	query *bun.SelectQuery,
	subject interface{},
	search string,
) *bun.SelectQuery {
	// Escape existing wildcard + escape
	// chars in the search query string.
	search = likeEscaper.Replace(search)

	// Add our own wildcards back in; search
	// zero or more chars around the query.
	search = `%` + search + `%`

	// Append resulting WHERE
	// clause to the main query.
	return query.Where(
		"(?) LIKE ? ESCAPE ?",
		subject, search, `\`,
	)
}

// whereStartsLike is like whereLike,
// but only searches for strings that
// START WITH `search`.
func whereStartsLike(
	query *bun.SelectQuery,
	subject interface{},
	search string,
) *bun.SelectQuery {
	// Escape existing wildcard + escape
	// chars in the search query string.
	search = likeEscaper.Replace(search)

	// Add our own wildcards back in; search
	// zero or more chars after the query.
	search += `%`

	// Append resulting WHERE
	// clause to the main query.
	return query.Where(
		"(?) LIKE ? ESCAPE ?",
		subject, search, `\`,
	)
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
