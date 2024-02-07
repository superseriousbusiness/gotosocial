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
	"strings"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
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
	accounts := make([]*gtsmodel.Account, 0, len(ids))

	for _, id := range ids {
		// Attempt to fetch account from DB.
		account, err := a.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			id,
		)
		if err != nil {
			log.Errorf(ctx, "error getting account %q: %v", id, err)
			continue
		}

		// Append account to return slice.
		accounts = append(accounts, account)
	}

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

func (a *accountDB) GetAccountByURL(ctx context.Context, url string) (*gtsmodel.Account, error) {
	return a.getAccount(
		ctx,
		"URL",
		func(account *gtsmodel.Account) error {
			return a.db.NewSelect().
				Model(account).
				Where("? = ?", bun.Ident("account.url"), url).
				Scan(ctx)
		},
		url,
	)
}

func (a *accountDB) GetAccountByUsernameDomain(ctx context.Context, username string, domain string) (*gtsmodel.Account, error) {
	if domain != "" {
		// Normalize the domain as punycode
		var err error
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

func (a *accountDB) GetAccountByInboxURI(ctx context.Context, uri string) (*gtsmodel.Account, error) {
	return a.getAccount(
		ctx,
		"InboxURI",
		func(account *gtsmodel.Account) error {
			return a.db.NewSelect().
				Model(account).
				Where("? = ?", bun.Ident("account.inbox_uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (a *accountDB) GetAccountByOutboxURI(ctx context.Context, uri string) (*gtsmodel.Account, error) {
	return a.getAccount(
		ctx,
		"OutboxURI",
		func(account *gtsmodel.Account) error {
			return a.db.NewSelect().
				Model(account).
				Where("? = ?", bun.Ident("account.outbox_uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (a *accountDB) GetAccountByFollowersURI(ctx context.Context, uri string) (*gtsmodel.Account, error) {
	return a.getAccount(
		ctx,
		"FollowersURI",
		func(account *gtsmodel.Account) error {
			return a.db.NewSelect().
				Model(account).
				Where("? = ?", bun.Ident("account.followers_uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (a *accountDB) GetAccountByFollowingURI(ctx context.Context, uri string) (*gtsmodel.Account, error) {
	return a.getAccount(
		ctx,
		"FollowingURI",
		func(account *gtsmodel.Account) error {
			return a.db.NewSelect().
				Model(account).
				Where("? = ?", bun.Ident("account.following_uri"), uri).
				Scan(ctx)
		},
		uri,
	)
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

func (a *accountDB) getAccount(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Account) error, keyParts ...any) (*gtsmodel.Account, error) {
	// Fetch account from database cache with loader callback
	account, err := a.state.Caches.GTS.Account.LoadOne(lookup, func() (*gtsmodel.Account, error) {
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

	if account.MovedTo == nil && account.MovedToURI != "" {
		// Account movedTo is not set, fetch from database.
		account.MovedTo, err = a.state.DB.GetAccountByURI(
			gtscontext.SetBarebones(ctx),
			account.MovedToURI,
		)
		if err != nil {
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

	return errs.Combine()
}

func (a *accountDB) PutAccount(ctx context.Context, account *gtsmodel.Account) error {
	return a.state.Caches.GTS.Account.Store(account, func() error {
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

	return a.state.Caches.GTS.Account.Store(account, func() error {
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
	defer a.state.Caches.GTS.Account.Invalidate("ID", id)

	// Load account into cache before attempting a delete,
	// as we need it cached in order to trigger the invalidate
	// callback. This in turn invalidates others.
	_, err := a.GetAccountByID(gtscontext.SetBarebones(ctx), id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// NOTE: even if db.ErrNoEntries is returned, we
		// still run the below transaction to ensure related
		// objects are appropriately deleted.
		return err
	}

	return a.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
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
			TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
			Where("? = ?", bun.Ident("account.id"), id).
			Exec(ctx)
		return err
	})
}

func (a *accountDB) GetAccountLastPosted(ctx context.Context, accountID string, webOnly bool) (time.Time, error) {
	createdAt := time.Time{}

	q := a.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		Column("status.created_at").
		Where("? = ?", bun.Ident("status.account_id"), accountID).
		Order("status.id DESC").
		Limit(1)

	if webOnly {
		q = q.
			Where("? IS NULL", bun.Ident("status.in_reply_to_uri")).
			Where("? IS NULL", bun.Ident("status.boost_of_id")).
			Where("? = ?", bun.Ident("status.visibility"), gtsmodel.VisibilityPublic).
			Where("? = ?", bun.Ident("status.federated"), true)
	}

	if err := q.Scan(ctx, &createdAt); err != nil {
		return time.Time{}, err
	}
	return createdAt, nil
}

func (a *accountDB) SetAccountHeaderOrAvatar(ctx context.Context, mediaAttachment *gtsmodel.MediaAttachment, accountID string) error {
	if *mediaAttachment.Avatar && *mediaAttachment.Header {
		return errors.New("one media attachment cannot be both header and avatar")
	}

	var column bun.Ident
	switch {
	case *mediaAttachment.Avatar:
		column = bun.Ident("account.avatar_media_attachment_id")
	case *mediaAttachment.Header:
		column = bun.Ident("account.header_media_attachment_id")
	default:
		return errors.New("given media attachment was neither a header nor an avatar")
	}

	// TODO: there are probably more side effects here that need to be handled
	if _, err := a.db.
		NewInsert().
		Model(mediaAttachment).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := a.db.
		NewUpdate().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		Set("? = ?", column, mediaAttachment.ID).
		Where("? = ?", bun.Ident("account.id"), accountID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (a *accountDB) GetAccountCustomCSSByUsername(ctx context.Context, username string) (string, error) {
	account, err := a.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return "", err
	}

	return account.CustomCSS, nil
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

func (a *accountDB) CountAccountStatuses(ctx context.Context, accountID string) (int, error) {
	return a.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		Where("? = ?", bun.Ident("status.account_id"), accountID).
		Count(ctx)
}

func (a *accountDB) CountAccountPinned(ctx context.Context, accountID string) (int, error) {
	return a.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		Where("? = ?", bun.Ident("status.account_id"), accountID).
		Where("? IS NOT NULL", bun.Ident("status.pinned_at")).
		Count(ctx)
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
		q = q.WhereGroup(" AND ", func(*bun.SelectQuery) *bun.SelectQuery {
			return q.
				// Do include self replies (threads), but
				// don't include replies to other people.
				Where("? = ?", bun.Ident("status.in_reply_to_account_id"), accountID).
				WhereOr("? IS NULL", bun.Ident("status.in_reply_to_uri"))
		})
	}

	if excludeReblogs {
		q = q.Where("? IS NULL", bun.Ident("status.boost_of_id"))
	}

	if mediaOnly {
		// Attachments are stored as a json object; this
		// implementation differs between SQLite and Postgres,
		// so we have to be thorough to cover all eventualities
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			switch a.db.Dialect().Name() {
			case dialect.PG:
				return q.
					Where("? IS NOT NULL", bun.Ident("status.attachments")).
					Where("? != '{}'", bun.Ident("status.attachments"))
			case dialect.SQLite:
				return q.
					Where("? IS NOT NULL", bun.Ident("status.attachments")).
					Where("? != ''", bun.Ident("status.attachments")).
					Where("? != 'null'", bun.Ident("status.attachments")).
					Where("? != '{}'", bun.Ident("status.attachments")).
					Where("? != '[]'", bun.Ident("status.attachments"))
			default:
				log.Panic(ctx, "db dialect was neither pg nor sqlite")
				return q
			}
		})
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

func (a *accountDB) GetAccountWebStatuses(ctx context.Context, accountID string, limit int, maxID string) ([]*gtsmodel.Status, error) {
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
		Where("? = ?", bun.Ident("status.account_id"), accountID).
		// Don't show replies or boosts.
		Where("? IS NULL", bun.Ident("status.in_reply_to_uri")).
		Where("? IS NULL", bun.Ident("status.boost_of_id")).
		// Only Public statuses.
		Where("? = ?", bun.Ident("status.visibility"), gtsmodel.VisibilityPublic).
		// Don't show local-only statuses on the web view.
		Where("? = ?", bun.Ident("status.federated"), true)

	// return only statuses LOWER (ie., older) than maxID
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
