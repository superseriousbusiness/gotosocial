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

package bundb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type accountDB struct {
	config *config.Config
	conn   *DBConn
}

func (a *accountDB) newAccountQ(account *gtsmodel.Account) *bun.SelectQuery {
	return a.conn.
		NewSelect().
		Model(account).
		Relation("AvatarMediaAttachment").
		Relation("HeaderMediaAttachment")
}

func (a *accountDB) GetAccountByID(ctx context.Context, id string) (*gtsmodel.Account, db.Error) {
	account := new(gtsmodel.Account)

	q := a.newAccountQ(account).
		Where("account.id = ?", id)

	err := q.Scan(ctx)
	if err != nil {
		return nil, a.conn.ProcessError(err)
	}
	return account, nil
}

func (a *accountDB) GetAccountByURI(ctx context.Context, uri string) (*gtsmodel.Account, db.Error) {
	account := new(gtsmodel.Account)

	q := a.newAccountQ(account).
		Where("account.uri = ?", uri)

	err := q.Scan(ctx)
	if err != nil {
		return nil, a.conn.ProcessError(err)
	}
	return account, nil
}

func (a *accountDB) GetAccountByURL(ctx context.Context, uri string) (*gtsmodel.Account, db.Error) {
	account := new(gtsmodel.Account)

	q := a.newAccountQ(account).
		Where("account.url = ?", uri)

	err := q.Scan(ctx)
	if err != nil {
		return nil, a.conn.ProcessError(err)
	}
	return account, nil
}

func (a *accountDB) UpdateAccount(ctx context.Context, account *gtsmodel.Account) (*gtsmodel.Account, db.Error) {
	if strings.TrimSpace(account.ID) == "" {
		return nil, errors.New("account had no ID")
	}

	account.UpdatedAt = time.Now()

	q := a.conn.
		NewUpdate().
		Model(account).
		WherePK()

	_, err := q.Exec(ctx)
	if err != nil {
		return nil, a.conn.ProcessError(err)
	}
	return account, nil
}

func (a *accountDB) GetInstanceAccount(ctx context.Context, domain string) (*gtsmodel.Account, db.Error) {
	account := new(gtsmodel.Account)

	q := a.newAccountQ(account)

	if domain == "" {
		q = q.
			Where("account.username = ?", domain).
			Where("account.domain = ?", domain)
	} else {
		q = q.
			Where("account.username = ?", domain).
			WhereGroup(" AND ", whereEmptyOrNull("domain"))
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, a.conn.ProcessError(err)
	}
	return account, nil
}

func (a *accountDB) GetAccountLastPosted(ctx context.Context, accountID string) (time.Time, db.Error) {
	status := new(gtsmodel.Status)

	q := a.conn.
		NewSelect().
		Model(status).
		Order("id DESC").
		Limit(1).
		Where("account_id = ?", accountID).
		Column("created_at")

	err := q.Scan(ctx)
	if err != nil {
		return time.Time{}, a.conn.ProcessError(err)
	}
	return status.CreatedAt, nil
}

func (a *accountDB) SetAccountHeaderOrAvatar(ctx context.Context, mediaAttachment *gtsmodel.MediaAttachment, accountID string) db.Error {
	if mediaAttachment.Avatar && mediaAttachment.Header {
		return errors.New("one media attachment cannot be both header and avatar")
	}

	var headerOrAVI string
	if mediaAttachment.Avatar {
		headerOrAVI = "avatar"
	} else if mediaAttachment.Header {
		headerOrAVI = "header"
	} else {
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
		Model(&gtsmodel.Account{}).
		Set(fmt.Sprintf("%s_media_attachment_id = ?", headerOrAVI), mediaAttachment.ID).
		Where("id = ?", accountID).
		Exec(ctx); err != nil {
		return a.conn.ProcessError(err)
	}

	return nil
}

func (a *accountDB) GetLocalAccountByUsername(ctx context.Context, username string) (*gtsmodel.Account, db.Error) {
	account := new(gtsmodel.Account)

	q := a.newAccountQ(account).
		Where("username = ?", username).
		WhereGroup(" AND ", whereEmptyOrNull("domain"))

	err := q.Scan(ctx)
	if err != nil {
		return nil, a.conn.ProcessError(err)
	}
	return account, nil
}

func (a *accountDB) GetAccountFaves(ctx context.Context, accountID string) ([]*gtsmodel.StatusFave, db.Error) {
	faves := new([]*gtsmodel.StatusFave)

	if err := a.conn.
		NewSelect().
		Model(faves).
		Where("account_id = ?", accountID).
		Scan(ctx); err != nil {
		return nil, a.conn.ProcessError(err)
	}

	return *faves, nil
}

func (a *accountDB) CountAccountStatuses(ctx context.Context, accountID string) (int, db.Error) {
	return a.conn.
		NewSelect().
		Model(&gtsmodel.Status{}).
		Where("account_id = ?", accountID).
		Count(ctx)
}

func (a *accountDB) GetAccountStatuses(ctx context.Context, accountID string, limit int, excludeReplies bool, maxID string, pinnedOnly bool, mediaOnly bool) ([]*gtsmodel.Status, db.Error) {
	statuses := []*gtsmodel.Status{}

	q := a.conn.
		NewSelect().
		Model(&statuses).
		Order("id DESC")

	if accountID != "" {
		q = q.Where("account_id = ?", accountID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	if pinnedOnly {
		q = q.Where("pinned = ?", true)
	}

	if maxID != "" {
		q = q.Where("id < ?", maxID)
	}

	if mediaOnly {
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				WhereOr("? IS NOT NULL", bun.Ident("attachments")).
				WhereOr("attachments != '{}'")
		})
	}

	if excludeReplies {
		q = q.WhereGroup(" AND ", whereEmptyOrNull("in_reply_to_id"))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, a.conn.ProcessError(err)
	}

	if len(statuses) == 0 {
		return nil, db.ErrNoEntries
	}

	return statuses, nil
}

func (a *accountDB) GetAccountBlocks(ctx context.Context, accountID string, maxID string, sinceID string, limit int) ([]*gtsmodel.Account, string, string, db.Error) {
	blocks := []*gtsmodel.Block{}

	fq := a.conn.
		NewSelect().
		Model(&blocks).
		Where("block.account_id = ?", accountID).
		Relation("TargetAccount").
		Order("block.id DESC")

	if maxID != "" {
		fq = fq.Where("block.id < ?", maxID)
	}

	if sinceID != "" {
		fq = fq.Where("block.id > ?", sinceID)
	}

	if limit > 0 {
		fq = fq.Limit(limit)
	}

	err := fq.Scan(ctx)
	if err != nil {
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
