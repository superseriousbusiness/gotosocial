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

	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

type accountDB struct {
	conn   *DBConn
	cache  *cache.AccountCache
	status *statusDB
}

func (a *accountDB) newAccountQ(account *gtsmodel.Account) *bun.SelectQuery {
	return a.conn.
		NewSelect().
		Model(account).
		Relation("AvatarMediaAttachment").
		Relation("HeaderMediaAttachment").
		Relation("Emojis")
}

func (a *accountDB) GetAccountByID(ctx context.Context, id string) (*gtsmodel.Account, db.Error) {
	return a.getAccount(
		ctx,
		func() (*gtsmodel.Account, bool) {
			return a.cache.GetByID(id)
		},
		func(account *gtsmodel.Account) error {
			return a.newAccountQ(account).Where("? = ?", bun.Ident("account.id"), id).Scan(ctx)
		},
	)
}

func (a *accountDB) GetAccountByURI(ctx context.Context, uri string) (*gtsmodel.Account, db.Error) {
	return a.getAccount(
		ctx,
		func() (*gtsmodel.Account, bool) {
			return a.cache.GetByURI(uri)
		},
		func(account *gtsmodel.Account) error {
			return a.newAccountQ(account).Where("? = ?", bun.Ident("account.uri"), uri).Scan(ctx)
		},
	)
}

func (a *accountDB) GetAccountByURL(ctx context.Context, url string) (*gtsmodel.Account, db.Error) {
	return a.getAccount(
		ctx,
		func() (*gtsmodel.Account, bool) {
			return a.cache.GetByURL(url)
		},
		func(account *gtsmodel.Account) error {
			return a.newAccountQ(account).Where("? = ?", bun.Ident("account.url"), url).Scan(ctx)
		},
	)
}

func (a *accountDB) GetAccountByUsernameDomain(ctx context.Context, username string, domain string) (*gtsmodel.Account, db.Error) {
	return a.getAccount(
		ctx,
		func() (*gtsmodel.Account, bool) {
			return a.cache.GetByUsernameDomain(username, domain)
		},
		func(account *gtsmodel.Account) error {
			q := a.newAccountQ(account)

			if domain != "" {
				q = q.Where("? = ?", bun.Ident("account.username"), username)
				q = q.Where("? = ?", bun.Ident("account.domain"), domain)
			} else {
				q = q.Where("? = ?", bun.Ident("account.username"), strings.ToLower(username))
				q = q.Where("? IS NULL", bun.Ident("account.domain"))
			}

			return q.Scan(ctx)
		},
	)
}

func (a *accountDB) GetAccountByPubkeyID(ctx context.Context, id string) (*gtsmodel.Account, db.Error) {
	return a.getAccount(
		ctx,
		func() (*gtsmodel.Account, bool) {
			return a.cache.GetByPubkeyID(id)
		},
		func(account *gtsmodel.Account) error {
			return a.newAccountQ(account).Where("? = ?", bun.Ident("account.public_key_uri"), id).Scan(ctx)
		},
	)
}

func (a *accountDB) getAccount(ctx context.Context, cacheGet func() (*gtsmodel.Account, bool), dbQuery func(*gtsmodel.Account) error) (*gtsmodel.Account, db.Error) {
	// Attempt to fetch cached account
	account, cached := cacheGet()

	if !cached {
		account = &gtsmodel.Account{}

		// Not cached! Perform database query
		err := dbQuery(account)
		if err != nil {
			return nil, a.conn.ProcessError(err)
		}

		// Place in the cache
		a.cache.Put(account)
	}

	return account, nil
}

func (a *accountDB) PutAccount(ctx context.Context, account *gtsmodel.Account) (*gtsmodel.Account, db.Error) {
	if err := a.conn.RunInTx(ctx, func(tx bun.Tx) error {
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
	}); err != nil {
		return nil, a.conn.ProcessError(err)
	}

	a.cache.Put(account)
	return account, nil
}

func (a *accountDB) UpdateAccount(ctx context.Context, account *gtsmodel.Account) (*gtsmodel.Account, db.Error) {
	// Update the account's last-updated
	account.UpdatedAt = time.Now()

	if err := a.conn.RunInTx(ctx, func(tx bun.Tx) error {
		// create links between this account and any emojis it uses
		// first clear out any old emoji links
		if _, err := tx.NewDelete().
			TableExpr("? AS ?", bun.Ident("account_to_emojis"), bun.Ident("account_to_emoji")).
			Where("? = ?", bun.Ident("account_to_emoji.account_id"), account.ID).
			Exec(ctx); err != nil {
			return err
		}

		// now populate new emoji links
		for _, i := range account.EmojiIDs {
			if _, err := tx.NewInsert().Model(&gtsmodel.AccountToEmoji{
				AccountID: account.ID,
				EmojiID:   i,
			}).Exec(ctx); err != nil {
				return err
			}
		}

		// update the account
		_, err := tx.NewUpdate().Model(account).WherePK().Exec(ctx)
		return err
	}); err != nil {
		return nil, a.conn.ProcessError(err)
	}

	a.cache.Put(account)
	return account, nil
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
			NewUpdate().
			Model(&gtsmodel.Account{ID: id}).
			WherePK().
			Exec(ctx)
		return err
	}); err != nil {
		return a.conn.ProcessError(err)
	}

	a.cache.Invalidate(id)
	return nil
}

func (a *accountDB) GetInstanceAccount(ctx context.Context, domain string) (*gtsmodel.Account, db.Error) {
	account := new(gtsmodel.Account)

	q := a.newAccountQ(account)

	if domain != "" {
		q = q.
			Where("? = ?", bun.Ident("account.username"), domain).
			Where("? = ?", bun.Ident("account.domain"), domain)
	} else {
		q = q.
			Where("? = ?", bun.Ident("account.username"), config.GetHost()).
			WhereGroup(" AND ", whereEmptyOrNull("domain"))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, a.conn.ProcessError(err)
	}
	return account, nil
}

func (a *accountDB) GetAccountLastPosted(ctx context.Context, accountID string) (time.Time, db.Error) {
	status := new(gtsmodel.Status)

	q := a.conn.
		NewSelect().
		Model(status).
		Column("status.created_at").
		Where("? = ?", bun.Ident("status.account_id"), accountID).
		Order("status.id DESC").
		Limit(1)

	if err := q.Scan(ctx); err != nil {
		return time.Time{}, a.conn.ProcessError(err)
	}
	return status.CreatedAt, nil
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
		Model(&gtsmodel.Status{}).
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
		Model(&gtsmodel.Status{}).
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
