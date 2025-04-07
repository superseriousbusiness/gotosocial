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
	"errors"
	"fmt"
	"net/netip"
	"slices"
	"strings"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

type accountDB struct {
	db    *bun.DB
	state *state.State
}

func (a *accountDB) GetAccountByID(ctx context.Context, id string) (*gtsmodel.Account, error) {
	return a.getAccount(
		ctx,
		"ID",
		func(account *gtsmodel.Account) error {
			return a.db.NewSelect().
				Model(account).
				Where("? = ?", bun.Ident("account.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (a *accountDB) GetAccountsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Account, error) {
	// Load all input account IDs via cache loader callback.
	accounts, err := a.state.Caches.DB.Account.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.Account, error) {
			// Preallocate expected length of uncached accounts.
			accounts := make([]*gtsmodel.Account, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) account IDs.
			if err := a.db.NewSelect().
				Model(&accounts).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return accounts, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the statuses by their
	// IDs to ensure in correct order.
	getID := func(a *gtsmodel.Account) string { return a.ID }
	xslices.OrderBy(accounts, ids, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return accounts, nil
	}

	// Populate all loaded accounts, removing those we fail to
	// populate (removes needing so many nil checks everywhere).
	accounts = slices.DeleteFunc(accounts, func(account *gtsmodel.Account) bool {
		if err := a.PopulateAccount(ctx, account); err != nil {
			log.Errorf(ctx, "error populating account %s: %v", account.ID, err)
			return true
		}
		return false
	})

	return accounts, nil
}

func (a *accountDB) GetAccountByURI(ctx context.Context, uri string) (*gtsmodel.Account, error) {
	return a.getAccount(
		ctx,
		"URI",
		func(account *gtsmodel.Account) error {
			return a.db.NewSelect().
				Model(account).
				Where("? = ?", bun.Ident("account.uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (a *accountDB) GetOneAccountByURL(ctx context.Context, url string) (*gtsmodel.Account, error) {
	// Select IDs of all
	// accounts with this url.
	var ids []string
	if err := a.db.NewSelect().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		Column("account.id").
		Where("? = ?", bun.Ident("account.url"), url).
		Scan(ctx, &ids); err != nil {
		return nil, err
	}

	// Ensure exactly one account.
	if len(ids) == 0 {
		return nil, db.ErrNoEntries
	}
	if len(ids) > 1 {
		return nil, db.ErrMultipleEntries
	}

	return a.GetAccountByID(ctx, ids[0])
}

func (a *accountDB) GetAccountsByURL(ctx context.Context, url string) ([]*gtsmodel.Account, error) {
	// Select IDs of all
	// accounts with this url.
	var ids []string
	if err := a.db.NewSelect().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		Column("account.id").
		Where("? = ?", bun.Ident("account.url"), url).
		Scan(ctx, &ids); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, db.ErrNoEntries
	}

	return a.GetAccountsByIDs(ctx, ids)
}

func (a *accountDB) GetAccountByUsernameDomain(ctx context.Context, username string, domain string) (*gtsmodel.Account, error) {
	if domain != "" {
		var err error

		// Normalize the domain as punycode
		domain, err = util.Punify(domain)
		if err != nil {
			return nil, err
		}
	}

	return a.getAccount(
		ctx,
		"Username,Domain",
		func(account *gtsmodel.Account) error {
			q := a.db.NewSelect().
				Model(account)

			if domain != "" {
				q = q.
					Where("LOWER(?) = ?", bun.Ident("account.username"), strings.ToLower(username)).
					Where("? = ?", bun.Ident("account.domain"), domain)
			} else {
				q = q.
					Where("? = ?", bun.Ident("account.username"), strings.ToLower(username)). // usernames on our instance are always lowercase
					Where("? IS NULL", bun.Ident("account.domain"))
			}

			return q.Scan(ctx)
		},
		username,
		domain,
	)
}

func (a *accountDB) GetAccountByPubkeyID(ctx context.Context, id string) (*gtsmodel.Account, error) {
	return a.getAccount(
		ctx,
		"PublicKeyURI",
		func(account *gtsmodel.Account) error {
			return a.db.NewSelect().
				Model(account).
				Where("? = ?", bun.Ident("account.public_key_uri"), id).
				Scan(ctx)
		},
		id,
	)
}

func (a *accountDB) GetOneAccountByInboxURI(ctx context.Context, inboxURI string) (*gtsmodel.Account, error) {
	// Select IDs of all accounts
	// with this inbox_uri.
	var ids []string
	if err := a.db.NewSelect().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		Column("account.id").
		Where("? = ?", bun.Ident("account.inbox_uri"), inboxURI).
		Scan(ctx, &ids); err != nil {
		return nil, err
	}

	// Ensure exactly one account.
	if len(ids) == 0 {
		return nil, db.ErrNoEntries
	}
	if len(ids) > 1 {
		return nil, db.ErrMultipleEntries
	}

	return a.GetAccountByID(ctx, ids[0])
}

func (a *accountDB) GetOneAccountByOutboxURI(ctx context.Context, outboxURI string) (*gtsmodel.Account, error) {
	// Select IDs of all accounts
	// with this outbox_uri.
	var ids []string
	if err := a.db.NewSelect().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		Column("account.id").
		Where("? = ?", bun.Ident("account.outbox_uri"), outboxURI).
		Scan(ctx, &ids); err != nil {
		return nil, err
	}

	// Ensure exactly one account.
	if len(ids) == 0 {
		return nil, db.ErrNoEntries
	}
	if len(ids) > 1 {
		return nil, db.ErrMultipleEntries
	}

	return a.GetAccountByID(ctx, ids[0])
}

func (a *accountDB) GetInstanceAccount(ctx context.Context, domain string) (*gtsmodel.Account, error) {
	var username string

	if domain == "" {
		// I.e. our local instance account
		username = config.GetHost()
	} else {
		// A remote instance account
		username = domain
	}

	return a.GetAccountByUsernameDomain(ctx, username, domain)
}

func (a *accountDB) GetAccountsByMovedToURI(ctx context.Context, uri string) ([]*gtsmodel.Account, error) {
	var accountIDs []string

	// Find all account IDs with
	// given moved_to_uri column.
	if err := a.db.NewSelect().
		Table("accounts").
		Column("id").
		Where("? = ?", bun.Ident("moved_to_uri"), uri).
		Scan(ctx, &accountIDs); err != nil {
		return nil, err
	}

	if len(accountIDs) == 0 {
		return nil, nil
	}

	// Return account models for all found IDs.
	return a.GetAccountsByIDs(ctx, accountIDs)
}

// GetAccounts selects accounts using the given parameters.
// Unlike with other functions, the paging for GetAccounts
// is done not by ID, but by a concatenation of `[domain]/@[username]`,
// which allows callers to page through accounts in alphabetical
// order (much more useful for an admin overview of accounts,
// for example, than paging by ID (which is random) or by account
// created at date, which is not particularly interesting).
//
// Generated queries will look something like this
// (SQLite example, maxID was provided so we're paging down):
//
//	SELECT "account"."id", (COALESCE("domain", '') || '/@' || "username") AS "domain_username"
//	FROM "accounts" AS "account"
//	WHERE ("domain_username" > '/@the_mighty_zork')
//	ORDER BY "domain_username" ASC
//
// **NOTE ABOUT POSTGRES**: Postgres ordering expressions in
// this function specify COLLATE "C" to ensure that ordering
// is similar to SQLite (which uses BINARY ordering by default).
// This unfortunately means that A-Z > a-z, when ordering but
// that's an acceptable tradeoff for a query like this.
//
// See:
//
//   - https://www.postgresql.org/docs/current/collation.html#COLLATION-MANAGING-STANDARD
//   - https://sqlite.org/datatype3.html#collation
func (a *accountDB) GetAccounts(
	ctx context.Context,
	origin string,
	status string,
	mods bool,
	invitedBy string,
	username string,
	displayName string,
	domain string,
	email string,
	ip netip.Addr,
	page *paging.Page,
) (
	[]*gtsmodel.Account,
	error,
) {
	var (
		// We have to use different
		// syntax for this query
		// depending on dialect.
		dbDialect = a.db.Dialect().Name()

		// local users lists,
		// required for some
		// limiting parameters.
		users []*gtsmodel.User

		// lazyLoadUsers only loads the users
		// slice if it's required by params.
		lazyLoadUsers = func() (err error) {
			if users == nil {
				users, err = a.state.DB.GetAllUsers(gtscontext.SetBarebones(ctx))
				if err != nil {
					return fmt.Errorf("error getting users: %w", err)
				}
			}
			return nil
		}

		// Get paging params.
		minID = page.GetMin()
		maxID = page.GetMax()
		limit = page.GetLimit()
		order = page.GetOrder()

		// Make educated guess for slice size
		accountIDs  = make([]string, 0, limit)
		accountIDIn []string

		useAccountIDIn bool
	)

	q := a.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		// Select only IDs from table
		Column("account.id")

	var subQ *bun.RawQuery
	if dbDialect == dialect.SQLite {
		// For SQLite we can just select
		// our indexed expression once
		// as a column alias.
		q = q.ColumnExpr(
			"(COALESCE(?, ?) || ? || ?) AS ?",
			bun.Ident("domain"), "",
			"/@",
			bun.Ident("username"),
			bun.Ident("domain_username"),
		)
	} else {
		// Create a subquery for
		// Postgres to reuse.
		subQ = a.db.NewRaw(
			"(COALESCE(?, ?) || ? || ?) COLLATE ?",
			bun.Ident("domain"), "",
			"/@",
			bun.Ident("username"),
			bun.Ident("C"),
		)
	}

	// Return only accounts with `[domain]/@[username]`
	// later in the alphabet (a-z) than provided maxID.
	if maxID != "" {
		if dbDialect == dialect.SQLite {
			// Use aliased column.
			q = q.Where("? > ?", bun.Ident("domain_username"), maxID)
		} else {
			q = q.Where("? > ?", subQ, maxID)
		}
	}

	// Return only accounts with `[domain]/@[username]`
	// earlier in the alphabet (a-z) than provided minID.
	if minID != "" {
		if dbDialect == dialect.SQLite {
			// Use aliased column.
			q = q.Where("? < ?", bun.Ident("domain_username"), minID)
		} else {
			q = q.Where("? < ?", subQ, minID)
		}
	}

	switch status {

	case "active":
		// Get only enabled accounts.
		if err := lazyLoadUsers(); err != nil {
			return nil, err
		}
		for _, user := range users {
			if !*user.Disabled {
				accountIDIn = append(accountIDIn, user.AccountID)
			}
		}
		useAccountIDIn = true

	case "pending":
		// Get only unapproved accounts.
		if err := lazyLoadUsers(); err != nil {
			return nil, err
		}
		for _, user := range users {
			if !*user.Approved {
				accountIDIn = append(accountIDIn, user.AccountID)
			}
		}
		useAccountIDIn = true

	case "disabled":
		// Get only disabled accounts.
		if err := lazyLoadUsers(); err != nil {
			return nil, err
		}
		for _, user := range users {
			if *user.Disabled {
				accountIDIn = append(accountIDIn, user.AccountID)
			}
		}
		useAccountIDIn = true

	case "silenced":
		// Get only silenced accounts.
		q = q.Where("? IS NOT NULL", bun.Ident("account.silenced_at"))

	case "suspended":
		// Get only suspended accounts.
		q = q.Where("? IS NOT NULL", bun.Ident("account.suspended_at"))
	}

	if mods {
		// Get only mod accounts.
		if err := lazyLoadUsers(); err != nil {
			return nil, err
		}
		for _, user := range users {
			if *user.Moderator || *user.Admin {
				accountIDIn = append(accountIDIn, user.AccountID)
			}
		}
		useAccountIDIn = true
	}

	// TODO: invitedBy

	if username != "" {
		q = q.Where("? = ?", bun.Ident("account.username"), username)
	}

	if displayName != "" {
		q = q.Where("? = ?", bun.Ident("account.display_name"), displayName)
	}

	if domain != "" {
		q = q.Where("? = ?", bun.Ident("account.domain"), domain)
	}

	if email != "" {
		if err := lazyLoadUsers(); err != nil {
			return nil, err
		}
		for _, user := range users {
			if user.Email == email || user.UnconfirmedEmail == email {
				accountIDIn = append(accountIDIn, user.AccountID)
			}
		}
		useAccountIDIn = true
	}

	// Use ip if not zero value.
	if ip.IsValid() {
		if err := lazyLoadUsers(); err != nil {
			return nil, err
		}
		for _, user := range users {
			if user.SignUpIP.String() == ip.String() {
				accountIDIn = append(accountIDIn, user.AccountID)
			}
		}
		useAccountIDIn = true
	}

	if origin == "local" && !useAccountIDIn {
		// In the case we're not already limiting
		// by specific subset of account IDs, just
		// use existing list of user.AccountIDs
		// instead of adding WHERE to the query.
		if err := lazyLoadUsers(); err != nil {
			return nil, err
		}
		for _, user := range users {
			accountIDIn = append(accountIDIn, user.AccountID)
		}
		useAccountIDIn = true

	} else if origin == "remote" {
		if useAccountIDIn {
			// useAccountIDIn specifically indicates
			// a parameter that limits querying to
			// local accounts, there will be none.
			return nil, nil
		}

		// Get only remote accounts.
		q = q.Where("? IS NOT NULL", bun.Ident("account.domain"))
	}

	if useAccountIDIn {
		if len(accountIDIn) == 0 {
			// There will be no
			// possible answer.
			return nil, nil
		}

		q = q.Where("? IN (?)", bun.Ident("account.id"), bun.In(accountIDIn))
	}

	if limit > 0 {
		// Limit amount of
		// accounts returned.
		q = q.Limit(limit)
	}

	if order == paging.OrderAscending {
		// Page up.
		// It's counterintuitive because it
		// says DESC in the query, but we're
		// going backwards in the alphabet,
		// and a < z in a string comparison.
		if dbDialect == dialect.SQLite {
			q = q.OrderExpr("? DESC", bun.Ident("domain_username"))
		} else {
			q = q.OrderExpr("(?) DESC", subQ)
		}
	} else {
		// Page down.
		// It's counterintuitive because it
		// says ASC in the query, but we're
		// going forwards in the alphabet,
		// and z > a in a string comparison.
		if dbDialect == dialect.SQLite {
			q = q.OrderExpr("? ASC", bun.Ident("domain_username"))
		} else {
			q = q.OrderExpr("? ASC", subQ)
		}
	}

	if err := q.Scan(ctx, &accountIDs, new([]string)); err != nil {
		return nil, err
	}

	if len(accountIDs) == 0 {
		return nil, nil
	}

	// If we're paging up, we still want accounts
	// to be sorted by createdAt desc, so reverse ids slice.
	if order == paging.OrderAscending {
		slices.Reverse(accountIDs)
	}

	// Return account IDs loaded from cache + db.
	return a.state.DB.GetAccountsByIDs(ctx, accountIDs)
}

func (a *accountDB) getAccount(
	ctx context.Context,
	lookup string,
	dbQuery func(*gtsmodel.Account) error, keyParts ...any,
) (*gtsmodel.Account, error) {
	// Fetch account from database cache with loader callback
	account, err := a.state.Caches.DB.Account.LoadOne(lookup, func() (*gtsmodel.Account, error) {
		var account gtsmodel.Account

		// Not cached! Perform database query
		if err := dbQuery(&account); err != nil {
			return nil, err
		}

		return &account, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return account, nil
	}

	// Further populate the account fields where applicable.
	if err := a.PopulateAccount(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

func (a *accountDB) PopulateAccount(ctx context.Context, account *gtsmodel.Account) error {
	var (
		err  error
		errs = gtserror.NewMultiError(5)
	)

	if account.AvatarMediaAttachment == nil && account.AvatarMediaAttachmentID != "" {
		// Account avatar attachment is not set, fetch from database.
		account.AvatarMediaAttachment, err = a.state.DB.GetAttachmentByID(
			ctx, // these are already barebones
			account.AvatarMediaAttachmentID,
		)
		if err != nil {
			errs.Appendf("error populating account avatar: %w", err)
		}
	}

	if account.HeaderMediaAttachment == nil && account.HeaderMediaAttachmentID != "" {
		// Account header attachment is not set, fetch from database.
		account.HeaderMediaAttachment, err = a.state.DB.GetAttachmentByID(
			ctx, // these are already barebones
			account.HeaderMediaAttachmentID,
		)
		if err != nil {
			errs.Appendf("error populating account header: %w", err)
		}
	}

	// Only try to populate AlsoKnownAs for local accounts,
	// since those are the only accounts to which it's relevant.
	//
	// AKA from remotes might have loads of random-ass values
	// set here, and we don't want to do lots of failing DB calls.
	if account.IsLocal() && !account.AlsoKnownAsPopulated() {
		// Account alsoKnownAs accounts are
		// out-of-date with URIs, repopulate.
		alsoKnownAs := make([]*gtsmodel.Account, 0)
		for _, uri := range account.AlsoKnownAsURIs {
			akaAcct, err := a.state.DB.GetAccountByURI(
				gtscontext.SetBarebones(ctx),
				uri,
			)
			if err != nil {
				errs.Appendf("error populating also known as account %s: %w", uri, err)
				continue
			}

			alsoKnownAs = append(alsoKnownAs, akaAcct)
		}

		account.AlsoKnownAs = alsoKnownAs
	}

	if account.Move == nil && account.MoveID != "" {
		// Account move is not set, fetch from database.
		account.Move, err = a.state.DB.GetMoveByID(
			ctx,
			account.MoveID,
		)
		if err != nil {
			errs.Appendf("error populating move: %w", err)
		}
	}

	if account.MovedTo == nil && account.MovedToURI != "" {
		// Account movedTo is not set, try to fetch from database,
		// but only error on real errors since this field is optional.
		account.MovedTo, err = a.state.DB.GetAccountByURI(
			gtscontext.SetBarebones(ctx),
			account.MovedToURI,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			errs.Appendf("error populating moved to account: %w", err)
		}
	}

	if !account.EmojisPopulated() {
		// Account emojis are out-of-date with IDs, repopulate.
		account.Emojis, err = a.state.DB.GetEmojisByIDs(
			ctx, // these are already barebones
			account.EmojiIDs,
		)
		if err != nil {
			errs.Appendf("error populating account emojis: %w", err)
		}
	}

	if account.IsLocal() && account.Settings == nil && !account.IsInstance() {
		// Account settings not set, fetch from db.
		account.Settings, err = a.state.DB.GetAccountSettings(
			ctx, // these are already barebones
			account.ID,
		)
		if err != nil {
			errs.Appendf("error populating account settings: %w", err)
		}
	}

	// Get / Create stats for this account (handles case of already set).
	if err := a.state.DB.PopulateAccountStats(ctx, account); err != nil {
		errs.Appendf("error populating account stats: %w", err)
	}

	return errs.Combine()
}

func (a *accountDB) PutAccount(ctx context.Context, account *gtsmodel.Account) error {
	return a.state.Caches.DB.Account.Store(account, func() error {
		// It is safe to run this database transaction within cache.Store
		// as the cache does not attempt a mutex lock until AFTER hook.
		//
		return a.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// create links between this account and any emojis it uses
			for _, i := range account.EmojiIDs {
				if _, err := tx.NewInsert().Model(&gtsmodel.AccountToEmoji{
					AccountID: account.ID,
					EmojiID:   i,
				}).Exec(ctx); err != nil {
					return err
				}
			}

			// insert the account
			_, err := tx.NewInsert().Model(account).Exec(ctx)
			return err
		})
	})
}

func (a *accountDB) UpdateAccount(ctx context.Context, account *gtsmodel.Account, columns ...string) error {
	account.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	return a.state.Caches.DB.Account.Store(account, func() error {
		// It is safe to run this database transaction within cache.Store
		// as the cache does not attempt a mutex lock until AFTER hook.
		//
		return a.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// create links between this account and any emojis it uses
			// first clear out any old emoji links
			if _, err := tx.
				NewDelete().
				TableExpr("? AS ?", bun.Ident("account_to_emojis"), bun.Ident("account_to_emoji")).
				Where("? = ?", bun.Ident("account_to_emoji.account_id"), account.ID).
				Exec(ctx); err != nil {
				return err
			}

			// now populate new emoji links
			for _, i := range account.EmojiIDs {
				if _, err := tx.
					NewInsert().
					Model(&gtsmodel.AccountToEmoji{
						AccountID: account.ID,
						EmojiID:   i,
					}).Exec(ctx); err != nil {
					return err
				}
			}

			// update the account
			_, err := tx.NewUpdate().
				Model(account).
				Where("? = ?", bun.Ident("account.id"), account.ID).
				Column(columns...).
				Exec(ctx)
			return err
		})
	})
}

func (a *accountDB) DeleteAccount(ctx context.Context, id string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted gtsmodel.Account
	deleted.ID = id

	// Delete account from database and any related links in a transaction.
	if err := a.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

		// clear out any emoji links
		if _, err := tx.
			NewDelete().
			TableExpr("? AS ?", bun.Ident("account_to_emojis"), bun.Ident("account_to_emoji")).
			Where("? = ?", bun.Ident("account_to_emoji.account_id"), id).
			Exec(ctx); err != nil {
			return err
		}

		// delete the account
		_, err := tx.
			NewDelete().
			Model(&deleted).
			Where("? = ?", bun.Ident("id"), id).
			Returning("?", bun.Ident("uri")).
			Exec(ctx)
		return err
	}); err != nil {
		return err
	}

	// Invalidate cached account by its ID, manually
	// call invalidate hook in case not cached.
	a.state.Caches.DB.Account.Invalidate("ID", id)
	a.state.Caches.OnInvalidateAccount(&deleted)

	return nil
}

func (a *accountDB) GetAccountCustomCSSByUsername(ctx context.Context, username string) (string, error) {
	// Get local account.
	account, err := a.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return "", err
	}

	// Ensure settings populated, in case
	// barebones context was passed.
	if account.Settings == nil {
		account.Settings, err = a.GetAccountSettings(ctx, account.ID)
		if err != nil {
			return "", err
		}
	}

	return account.Settings.CustomCSS, nil
}

func (a *accountDB) GetAccountsUsingEmoji(ctx context.Context, emojiID string) ([]*gtsmodel.Account, error) {
	var accountIDs []string

	// SELECT all accounts using this emoji,
	// using a relational table for improved perf.
	if _, err := a.db.NewSelect().
		Table("account_to_emojis").
		Column("account_id").
		Where("? = ?", bun.Ident("emoji_id"), emojiID).
		Exec(ctx, &accountIDs); err != nil {
		return nil, err
	}

	// Convert account IDs into account objects.
	return a.GetAccountsByIDs(ctx, accountIDs)
}

func (a *accountDB) GetAccountFaves(ctx context.Context, accountID string) ([]*gtsmodel.StatusFave, error) {
	faves := new([]*gtsmodel.StatusFave)

	if err := a.db.
		NewSelect().
		Model(faves).
		Where("? = ?", bun.Ident("status_fave.account_id"), accountID).
		Scan(ctx); err != nil {
		return nil, err
	}

	return *faves, nil
}

func qMediaOnly(q *bun.SelectQuery) *bun.SelectQuery {
	// Attachments are stored as a json object; this
	// implementation differs between SQLite and Postgres,
	// so we have to be thorough to cover all eventualities
	return q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
		switch d := q.Dialect().Name(); d {
		case dialect.PG:
			return q.
				Where("? IS NOT NULL", bun.Ident("status.attachments")).
				Where("? != '{}'", bun.Ident("status.attachments"))

		case dialect.SQLite:
			return q.
				Where("? IS NOT NULL", bun.Ident("status.attachments")).
				Where("? != 'null'", bun.Ident("status.attachments")).
				Where("? != '[]'", bun.Ident("status.attachments"))

		default:
			panic("dialect " + d.String() + " was neither pg nor sqlite")
		}
	})
}

func (a *accountDB) GetAccountStatuses(ctx context.Context, accountID string, limit int, excludeReplies bool, excludeReblogs bool, maxID string, minID string, mediaOnly bool, publicOnly bool) ([]*gtsmodel.Status, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	var (
		statusIDs   = make([]string, 0, limit)
		frontToBack = true
	)

	q := a.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		// Select only IDs from table
		Column("status.id").
		Where("? = ?", bun.Ident("status.account_id"), accountID)

	if excludeReplies {
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			// We're excluding replies so
			// only include posts if they:
			return q.
				// Don't reply to anything OR
				Where("? IS NULL", bun.Ident("status.in_reply_to_uri")).
				// reply to self AND don't mention
				// anyone (ie., self-reply threads).
				WhereGroup(" OR ", func(q *bun.SelectQuery) *bun.SelectQuery {
					q = q.Where("? = ?", bun.Ident("status.in_reply_to_account_id"), accountID)
					q = whereArrayIsNullOrEmpty(q, bun.Ident("status.mentions"))
					return q
				})
		})
	}

	if excludeReblogs {
		q = q.Where("? IS NULL", bun.Ident("status.boost_of_id"))
	}

	// Respect media-only preference.
	if mediaOnly {
		q = qMediaOnly(q)
	}

	if publicOnly {
		q = q.Where("? = ?", bun.Ident("status.visibility"), gtsmodel.VisibilityPublic)
	}

	// return only statuses LOWER (ie., older) than maxID
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

	if limit > 0 {
		// limit amount of statuses returned
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
		return nil, err
	}

	if len(statusIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	// If we're paging up, we still want statuses
	// to be sorted by ID desc, so reverse ids slice.
	// https://zchee.github.io/golang-wiki/SliceTricks/#reversing
	if !frontToBack {
		for l, r := 0, len(statusIDs)-1; l < r; l, r = l+1, r-1 {
			statusIDs[l], statusIDs[r] = statusIDs[r], statusIDs[l]
		}
	}

	return a.state.DB.GetStatusesByIDs(ctx, statusIDs)
}

func (a *accountDB) GetAccountPinnedStatuses(ctx context.Context, accountID string) ([]*gtsmodel.Status, error) {
	statusIDs := []string{}

	q := a.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		Column("status.id").
		Where("? = ?", bun.Ident("status.account_id"), accountID).
		Where("? IS NOT NULL", bun.Ident("status.pinned_at")).
		Order("status.pinned_at DESC")

	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, err
	}

	if len(statusIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	return a.state.DB.GetStatusesByIDs(ctx, statusIDs)
}

func (a *accountDB) GetAccountWebStatuses(
	ctx context.Context,
	account *gtsmodel.Account,
	mediaOnly bool,
	limit int,
	maxID string,
) ([]*gtsmodel.Status, error) {
	if account.Username == config.GetHost() {
		// Instance account
		// doesn't post statuses.
		return nil, nil
	}

	// Check for an easy case: account exposes no statuses via the web.
	webVisibility := account.Settings.WebVisibility
	if webVisibility == gtsmodel.VisibilityNone {
		return nil, db.ErrNoEntries
	}

	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	statusIDs := make([]string, 0, limit)

	q := a.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		// Select only IDs from table
		Column("status.id").
		Where("? = ?", bun.Ident("status.account_id"), account.ID)

	// Select statuses for this account according
	// to their web visibility preference.
	switch webVisibility {

	case gtsmodel.VisibilityPublic:
		// Only Public statuses.
		q = q.Where("? = ?", bun.Ident("status.visibility"), gtsmodel.VisibilityPublic)

	case gtsmodel.VisibilityUnlocked:
		// Public or Unlocked.
		visis := []gtsmodel.Visibility{
			gtsmodel.VisibilityPublic,
			gtsmodel.VisibilityUnlocked,
		}
		q = q.Where("? IN (?)", bun.Ident("status.visibility"), bun.In(visis))

	default:
		return nil, gtserror.Newf(
			"unrecognized web visibility for account %s: %s",
			account.ID, webVisibility,
		)
	}

	// Don't show replies, boosts, or
	// local-only statuses on the web view.
	q = q.
		Where("? IS NULL", bun.Ident("status.in_reply_to_uri")).
		Where("? IS NULL", bun.Ident("status.boost_of_id")).
		Where("? = ?", bun.Ident("status.federated"), true)

	// Respect media-only preference.
	if mediaOnly {
		q = qMediaOnly(q)
	}

	// Return only statuses LOWER (ie., older) than maxID
	if maxID == "" {
		maxID = id.Highest
	}
	q = q.Where("? < ?", bun.Ident("status.id"), maxID)

	if limit > 0 {
		// limit amount of statuses returned
		q = q.Limit(limit)
	}

	if limit > 0 {
		// limit amount of statuses returned
		q = q.Limit(limit)
	}

	q = q.Order("status.id DESC")

	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, err
	}

	if len(statusIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	return a.state.DB.GetStatusesByIDs(ctx, statusIDs)
}

func (a *accountDB) GetAccountSettings(
	ctx context.Context,
	accountID string,
) (*gtsmodel.AccountSettings, error) {
	// Fetch settings from db cache with loader callback.
	return a.state.Caches.DB.AccountSettings.LoadOne(
		"AccountID",
		func() (*gtsmodel.AccountSettings, error) {
			// Not cached! Perform database query.
			var settings gtsmodel.AccountSettings
			if err := a.db.
				NewSelect().
				Model(&settings).
				Where("? = ?", bun.Ident("account_settings.account_id"), accountID).
				Scan(ctx); err != nil {
				return nil, err
			}
			return &settings, nil
		},
		accountID,
	)
}

func (a *accountDB) PutAccountSettings(
	ctx context.Context,
	settings *gtsmodel.AccountSettings,
) error {
	return a.state.Caches.DB.AccountSettings.Store(settings, func() error {
		if _, err := a.db.
			NewInsert().
			Model(settings).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}

func (a *accountDB) UpdateAccountSettings(
	ctx context.Context,
	settings *gtsmodel.AccountSettings,
	columns ...string,
) error {
	return a.state.Caches.DB.AccountSettings.Store(settings, func() error {
		settings.UpdatedAt = time.Now()

		switch {

		case len(columns) != 0:
			// If we're updating by column,
			// ensure "updated_at" is included.
			columns = append(columns, "updated_at")

			// If we're updating web_visibility we should
			// fall through + invalidate visibility cache.
			if !slices.Contains(columns, "web_visibility") {
				break // No need to invalidate.
			}

			// Fallthrough
			// to invalidate.
			fallthrough

		case len(columns) == 0:
			// Status visibility may be changing for this account.
			// Clear the visibility cache for unauthed requesters.
			//
			// todo: invalidate JUST this account's statuses.
			defer a.state.Caches.Visibility.Clear()
		}

		if _, err := a.db.
			NewUpdate().
			Model(settings).
			Column(columns...).
			Where("? = ?", bun.Ident("account_settings.account_id"), settings.AccountID).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}

func (a *accountDB) PopulateAccountStats(ctx context.Context, account *gtsmodel.Account) error {
	if account.Stats != nil {
		// Already populated!
		return nil
	}

	// Fetch stats from db cache with loader callback.
	stats, err := a.state.Caches.DB.AccountStats.LoadOne(
		"AccountID",
		func() (*gtsmodel.AccountStats, error) {
			// Not cached! Perform database query.
			var stats gtsmodel.AccountStats
			if err := a.db.
				NewSelect().
				Model(&stats).
				Where("? = ?", bun.Ident("account_stats.account_id"), account.ID).
				Scan(ctx); err != nil {
				return nil, err
			}
			return &stats, nil
		},
		account.ID,
	)

	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Real error.
		return err
	}

	if stats == nil {
		// Don't have stats yet, generate them.
		return a.RegenerateAccountStats(ctx, account)
	}

	// We have a stats, attach
	// it to the account.
	account.Stats = stats

	// Check if this is a local
	// stats by looking at the
	// account they pertain to.
	if account.IsRemote() {
		// Account is remote. Updating
		// stats for remote accounts is
		// handled in the dereferencer.
		//
		// Nothing more to do!
		return nil
	}

	// Stats account is local, check
	// if we need to regenerate.
	const statsFreshness = 48 * time.Hour
	expiry := stats.RegeneratedAt.Add(statsFreshness)
	if time.Now().After(expiry) {
		// Stats have expired, regenerate them.
		return a.RegenerateAccountStats(ctx, account)
	}

	// Stats are still fresh.
	return nil
}

func (a *accountDB) StubAccountStats(ctx context.Context, account *gtsmodel.Account) error {
	stats := &gtsmodel.AccountStats{
		AccountID:           account.ID,
		RegeneratedAt:       time.Now(),
		FollowersCount:      util.Ptr(0),
		FollowingCount:      util.Ptr(0),
		FollowRequestsCount: util.Ptr(0),
		StatusesCount:       util.Ptr(0),
		StatusesPinnedCount: util.Ptr(0),
	}

	// Upsert this stats in case a race
	// meant someone else inserted it first.
	if err := a.state.Caches.DB.AccountStats.Store(stats, func() error {
		if _, err := NewUpsert(a.db).
			Model(stats).
			Constraint("account_id").
			Exec(ctx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	account.Stats = stats
	return nil
}

func (a *accountDB) RegenerateAccountStats(ctx context.Context, account *gtsmodel.Account) error {
	// Initialize a new stats struct.
	stats := &gtsmodel.AccountStats{
		AccountID:     account.ID,
		RegeneratedAt: time.Now(),
	}

	// Count followers outside of transaction since
	// it uses a cache + requires its own db calls.
	followerIDs, err := a.state.DB.GetAccountFollowerIDs(ctx, account.ID, nil)
	if err != nil {
		return err
	}
	stats.FollowersCount = util.Ptr(len(followerIDs))

	// Count following outside of transaction since
	// it uses a cache + requires its own db calls.
	followIDs, err := a.state.DB.GetAccountFollowIDs(ctx, account.ID, nil)
	if err != nil {
		return err
	}
	stats.FollowingCount = util.Ptr(len(followIDs))

	// Count follow requests outside of transaction since
	// it uses a cache + requires its own db calls.
	followRequestIDs, err := a.state.DB.GetAccountFollowRequestIDs(ctx, account.ID, nil)
	if err != nil {
		return err
	}
	stats.FollowRequestsCount = util.Ptr(len(followRequestIDs))

	// Populate remaining stats struct fields.
	// This can be done inside a transaction.
	if err := a.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var err error

		// Scan database for account statuses, ignoring
		// statuses that are currently pending approval.
		statusesCount, err := tx.NewSelect().
			TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
			Where("? = ?", bun.Ident("status.account_id"), account.ID).
			Where("NOT ? = ?", bun.Ident("status.pending_approval"), true).
			Count(ctx)
		if err != nil {
			return err
		}
		stats.StatusesCount = &statusesCount

		// Scan database for pinned statuses, ignoring
		// statuses that are currently pending approval.
		statusesPinnedCount, err := tx.NewSelect().
			TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
			Where("? = ?", bun.Ident("status.account_id"), account.ID).
			Where("? IS NOT NULL", bun.Ident("status.pinned_at")).
			Where("NOT ? = ?", bun.Ident("status.pending_approval"), true).
			Count(ctx)
		if err != nil {
			return err
		}
		stats.StatusesPinnedCount = &statusesPinnedCount

		// Scan database for last status, ignoring
		// statuses that are currently pending approval.
		lastStatusAt := time.Time{}
		err = tx.
			NewSelect().
			TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
			Column("status.created_at").
			Where("? = ?", bun.Ident("status.account_id"), account.ID).
			Where("NOT ? = ?", bun.Ident("status.pending_approval"), true).
			Order("status.id DESC").
			Limit(1).
			Scan(ctx, &lastStatusAt)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return err
		}
		stats.LastStatusAt = lastStatusAt

		return nil
	}); err != nil {
		return err
	}

	// Upsert this stats in case a race
	// meant someone else inserted it first.
	if err := a.state.Caches.DB.AccountStats.Store(stats, func() error {
		if _, err := NewUpsert(a.db).
			Model(stats).
			Constraint("account_id").
			Exec(ctx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	account.Stats = stats
	return nil
}

func (a *accountDB) UpdateAccountStats(ctx context.Context, stats *gtsmodel.AccountStats, columns ...string) error {
	return a.state.Caches.DB.AccountStats.Store(stats, func() error {
		if _, err := a.db.
			NewUpdate().
			Model(stats).
			Column(columns...).
			Where("? = ?", bun.Ident("account_stats.account_id"), stats.AccountID).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}

func (a *accountDB) DeleteAccountStats(ctx context.Context, accountID string) error {
	defer a.state.Caches.DB.AccountStats.Invalidate("AccountID", accountID)

	if _, err := a.db.
		NewDelete().
		Table("account_stats").
		Where("? = ?", bun.Ident("account_id"), accountID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}
