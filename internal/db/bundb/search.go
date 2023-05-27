package bundb

import (
	"context"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

type searchDB struct {
	conn  *DBConn
	state *state.State
}

const (
	escapeChar    = "\\"
	zeroOrMore    = "%"
	exactlyOne    = "_"
	escEscapeChar = escapeChar + escapeChar
	escZeroOrMore = escapeChar + zeroOrMore
	escExactlyOne = escapeChar + exactlyOne
)

var replacer = strings.NewReplacer(
	escapeChar, escEscapeChar,
	zeroOrMore, escZeroOrMore,
	exactlyOne, escExactlyOne,
)

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
	accountIDs := make([]string, 0, limit)

	q := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		// Select only IDs from table
		Column("account.id").
		// Try to ignore instance accounts.
		Where("? != ?", bun.Ident("account.domain"), bun.Ident("account.username")).
		// Sort by ID. Account ID's are random so
		// this is not 'newest first' or anything.
		Order("account.id DESC")

	// Return only items with a LOWER id than maxID.
	if maxID == "" {
		maxID = id.Highest
	}
	q = q.Where("? < ?", bun.Ident("account.id"), maxID)

	if minID != "" {
		// Return only items with a HIGHER id than minID.
		q = q.Where("? > ?", bun.Ident("account.id"), minID)
	}

	if following {
		// Select only from accounts followed by accountID.
		q = q.Where(
			"? IN (?)",
			bun.Ident("account.id"),
			s.followedAccountIDs(accountID),
		)
	}

	// Concatenate account username, displayname,
	// and bio/note (only if following). The main
	// query will search within this subquery.
	q = q.Where(
		"(?) LIKE ? ESCAPE ?",
		s.accountText(following),
		normalizeQuery(query),
		escapeChar,
	)

	if limit > 0 {
		// Limit amount of accounts returned.
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &accountIDs); err != nil {
		return nil, s.conn.ProcessError(err)
	}

	if len(accountIDs) == 0 {
		return nil, db.ErrNoEntries
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
	statusIDs := make([]string, 0, limit)

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
		Where("? IS NULL", bun.Ident("status.boost_of_id")).
		// Sort newest -> oldest.
		Order("status.id DESC")

	// Return only items with a LOWER id than maxID.
	if maxID == "" {
		maxID = id.Highest
	}
	q = q.Where("? < ?", bun.Ident("status.id"), maxID)

	if minID != "" {
		// Return only items with a HIGHER id than minID.
		q = q.Where("? > ?", bun.Ident("status.id"), minID)
	}

	// Concatenate status content warning
	// and content. The main query will
	// search within this subquery.
	q = q.Where(
		"(?) LIKE ? ESCAPE ?",
		s.statusText(),
		normalizeQuery(query),
		escapeChar,
	)

	if limit > 0 {
		// Limit amount of statuses returned.
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, s.conn.ProcessError(err)
	}

	if len(statusIDs) == 0 {
		return nil, db.ErrNoEntries
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

func normalizeQuery(query string) string {
	// Escape existing wildcard + escape chars.
	query = replacer.Replace(query)

	// Add our own wildcards back in.
	query = zeroOrMore + query + zeroOrMore

	return query
}

func getPlaceHolders(count int, sep string) string {
	// Fill a slice with placeholder chars.
	placeHolders := make([]string, count)
	for i := 0; i < count; i++ {
		placeHolders[i] = "?"
	}

	// Join them with provided separator.
	return strings.Join(placeHolders, sep)
}

func (s *searchDB) followedAccountIDs(accountID string) *bun.SelectQuery {
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
		sb          strings.Builder
	)

	if following {
		// If querying for accounts we follow,
		// include note in text search params.
		args = []interface{}{
			bun.Ident("account.username"),
			bun.Ident("account.display_name"),
			bun.Ident("account.note"),
			bun.Ident("account_text"),
		}
	} else {
		// If querying for accounts we're not following,
		// don't include note in text search params.
		args = []interface{}{
			bun.Ident("account.username"),
			bun.Ident("account.display_name"),
			bun.Ident("account_text"),
		}
	}

	// SQLite and Postgres use different
	// syntaxes for concatenation.
	switch s.conn.Dialect().Name() {

	case dialect.SQLite:
		// Produce something like:
		// "LOWER(? || ? || ?) AS ?"
		placeHolders := getPlaceHolders(len(args)-1, " || ")
		_, _ = sb.WriteString("LOWER(")
		_, _ = sb.WriteString(placeHolders)
		_, _ = sb.WriteString(") AS ?")
		query = sb.String()

	case dialect.PG:
		// Produce something like:
		// "LOWER(CONCAT(?, ?, ?)) AS ?"
		placeHolders := getPlaceHolders(len(args)-1, ", ")
		_, _ = sb.WriteString("LOWER(CONCAT(")
		_, _ = sb.WriteString(placeHolders)
		_, _ = sb.WriteString(")) AS ?")
		query = sb.String()

	default:
		panic("db conn was neither pg not sqlite")
	}

	accountText = accountText.ColumnExpr(query, args...)
	return accountText
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
