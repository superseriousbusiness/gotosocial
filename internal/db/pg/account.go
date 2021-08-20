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

package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type accountDB struct {
	config *config.Config
	conn   *pg.DB
	log    *logrus.Logger
	cancel context.CancelFunc
}

func (a *accountDB) newAccountQ(account *gtsmodel.Account) *orm.Query {
	return a.conn.Model(account).
		Relation("AvatarMediaAttachment").
		Relation("HeaderMediaAttachment")
}

func (a *accountDB) GetAccountByID(id string) (*gtsmodel.Account, db.Error) {
	account := &gtsmodel.Account{}

	q := a.newAccountQ(account).
		Where("account.id = ?", id)

	err := processErrorResponse(q.Select())

	return account, err
}

func (a *accountDB) GetAccountByURI(uri string) (*gtsmodel.Account, db.Error) {
	account := &gtsmodel.Account{}

	q := a.newAccountQ(account).
		Where("account.uri = ?", uri)

	err := processErrorResponse(q.Select())

	return account, err
}

func (a *accountDB) GetAccountByURL(uri string) (*gtsmodel.Account, db.Error) {
	account := &gtsmodel.Account{}

	q := a.newAccountQ(account).
		Where("account.url = ?", uri)

	err := processErrorResponse(q.Select())

	return account, err
}

func (a *accountDB) GetInstanceAccount(domain string) (*gtsmodel.Account, db.Error) {
	account := &gtsmodel.Account{}

	q := a.newAccountQ(account)

	if domain == "" {
		q = q.
			Where("account.username = ?", domain).
			Where("account.domain = ?", domain)
	} else {
		q = q.
			Where("account.username = ?", domain).
			Where("? IS NULL", pg.Ident("domain"))
	}

	err := processErrorResponse(q.Select())

	return account, err
}

func (a *accountDB) GetAccountLastPosted(accountID string) (time.Time, db.Error) {
	status := &gtsmodel.Status{}

	q := a.conn.Model(status).
		Order("id DESC").
		Limit(1).
		Where("account_id = ?", accountID).
		Column("created_at")

	err := processErrorResponse(q.Select())

	return status.CreatedAt, err
}

func (a *accountDB) SetAccountHeaderOrAvatar(mediaAttachment *gtsmodel.MediaAttachment, accountID string) db.Error {
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
	if _, err := a.conn.Model(mediaAttachment).OnConflict("(id) DO UPDATE").Insert(); err != nil {
		return err
	}

	if _, err := a.conn.Model(&gtsmodel.Account{}).Set(fmt.Sprintf("%s_media_attachment_id = ?", headerOrAVI), mediaAttachment.ID).Where("id = ?", accountID).Update(); err != nil {
		return err
	}
	return nil
}

func (a *accountDB) GetLocalAccountByUsername(username string) (*gtsmodel.Account, db.Error) {
	account := &gtsmodel.Account{}

	q := a.newAccountQ(account).
		Where("username = ?", username).
		Where("? IS NULL", pg.Ident("domain"))

	err := processErrorResponse(q.Select())

	return account, err
}

func (a *accountDB) GetAccountFaves(accountID string) ([]*gtsmodel.StatusFave, db.Error) {
	faves := []*gtsmodel.StatusFave{}

	if err := a.conn.Model(&faves).
		Where("account_id = ?", accountID).
		Select(); err != nil {
		if err == pg.ErrNoRows {
			return faves, nil
		}
		return nil, err
	}
	return faves, nil
}

func (a *accountDB) CountAccountStatuses(accountID string) (int, db.Error) {
	return a.conn.Model(&gtsmodel.Status{}).Where("account_id = ?", accountID).Count()
}

func (a *accountDB) GetAccountStatuses(accountID string, limit int, excludeReplies bool, maxID string, pinnedOnly bool, mediaOnly bool) ([]*gtsmodel.Status, db.Error) {
	a.log.Debugf("getting statuses for account %s", accountID)
	statuses := []*gtsmodel.Status{}

	q := a.conn.Model(&statuses).Order("id DESC")
	if accountID != "" {
		q = q.Where("account_id = ?", accountID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	if excludeReplies {
		q = q.Where("? IS NULL", pg.Ident("in_reply_to_id"))
	}

	if pinnedOnly {
		q = q.Where("pinned = ?", true)
	}

	if mediaOnly {
		q = q.WhereGroup(func(q *pg.Query) (*pg.Query, error) {
			return q.Where("? IS NOT NULL", pg.Ident("attachments")).Where("attachments != '{}'"), nil
		})
	}

	if maxID != "" {
		q = q.Where("id < ?", maxID)
	}

	if err := q.Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil, db.ErrNoEntries
		}
		return nil, err
	}

	if len(statuses) == 0 {
		return nil, db.ErrNoEntries
	}

	a.log.Debugf("returning statuses for account %s", accountID)
	return statuses, nil
}

func (a *accountDB) GetAccountBlocks(accountID string, maxID string, sinceID string, limit int) ([]*gtsmodel.Account, string, string, db.Error) {
	blocks := []*gtsmodel.Block{}

	fq := a.conn.Model(&blocks).
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

	err := fq.Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, "", "", db.ErrNoEntries
		}
		return nil, "", "", err
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
