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
)

type relationshipDB struct {
	config *config.Config
	conn   *pg.DB
	log    *logrus.Logger
	cancel context.CancelFunc
}

func (r *relationshipDB) newBlockQ(block *gtsmodel.Block) *orm.Query {
	return r.conn.Model(block).
		Relation("Account").
		Relation("TargetAccount")
}

func (r *relationshipDB) processResponse(block *gtsmodel.Block, err error) (*gtsmodel.Block, db.Error) {
	switch err {
	case pg.ErrNoRows:
		return nil, db.ErrNoEntries
	case nil:
		return block, nil
	default:
		return nil, err
	}
}

func (r *relationshipDB) Blocked(account1 string, account2 string, eitherDirection bool) (bool, db.Error) {
	q := r.conn.Model(&gtsmodel.Block{}).Where("account_id = ?", account1).Where("target_account_id = ?", account2)

	if eitherDirection {
		q = q.WhereOr("target_account_id = ?", account1).Where("account_id = ?", account2)
	}

	return q.Exists()
}

func (r *relationshipDB) GetBlock(account1 string, account2 string) (*gtsmodel.Block, db.Error) {
	block := &gtsmodel.Block{}

	q := r.newBlockQ(block).
		Where("block.account_id = ?", account1).
		Where("block.target_account_id = ?", account2)

	return r.processResponse(block, q.Select())
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

func (r *relationshipDB) Follows(sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, db.Error) {
	if sourceAccount == nil || targetAccount == nil {
		return false, nil
	}

	return r.conn.Model(&gtsmodel.Follow{}).Where("account_id = ?", sourceAccount.ID).Where("target_account_id = ?", targetAccount.ID).Exists()
}

func (r *relationshipDB) FollowRequested(sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, db.Error) {
	if sourceAccount == nil || targetAccount == nil {
		return false, nil
	}

	return r.conn.Model(&gtsmodel.FollowRequest{}).Where("account_id = ?", sourceAccount.ID).Where("target_account_id = ?", targetAccount.ID).Exists()
}

func (r *relationshipDB) Mutuals(account1 *gtsmodel.Account, account2 *gtsmodel.Account) (bool, db.Error) {
	if account1 == nil || account2 == nil {
		return false, nil
	}

	// make sure account 1 follows account 2
	f1, err := r.conn.Model(&gtsmodel.Follow{}).Where("account_id = ?", account1.ID).Where("target_account_id = ?", account2.ID).Exists()
	if err != nil {
		if err == pg.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	// make sure account 2 follows account 1
	f2, err := r.conn.Model(&gtsmodel.Follow{}).Where("account_id = ?", account2.ID).Where("target_account_id = ?", account1.ID).Exists()
	if err != nil {
		if err == pg.ErrNoRows {
			return false, nil
		}
		return false, err
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
