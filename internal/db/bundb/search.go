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
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

// todo: currently we pass an 'offset' parameter into functions owned by this struct,
// which is ignored.
//
// The idea of 'offset' is to allow callers to page through results without supplying
// maxID or minID params; they simply use the offset as more or less a 'page number'.
// This works fine when you're dealing with something like Elasticsearch, but for
// SQLite or Postgres 'LIKE' queries it doesn't really, because for each higher offset
// you have to calculate the value of all the previous offsets as well *within the
// execution time of the query*. It's MUCH more efficient to page using maxID and
// minID for queries like this. For now, then, we just ignore the offset and hope that
// the caller will page using maxID and minID instead.
//
// In future, however, it would be good to support offset in a way that doesn't totally
// destroy database queries. One option would be to cache previous offsets when paging
// down (which is the most common use case).
//
// For example, say a caller makes a call with offset 0: we run the query as normal,
// and in a 10 minute cache or something, store the next maxID value as it would be for
// offset 1, for the supplied query, limit, following, etc. Then when they call for
// offset 1, instead of supplying 'offset' in the query and causing slowdown, we check
// the cache to see if we have the next maxID value stored for that query, and use that
// instead. If a caller out of the blue requests offset 4 or something, on an empty cache,
// we could run the previous 4 queries and store the offsets for those before making the
// 5th call for page 4.
//
// This isn't ideal, of course, but at least we could cover the most common use case of
// a caller paging down through results.
type searchDB struct {
	conn  *DBConn
	state *state.State
}

// replacer is a thread-safe string replacer which escapes
// common SQLite + Postgres `LIKE` wildcard chars using the
// escape character `\`. Initialized as a var in this package
// so it can be reused.
var replacer = strings.NewReplacer(
	`\`, `\\`, // Escape char.
	`%`, `\%`, // Zero or more char.
	`_`, `\_`, // Exactly one char.
)

// whereLikeSubquery appends a WHERE clause to the
// given SelectQuery q, which searches for matches
// of searchQuery in the given subQuery using LIKE.
func whereLikeSubquery(
	q *bun.SelectQuery,
	subQuery *bun.SelectQuery,
	searchQuery string,
) {
	// Escape existing wildcard + escape
	// chars in the search query string.
	searchQuery = replacer.Replace(searchQuery)

	// Add our own wildcards back in; search
	// zero or more chars around the query.
	searchQuery = `%` + searchQuery + `%`

	// Append resulting WHERE
	// clause to the main query.
	q = q.Where(
		"(?) LIKE ? ESCAPE ?",
		subQuery, searchQuery, `\`,
	)
}

// Query example (SQLite):
//
//	SELECT "account"."id" FROM "accounts" AS "account"
//	WHERE (("account"."domain" IS NULL) OR ("account"."domain" != "account"."username"))
//	AND ("account"."id" < 'ZZZZZZZZZZZZZZZZZZZZZZZZZZ')
//	AND ("account"."id" IN (SELECT "target_account_id" FROM "follows" WHERE ("account_id" = '016T5Q3SQKBT337DAKVSKNXXW1')))
//	AND ((SELECT LOWER("account"."username" || COALESCE("account"."display_name", '') || COALESCE("account"."note", '')) AS "account_text") LIKE '%turtle%' ESCAPE '\')
//	ORDER BY "account"."id" DESC LIMIT 10
func (s *searchDB) SearchForAccounts(
	ctx context.Context,
	accountID string,
	query string,
	maxID string,
	minID string,
	limit int,
	following bool,
	offset int,
) ([]*gtsmodel.Account, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	var (
		accountIDs  = make([]string, 0, limit)
		frontToBack = true
	)

	q := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		// Select only IDs from table.
		Column("account.id").
		// Try to ignore instance accounts. Account domain must
		// be either nil or, if set, not equal to the account's
		// username (which is commonly used to indicate it's an
		// instance service account).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				Where("? IS NULL", bun.Ident("account.domain")).
				WhereOr("? != ?", bun.Ident("account.domain"), bun.Ident("account.username"))
		})

	// Return only items with a LOWER id than maxID.
	if maxID == "" {
		maxID = id.Highest
	}
	q = q.Where("? < ?", bun.Ident("account.id"), maxID)

	if minID != "" {
		// Return only items with a HIGHER id than minID.
		q = q.Where("? > ?", bun.Ident("account.id"), minID)

		// page up
		frontToBack = false
	}

	if following {
		// Select only from accounts followed by accountID.
		q = q.Where(
			"? IN (?)",
			bun.Ident("account.id"),
			s.followedAccounts(accountID),
		)
	}

	// Search for matches of query
	// within the accountText subquery.
	whereLikeSubquery(q, s.accountText(following), query)

	if limit > 0 {
		// Limit amount of accounts returned.
		q = q.Limit(limit)
	}

	if frontToBack {
		// Page down.
		q = q.Order("account.id DESC")
	} else {
		// Page up.
		q = q.Order("account.id ASC")
	}

	if err := q.Scan(ctx, &accountIDs); err != nil {
		return nil, s.conn.ProcessError(err)
	}

	if len(accountIDs) == 0 {
		return nil, nil
	}

	// If we're paging up, we still want accounts
	// to be sorted by ID desc, so reverse ids slice.
	// https://zchee.github.io/golang-wiki/SliceTricks/#reversing
	if !frontToBack {
		for l, r := 0, len(accountIDs)-1; l < r; l, r = l+1, r-1 {
			accountIDs[l], accountIDs[r] = accountIDs[r], accountIDs[l]
		}
	}

	accounts := make([]*gtsmodel.Account, 0, len(accountIDs))
	for _, id := range accountIDs {
		// Fetch account from db for ID
		account, err := s.state.DB.GetAccountByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error fetching account %q: %v", id, err)
			continue
		}

		// Append account to slice
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// Query example (SQLite):
//
//	SELECT "status"."id"
//	FROM "statuses" AS "status"
//	WHERE (("status"."account_id" = '01F8MH1H7YV1Z7D2C8K2730QBF') OR ("status"."in_reply_to_account_id" = '01F8MH1H7YV1Z7D2C8K2730QBF'))
//	AND ("status"."boost_of_id" IS NULL)
//	AND ("status"."id" < 'ZZZZZZZZZZZZZZZZZZZZZZZZZZ')
//	AND ((SELECT LOWER("status"."content" || COALESCE("status"."content_warning", '')) AS "status_text") LIKE '%hello%' ESCAPE '\')
//	ORDER BY "status"."id" DESC LIMIT 10
func (s *searchDB) SearchForStatuses(
	ctx context.Context,
	accountID string,
	query string,
	maxID string,
	minID string,
	limit int,
	offset int,
) ([]*gtsmodel.Status, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	var (
		statusIDs   = make([]string, 0, limit)
		frontToBack = true
	)

	q := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		// Select only IDs from table
		Column("status.id").
		// Search only for statuses created by accountID,
		// or statuses posted as a reply to accountID.
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				Where("? = ?", bun.Ident("status.account_id"), accountID).
				WhereOr("? = ?", bun.Ident("status.in_reply_to_account_id"), accountID)
		}).
		// Ignore boosts.
		Where("? IS NULL", bun.Ident("status.boost_of_id"))

	// Return only items with a LOWER id than maxID.
	if maxID == "" {
		maxID = id.Highest
	}
	q = q.Where("? < ?", bun.Ident("status.id"), maxID)

	if minID != "" {
		// return only statuses HIGHER (ie., newer) than minID
		q = q.Where("? > ?", bun.Ident("status.id"), minID)

		// page up
		frontToBack = false
	}

	// Search for matches of query
	// within the statusText subquery.
	whereLikeSubquery(q, s.statusText(), query)

	if limit > 0 {
		// Limit amount of statuses returned.
		q = q.Limit(limit)
	}

	if frontToBack {
		// Page down.
		q = q.Order("status.id DESC")
	} else {
		// Page up.
		q = q.Order("status.id ASC")
	}

	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, s.conn.ProcessError(err)
	}

	if len(statusIDs) == 0 {
		return nil, nil
	}

	// If we're paging up, we still want statuses
	// to be sorted by ID desc, so reverse ids slice.
	// https://zchee.github.io/golang-wiki/SliceTricks/#reversing
	if !frontToBack {
		for l, r := 0, len(statusIDs)-1; l < r; l, r = l+1, r-1 {
			statusIDs[l], statusIDs[r] = statusIDs[r], statusIDs[l]
		}
	}

	statuses := make([]*gtsmodel.Status, 0, len(statusIDs))
	for _, id := range statusIDs {
		// Fetch status from db for ID
		status, err := s.state.DB.GetStatusByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error fetching status %q: %v", id, err)
			continue
		}

		// Append status to slice
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// followedAccounts returns a subquery that selects only IDs
// of accounts that are followed by the given accountID.
func (s *searchDB) followedAccounts(accountID string) *bun.SelectQuery {
	return s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("follows"), bun.Ident("follow")).
		Column("follow.target_account_id").
		Where("? = ?", bun.Ident("follow.account_id"), accountID)
}

func (s *searchDB) accountText(following bool) *bun.SelectQuery {
	var (
		accountText = s.conn.NewSelect()
		query       string
		args        []interface{}
	)

	if following {
		// If querying for accounts we follow,
		// include note in text search params.
		args = []interface{}{
			bun.Ident("account.username"),
			bun.Ident("account.display_name"), "",
			bun.Ident("account.note"), "",
			bun.Ident("account_text"),
		}
	} else {
		// If querying for accounts we're not following,
		// don't include note in text search params.
		args = []interface{}{
			bun.Ident("account.username"),
			bun.Ident("account.display_name"), "",
			bun.Ident("account_text"),
		}
	}

	// SQLite and Postgres use different syntaxes for
	// concatenation, and we also need to use a
	// different number of placeholders depending on
	// following/not following. COALESCE calls ensure
	// that we're not trying to concatenate null values.
	d := s.conn.Dialect().Name()
	switch {

	case d == dialect.SQLite && following:
		query = "LOWER(? || COALESCE(?, ?) || COALESCE(?, ?)) AS ?"

	case d == dialect.SQLite && !following:
		query = "LOWER(? || COALESCE(?, ?)) AS ?"

	case d == dialect.PG && following:
		query = "LOWER(CONCAT(?, COALESCE(?, ?), COALESCE(?, ?))) AS ?"

	case d == dialect.PG && !following:
		query = "LOWER(CONCAT(?, COALESCE(?, ?))) AS ?"

	default:
		panic("db conn was neither pg not sqlite")
	}

	return accountText.ColumnExpr(query, args...)
}

func (s *searchDB) statusText() *bun.SelectQuery {
	statusText := s.conn.NewSelect()

	// SQLite and Postgres use different
	// syntaxes for concatenation.
	switch s.conn.Dialect().Name() {

	case dialect.SQLite:
		statusText = statusText.ColumnExpr(
			"LOWER(? || COALESCE(?, ?)) AS ?",
			bun.Ident("status.content"), bun.Ident("status.content_warning"), "",
			bun.Ident("status_text"))

	case dialect.PG:
		statusText = statusText.ColumnExpr(
			"LOWER(CONCAT(?, COALESCE(?, ?))) AS ?",
			bun.Ident("status.content"), bun.Ident("status.content_warning"), "",
			bun.Ident("status_text"))

	default:
		panic("db conn was neither pg not sqlite")
	}

	return statusText
}
