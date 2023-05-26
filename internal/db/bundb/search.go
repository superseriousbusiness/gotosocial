package bundb

import (
	"context"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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

func normalizeQuery(query string) string {
	// Escape existing wildcard + escape chars.
	query = replacer.Replace(query)

	// Add our own wildcards back in.
	query = zeroOrMore + query + zeroOrMore

	return query
}

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

	// Assemble a subquery that lowercases + concatenates
	// account username, displayname, and bio/note. The
	// main query will search within this subquery.
	accountText := s.conn.NewSelect()
	switch s.conn.Dialect().Name() {
	case dialect.SQLite:
		accountText = accountText.ColumnExpr(
			"LOWER(? || ? || ?) AS ?",
			bun.Ident("account.username"), bun.Ident("account.display_name"), bun.Ident("account.note"),
			bun.Ident("account_text"))
	case dialect.PG:
		accountText = accountText.ColumnExpr(
			"LOWER(CONCAT(?, ?, ?)) AS ?",
			bun.Ident("account.username"), bun.Ident("account.display_name"), bun.Ident("account.note"),
			bun.Ident("account_text"))
	default:
		panic("db conn was neither pg not sqlite")
	}

	q := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		// Select only IDs from table
		Column("account.id").
		// Search within accountText using the provided query.
		Where("(?) LIKE ? ESCAPE ?", accountText, normalizeQuery(query), escapeChar).
		Order("account.id DESC")

	if following {
		// Subquery to select targetAccountID
		// from all follows owned by this account.
		followedAccountIDs := s.conn.
			NewSelect().
			TableExpr("? AS ?", bun.Ident("follows"), bun.Ident("follow")).
			Column("follow.target_account_id").
			Where("? = ?", bun.Ident("follow.account_id"), accountID)

		// Only select from accounts that this accountID follows.
		q = q.Where("? IN (?)", bun.Ident("account.id"), followedAccountIDs)
	}

	if limit > 0 {
		// limit amount of statuses returned
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

func (s *searchDB) SearchForStatuses(
	ctx context.Context,
	accountID string,
	query string,
	maxID string,
	minID string,
	limit int,
	following bool,
	offset int,
) ([]*gtsmodel.Status, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	statusIDs := make([]string, 0, limit)

	// Assemble a subquery that lowercases + concatenates
	// status content-warning and note. The main query
	// will search within this subquery.
	statusText := s.conn.NewSelect()
	switch s.conn.Dialect().Name() {
	case dialect.SQLite:
		statusText = statusText.ColumnExpr(
			"LOWER(? || COALESCE(?, ?)) AS ?",
			bun.Ident("status.content_warning"), "", bun.Ident("status.note"),
			bun.Ident("status_text"))
	case dialect.PG:
		statusText = statusText.ColumnExpr(
			"LOWER(CONCAT(?, COALESCE(?, ?))) AS ?",
			bun.Ident("status.content_warning"), "", bun.Ident("status.note"),
			bun.Ident("status_text"))
	default:
		panic("db conn was neither pg not sqlite")
	}

	q := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		// Select only IDs from table
		Column("status.id").
		// Search within statusText using the provided query.
		Where("(?) LIKE ? ESCAPE ?", statusText, normalizeQuery(query), escapeChar).
		Order("status.id DESC")

	if following {
		// Subquery to select targetAccountID
		// from all follows owned by this account.
		followedAccountIDs := s.conn.
			NewSelect().
			TableExpr("? AS ?", bun.Ident("follows"), bun.Ident("follow")).
			Column("follow.target_account_id").
			Where("? = ?", bun.Ident("follow.account_id"), accountID)

		// Only select statuses from accounts that this accountID follows.
		q = q.Where("? IN (?)", bun.Ident("status.account_id"), followedAccountIDs)
	}

	if limit > 0 {
		// limit amount of statuses returned
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, s.conn.ProcessError(err)
	}

	if len(statusIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	accounts := make([]*gtsmodel.Account, 0, len(statusIDs))
	for _, id := range statusIDs {
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
