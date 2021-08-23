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
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type relationshipDB struct {
	config *config.Config
	conn   *bun.DB
	log    *logrus.Logger
	cancel context.CancelFunc
}

func (r *relationshipDB) newBlockQ(block *gtsmodel.Block) *orm.Query {
	return r.conn.Model(block).
		Relation("Account").
		Relation("TargetAccount")
}

func (r *relationshipDB) newFollowQ(follow interface{}) *orm.Query {
	return r.conn.Model(follow).
		Relation("Account").
		Relation("TargetAccount")
}

func (r *relationshipDB) IsBlocked(account1 string, account2 string, eitherDirection bool) (bool, db.Error) {
	q := r.conn.
		Model(&gtsmodel.Block{}).
		Where("account_id = ?", account1).
		Where("target_account_id = ?", account2)

	if eitherDirection {
		q = q.
			WhereOr("target_account_id = ?", account1).
			Where("account_id = ?", account2)
	}

	return q.Exists()
}

func (r *relationshipDB) GetBlock(account1 string, account2 string) (*gtsmodel.Block, db.Error) {
	block := &gtsmodel.Block{}

	q := r.newBlockQ(block).
		Where("block.account_id = ?", account1).
		Where("block.target_account_id = ?", account2)

	err := processErrorResponse(q.Select())

	return block, err
}

func (r *relationshipDB) GetRelationship(requestingAccount string, targetAccount string) (*gtsmodel.Relationship, db.Error) {
	rel := &gtsmodel.Relationship{
		ID: targetAccount,
	}

	// check if the requesting account follows the target account
	follow := &gtsmodel.Follow{}
	if err := r.conn.Model(follow).Where("account_id = ?", requestingAccount).Where("target_account_id = ?", targetAccount).Select(); err != nil {
		if err != pg.ErrNoRows {
			// a proper error
			return nil, fmt.Errorf("getrelationship: error checking follow existence: %s", err)
		}
		// no follow exists so these are all false
		rel.Following = false
		rel.ShowingReblogs = false
		rel.Notifying = false
	} else {
		// follow exists so we can fill these fields out...
		rel.Following = true
		rel.ShowingReblogs = follow.ShowReblogs
		rel.Notifying = follow.Notify
	}

	// check if the target account follows the requesting account
	followedBy, err := r.conn.Model(&gtsmodel.Follow{}).Where("account_id = ?", targetAccount).Where("target_account_id = ?", requestingAccount).Exists()
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking followed_by existence: %s", err)
	}
	rel.FollowedBy = followedBy

	// check if the requesting account blocks the target account
	blocking, err := r.conn.Model(&gtsmodel.Block{}).Where("account_id = ?", requestingAccount).Where("target_account_id = ?", targetAccount).Exists()
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking blocking existence: %s", err)
	}
	rel.Blocking = blocking

	// check if the target account blocks the requesting account
	blockedBy, err := r.conn.Model(&gtsmodel.Block{}).Where("account_id = ?", targetAccount).Where("target_account_id = ?", requestingAccount).Exists()
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking blocked existence: %s", err)
	}
	rel.BlockedBy = blockedBy

	// check if there's a pending following request from requesting account to target account
	requested, err := r.conn.Model(&gtsmodel.FollowRequest{}).Where("account_id = ?", requestingAccount).Where("target_account_id = ?", targetAccount).Exists()
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking blocked existence: %s", err)
	}
	rel.Requested = requested

	return rel, nil
}

func (r *relationshipDB) IsFollowing(sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, db.Error) {
	if sourceAccount == nil || targetAccount == nil {
		return false, nil
	}

	q := r.conn.
		Model(&gtsmodel.Follow{}).
		Where("account_id = ?", sourceAccount.ID).
		Where("target_account_id = ?", targetAccount.ID)

	return q.Exists()
}

func (r *relationshipDB) IsFollowRequested(sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, db.Error) {
	if sourceAccount == nil || targetAccount == nil {
		return false, nil
	}

	q := r.conn.
		Model(&gtsmodel.FollowRequest{}).
		Where("account_id = ?", sourceAccount.ID).
		Where("target_account_id = ?", targetAccount.ID)

	return q.Exists()
}

func (r *relationshipDB) IsMutualFollowing(account1 *gtsmodel.Account, account2 *gtsmodel.Account) (bool, db.Error) {
	if account1 == nil || account2 == nil {
		return false, nil
	}

	// make sure account 1 follows account 2
	f1, err := r.IsFollowing(account1, account2)
	if err != nil {
		return false, processErrorResponse(err)
	}

	// make sure account 2 follows account 1
	f2, err := r.IsFollowing(account2, account1)
	if err != nil {
		return false, processErrorResponse(err)
	}

	return f1 && f2, nil
}

func (r *relationshipDB) AcceptFollowRequest(originAccountID string, targetAccountID string) (*gtsmodel.Follow, db.Error) {
	// make sure the original follow request exists
	fr := &gtsmodel.FollowRequest{}
	if err := r.conn.Model(fr).Where("account_id = ?", originAccountID).Where("target_account_id = ?", targetAccountID).Select(); err != nil {
		if err == pg.ErrMultiRows {
			return nil, db.ErrNoEntries
		}
		return nil, err
	}

	// create a new follow to 'replace' the request with
	follow := &gtsmodel.Follow{
		ID:              fr.ID,
		AccountID:       originAccountID,
		TargetAccountID: targetAccountID,
		URI:             fr.URI,
	}

	// if the follow already exists, just update the URI -- we don't need to do anything else
	if _, err := r.conn.Model(follow).OnConflict("ON CONSTRAINT follows_account_id_target_account_id_key DO UPDATE set uri = ?", follow.URI).Insert(); err != nil {
		return nil, err
	}

	// now remove the follow request
	if _, err := r.conn.Model(&gtsmodel.FollowRequest{}).Where("account_id = ?", originAccountID).Where("target_account_id = ?", targetAccountID).Delete(); err != nil {
		return nil, err
	}

	return follow, nil
}

func (r *relationshipDB) GetAccountFollowRequests(accountID string) ([]*gtsmodel.FollowRequest, db.Error) {
	followRequests := []*gtsmodel.FollowRequest{}

	q := r.newFollowQ(&followRequests).
		Where("target_account_id = ?", accountID)

	err := processErrorResponse(q.Select())

	return followRequests, err
}

func (r *relationshipDB) GetAccountFollows(accountID string) ([]*gtsmodel.Follow, db.Error) {
	follows := []*gtsmodel.Follow{}

	q := r.newFollowQ(&follows).
		Where("account_id = ?", accountID)

	err := processErrorResponse(q.Select())

	return follows, err
}

func (r *relationshipDB) CountAccountFollows(accountID string, localOnly bool) (int, db.Error) {
	return r.conn.
		Model(&[]*gtsmodel.Follow{}).
		Where("account_id = ?", accountID).
		Count()
}

func (r *relationshipDB) GetAccountFollowedBy(accountID string, localOnly bool) ([]*gtsmodel.Follow, db.Error) {

	follows := []*gtsmodel.Follow{}

	q := r.conn.Model(&follows)

	if localOnly {
		// for local accounts let's get where domain is null OR where domain is an empty string, just to be safe
		whereGroup := func(q *pg.Query) (*pg.Query, error) {
			q = q.
				WhereOr("? IS NULL", pg.Ident("a.domain")).
				WhereOr("a.domain = ?", "")
			return q, nil
		}

		q = q.ColumnExpr("follow.*").
			Join("JOIN accounts AS a ON follow.account_id = TEXT(a.id)").
			Where("follow.target_account_id = ?", accountID).
			WhereGroup(whereGroup)
	} else {
		q = q.Where("target_account_id = ?", accountID)
	}

	if err := q.Select(); err != nil {
		if err == pg.ErrNoRows {
			return follows, nil
		}
		return nil, err
	}
	return follows, nil
}

func (r *relationshipDB) CountAccountFollowedBy(accountID string, localOnly bool) (int, db.Error) {
	return r.conn.
		Model(&[]*gtsmodel.Follow{}).
		Where("target_account_id = ?", accountID).
		Count()
}
