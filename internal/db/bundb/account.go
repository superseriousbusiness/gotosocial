/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"context"
	"errors"
	"strings"
	"time"

	"codeberg.org/gruf/go-cache/v3/result"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

type accountDB struct {
	conn   *DBConn
	cache  *result.Cache[*gtsmodel.Account]
	emojis *emojiDB
	status *statusDB
}

func (a *accountDB) init() {
	// Initialize account result cache
	a.cache = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "URL"},
		{Name: "Username.Domain"},
		{Name: "PublicKeyURI"},
	}, func(a1 *gtsmodel.Account) *gtsmodel.Account {
		a2 := new(gtsmodel.Account)
		*a2 = *a1
		return a2
	}, 1000)

	// Set cache TTL and start sweep routine
	a.cache.SetTTL(time.Minute*5, false)
	a.cache.Start(time.Second * 10)
}

func (a *accountDB) newAccountQ(account *gtsmodel.Account) *bun.SelectQuery {
	return a.conn.
		NewSelect().
		Model(account).
		Relation("AvatarMediaAttachment").
		Relation("HeaderMediaAttachment")
}

func (a *accountDB) GetAccountByID(ctx context.Context, id string) (*gtsmodel.Account, db.Error) {
	return a.getAccount(
		ctx,
		"ID",
		func(account *gtsmodel.Account) error {
			return a.newAccountQ(account).Where("? = ?", bun.Ident("account.id"), id).Scan(ctx)
		},
		id,
	)
}

func (a *accountDB) GetAccountByURI(ctx context.Context, uri string) (*gtsmodel.Account, db.Error) {
	return a.getAccount(
		ctx,
		"URI",
		func(account *gtsmodel.Account) error {
			return a.newAccountQ(account).Where("? = ?", bun.Ident("account.uri"), uri).Scan(ctx)
		},
		uri,
	)
}

func (a *accountDB) GetAccountByURL(ctx context.Context, url string) (*gtsmodel.Account, db.Error) {
	return a.getAccount(
		ctx,
		"URL",
		func(account *gtsmodel.Account) error {
			return a.newAccountQ(account).Where("? = ?", bun.Ident("account.url"), url).Scan(ctx)
		},
		url,
	)
}

func (a *accountDB) GetAccountByUsernameDomain(ctx context.Context, username string, domain string) (*gtsmodel.Account, db.Error) {
	username = strings.ToLower(username)
	return a.getAccount(
		ctx,
		"Username.Domain",
		func(account *gtsmodel.Account) error {
			q := a.newAccountQ(account)

			if domain != "" {
				q = q.Where("? = ?", bun.Ident("account.username"), username)
				q = q.Where("? = ?", bun.Ident("account.domain"), domain)
			} else {
				q = q.Where("? = ?", bun.Ident("account.username"), username)
				q = q.Where("? IS NULL", bun.Ident("account.domain"))
			}

			return q.Scan(ctx)
		},
		username,
		domain,
	)
}

func (a *accountDB) GetAccountByPubkeyID(ctx context.Context, id string) (*gtsmodel.Account, db.Error) {
	return a.getAccount(
		ctx,
		"PublicKeyURI",
		func(account *gtsmodel.Account) error {
			return a.newAccountQ(account).Where("? = ?", bun.Ident("account.public_key_uri"), id).Scan(ctx)
		},
		id,
	)
}

func (a *accountDB) GetInstanceAccount(ctx context.Context, domain string) (*gtsmodel.Account, db.Error) {
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

func (a *accountDB) getAccount(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Account) error, keyParts ...any) (*gtsmodel.Account, db.Error) {
	// Fetch account from database cache with loader callback
	account, err := a.cache.Load(lookup, func() (*gtsmodel.Account, error) {
		var account gtsmodel.Account

		// Not cached! Perform database query
		if err := dbQuery(&account); err != nil {
			return nil, a.conn.ProcessError(err)
		}

		return &account, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if len(account.EmojiIDs) > 0 {
		// Set the account's related emojis
		account.Emojis, err = a.emojis.emojisFromIDs(ctx, account.EmojiIDs)
		if err != nil {
			return nil, err
		}
	}

	return account, nil
}

func (a *accountDB) PutAccount(ctx context.Context, account *gtsmodel.Account) db.Error {
	return a.cache.Store(account, func() error {
		// It is safe to run this database transaction within cache.Store
		// as the cache does not attempt a mutex lock until AFTER hook.
		//
		return a.conn.RunInTx(ctx, func(tx bun.Tx) error {
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

func (a *accountDB) UpdateAccount(ctx context.Context, account *gtsmodel.Account) db.Error {
	// Update the account's last-updated
	account.UpdatedAt = time.Now()

	return a.cache.Store(account, func() error {
		// It is safe to run this database transaction within cache.Store
		// as the cache does not attempt a mutex lock until AFTER hook.
		//
		return a.conn.RunInTx(ctx, func(tx bun.Tx) error {
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
				Exec(ctx)
			return err
		})
	})
}

func (a *accountDB) DeleteAccount(ctx context.Context, id string) db.Error {
	if err := a.conn.RunInTx(ctx, func(tx bun.Tx) error {
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
	}); err != nil {
		return err
	}

	a.cache.Invalidate("ID", id)
	return nil
}

func (a *accountDB) GetAccountLastPosted(ctx context.Context, accountID string, webOnly bool) (time.Time, db.Error) {
	createdAt := time.Time{}

	q := a.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		Column("status.created_at").
		Where("? = ?", bun.Ident("status.account_id"), accountID).
		Order("status.id DESC").
		Limit(1)

	if webOnly {
		q = q.
			WhereGroup(" AND ", whereEmptyOrNull("status.in_reply_to_uri")).
			WhereGroup(" AND ", whereEmptyOrNull("status.boost_of_id")).
			Where("? = ?", bun.Ident("status.visibility"), gtsmodel.VisibilityPublic).
			Where("? = ?", bun.Ident("status.federated"), true)
	}

	if err := q.Scan(ctx, &createdAt); err != nil {
		return time.Time{}, a.conn.ProcessError(err)
	}
	return createdAt, nil
}

func (a *accountDB) SetAccountHeaderOrAvatar(ctx context.Context, mediaAttachment *gtsmodel.MediaAttachment, accountID string) db.Error {
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
	if _, err := a.conn.
		NewInsert().
		Model(mediaAttachment).
		Exec(ctx); err != nil {
		return a.conn.ProcessError(err)
	}

	if _, err := a.conn.
		NewUpdate().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		Set("? = ?", column, mediaAttachment.ID).
		Where("? = ?", bun.Ident("account.id"), accountID).
		Exec(ctx); err != nil {
		return a.conn.ProcessError(err)
	}

	return nil
}

func (a *accountDB) GetAccountCustomCSSByUsername(ctx context.Context, username string) (string, db.Error) {
	account, err := a.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return "", err
	}

	return account.CustomCSS, nil
}

func (a *accountDB) GetAccountFaves(ctx context.Context, accountID string) ([]*gtsmodel.StatusFave, db.Error) {
	faves := new([]*gtsmodel.StatusFave)

	if err := a.conn.
		NewSelect().
		Model(faves).
		Where("? = ?", bun.Ident("status_fave.account_id"), accountID).
		Scan(ctx); err != nil {
		return nil, a.conn.ProcessError(err)
	}

	return *faves, nil
}

func (a *accountDB) CountAccountStatuses(ctx context.Context, accountID string) (int, db.Error) {
	return a.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		Where("? = ?", bun.Ident("status.account_id"), accountID).
		Count(ctx)
}

func (a *accountDB) GetAccountStatuses(ctx context.Context, accountID string, limit int, excludeReplies bool, excludeReblogs bool, maxID string, minID string, pinnedOnly bool, mediaOnly bool, publicOnly bool) ([]*gtsmodel.Status, db.Error) {
	statusIDs := []string{}

	q := a.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		Column("status.id").
		Order("status.id DESC")

	if accountID != "" {
		q = q.Where("? = ?", bun.Ident("status.account_id"), accountID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	if excludeReplies {
		// include self-replies (threads)
		whereGroup := func(*bun.SelectQuery) *bun.SelectQuery {
			return q.
				WhereOr("? = ?", bun.Ident("status.in_reply_to_account_id"), accountID).
				WhereGroup(" OR ", whereEmptyOrNull("status.in_reply_to_uri"))
		}

		q = q.WhereGroup(" AND ", whereGroup)
	}

	if excludeReblogs {
		q = q.WhereGroup(" AND ", whereEmptyOrNull("status.boost_of_id"))
	}

	if maxID != "" {
		q = q.Where("? < ?", bun.Ident("status.id"), maxID)
	}

	if minID != "" {
		q = q.Where("? > ?", bun.Ident("status.id"), minID)
	}

	if pinnedOnly {
		q = q.Where("? = ?", bun.Ident("status.pinned"), true)
	}

	if mediaOnly {
		// attachments are stored as a json object;
		// this implementation differs between sqlite and postgres,
		// so we have to be thorough to cover all eventualities
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			switch a.conn.Dialect().Name() {
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
				log.Panic("db dialect was neither pg nor sqlite")
				return q
			}
		})
	}

	if publicOnly {
		q = q.Where("? = ?", bun.Ident("status.visibility"), gtsmodel.VisibilityPublic)
	}

	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, a.conn.ProcessError(err)
	}

	return a.statusesFromIDs(ctx, statusIDs)
}

func (a *accountDB) GetAccountWebStatuses(ctx context.Context, accountID string, limit int, maxID string) ([]*gtsmodel.Status, db.Error) {
	statusIDs := []string{}

	q := a.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		Column("status.id").
		Where("? = ?", bun.Ident("status.account_id"), accountID).
		WhereGroup(" AND ", whereEmptyOrNull("status.in_reply_to_uri")).
		WhereGroup(" AND ", whereEmptyOrNull("status.boost_of_id")).
		Where("? = ?", bun.Ident("status.visibility"), gtsmodel.VisibilityPublic).
		Where("? = ?", bun.Ident("status.federated"), true)

	if maxID != "" {
		q = q.Where("? < ?", bun.Ident("status.id"), maxID)
	}

	q = q.Limit(limit).Order("status.id DESC")

	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, a.conn.ProcessError(err)
	}

	return a.statusesFromIDs(ctx, statusIDs)
}

func (a *accountDB) GetAccountBlocks(ctx context.Context, accountID string, maxID string, sinceID string, limit int) ([]*gtsmodel.Account, string, string, db.Error) {
	blocks := []*gtsmodel.Block{}

	fq := a.conn.
		NewSelect().
		Model(&blocks).
		Where("? = ?", bun.Ident("block.account_id"), accountID).
		Relation("TargetAccount").
		Order("block.id DESC")

	if maxID != "" {
		fq = fq.Where("? < ?", bun.Ident("block.id"), maxID)
	}

	if sinceID != "" {
		fq = fq.Where("? > ?", bun.Ident("block.id"), sinceID)
	}

	if limit > 0 {
		fq = fq.Limit(limit)
	}

	if err := fq.Scan(ctx); err != nil {
		return nil, "", "", a.conn.ProcessError(err)
	}

	if len(blocks) == 0 {
		return nil, "", "", db.ErrNoEntries
	}

	accounts := []*gtsmodel.Account{}
	for _, b := range blocks {
		accounts = append(accounts, b.TargetAccount)
	}

	nextMaxID := blocks[len(blocks)-1].ID
	prevMinID := blocks[0].ID
	return accounts, nextMaxID, prevMinID, nil
}

func (a *accountDB) statusesFromIDs(ctx context.Context, statusIDs []string) ([]*gtsmodel.Status, db.Error) {
	// Catch case of no statuses early
	if len(statusIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	// Allocate return slice (will be at most len statusIDS)
	statuses := make([]*gtsmodel.Status, 0, len(statusIDs))

	for _, id := range statusIDs {
		// Fetch from status from database by ID
		status, err := a.status.GetStatusByID(ctx, id)
		if err != nil {
			log.Errorf("statusesFromIDs: error getting status %q: %v", id, err)
			continue
		}

		// Append to return slice
		statuses = append(statuses, status)
	}

	return statuses, nil
}
