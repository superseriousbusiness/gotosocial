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
	"database/sql"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type relationshipDB struct {
	config *config.Config
	conn   *DBConn
}

func (r *relationshipDB) newBlockQ(block *gtsmodel.Block) *bun.SelectQuery {
	return r.conn.
		NewSelect().
		Model(block).
		Relation("Account").
		Relation("TargetAccount")
}

func (r *relationshipDB) newFollowQ(follow interface{}) *bun.SelectQuery {
	return r.conn.
		NewSelect().
		Model(follow).
		Relation("Account").
		Relation("TargetAccount")
}

func (r *relationshipDB) IsBlocked(ctx context.Context, account1 string, account2 string, eitherDirection bool) (bool, db.Error) {
	q := r.conn.
		NewSelect().
		Model(&gtsmodel.Block{}).
		Where("account_id = ?", account1).
		Where("target_account_id = ?", account2).
		Limit(1)

	if eitherDirection {
		q = q.
			WhereOr("target_account_id = ?", account1).
			Where("account_id = ?", account2)
	}

	return r.conn.Exists(ctx, q)
}

func (r *relationshipDB) GetBlock(ctx context.Context, account1 string, account2 string) (*gtsmodel.Block, db.Error) {
	block := &gtsmodel.Block{}

	q := r.newBlockQ(block).
		Where("block.account_id = ?", account1).
		Where("block.target_account_id = ?", account2)

	err := q.Scan(ctx)
	if err != nil {
		return nil, r.conn.ProcessError(err)
	}
	return block, nil
}

func (r *relationshipDB) GetRelationship(ctx context.Context, requestingAccount string, targetAccount string) (*gtsmodel.Relationship, db.Error) {
	rel := &gtsmodel.Relationship{
		ID: targetAccount,
	}

	// check if the requesting account follows the target account
	follow := &gtsmodel.Follow{}
	if err := r.conn.
		NewSelect().
		Model(follow).
		Where("account_id = ?", requestingAccount).
		Where("target_account_id = ?", targetAccount).
		Limit(1).
		Scan(ctx); err != nil {
		if err != sql.ErrNoRows {
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
	count, err := r.conn.
		NewSelect().
		Model(&gtsmodel.Follow{}).
		Where("account_id = ?", targetAccount).
		Where("target_account_id = ?", requestingAccount).
		Limit(1).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking followed_by existence: %s", err)
	}
	rel.FollowedBy = count > 0

	// check if the requesting account blocks the target account
	count, err = r.conn.NewSelect().
		Model(&gtsmodel.Block{}).
		Where("account_id = ?", requestingAccount).
		Where("target_account_id = ?", targetAccount).
		Limit(1).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking blocking existence: %s", err)
	}
	rel.Blocking = count > 0

	// check if the target account blocks the requesting account
	count, err = r.conn.
		NewSelect().
		Model(&gtsmodel.Block{}).
		Where("account_id = ?", targetAccount).
		Where("target_account_id = ?", requestingAccount).
		Limit(1).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking blocked existence: %s", err)
	}
	rel.BlockedBy = count > 0

	// check if there's a pending following request from requesting account to target account
	count, err = r.conn.
		NewSelect().
		Model(&gtsmodel.FollowRequest{}).
		Where("account_id = ?", requestingAccount).
		Where("target_account_id = ?", targetAccount).
		Limit(1).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking blocked existence: %s", err)
	}
	rel.Requested = count > 0

	return rel, nil
}

func (r *relationshipDB) IsFollowing(ctx context.Context, sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, db.Error) {
	if sourceAccount == nil || targetAccount == nil {
		return false, nil
	}

	q := r.conn.
		NewSelect().
		Model(&gtsmodel.Follow{}).
		Where("account_id = ?", sourceAccount.ID).
		Where("target_account_id = ?", targetAccount.ID).
		Limit(1)

	return r.conn.Exists(ctx, q)
}

func (r *relationshipDB) IsFollowRequested(ctx context.Context, sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, db.Error) {
	if sourceAccount == nil || targetAccount == nil {
		return false, nil
	}

	q := r.conn.
		NewSelect().
		Model(&gtsmodel.FollowRequest{}).
		Where("account_id = ?", sourceAccount.ID).
		Where("target_account_id = ?", targetAccount.ID)

	return r.conn.Exists(ctx, q)
}

func (r *relationshipDB) IsMutualFollowing(ctx context.Context, account1 *gtsmodel.Account, account2 *gtsmodel.Account) (bool, db.Error) {
	if account1 == nil || account2 == nil {
		return false, nil
	}

	// make sure account 1 follows account 2
	f1, err := r.IsFollowing(ctx, account1, account2)
	if err != nil {
		return false, err
	}

	// make sure account 2 follows account 1
	f2, err := r.IsFollowing(ctx, account2, account1)
	if err != nil {
		return false, err
	}

	return f1 && f2, nil
}

func (r *relationshipDB) AcceptFollowRequest(ctx context.Context, originAccountID string, targetAccountID string) (*gtsmodel.Follow, db.Error) {
	// make sure the original follow request exists
	fr := &gtsmodel.FollowRequest{}
	if err := r.conn.
		NewSelect().
		Model(fr).
		Where("account_id = ?", originAccountID).
		Where("target_account_id = ?", targetAccountID).
		Scan(ctx); err != nil {
		return nil, r.conn.ProcessError(err)
	}

	// create a new follow to 'replace' the request with
	follow := &gtsmodel.Follow{
		ID:              fr.ID,
		AccountID:       originAccountID,
		TargetAccountID: targetAccountID,
		URI:             fr.URI,
	}

	// if the follow already exists, just update the URI -- we don't need to do anything else
	if _, err := r.conn.
		NewInsert().
		Model(follow).
		On("CONFLICT ON CONSTRAINT follows_account_id_target_account_id_key DO UPDATE set uri = ?", follow.URI).
		Exec(ctx); err != nil {
		return nil, r.conn.ProcessError(err)
	}

	// now remove the follow request
	if _, err := r.conn.
		NewDelete().
		Model(&gtsmodel.FollowRequest{}).
		Where("account_id = ?", originAccountID).
		Where("target_account_id = ?", targetAccountID).
		Exec(ctx); err != nil {
		return nil, r.conn.ProcessError(err)
	}

	return follow, nil
}

func (r *relationshipDB) GetAccountFollowRequests(ctx context.Context, accountID string) ([]*gtsmodel.FollowRequest, db.Error) {
	followRequests := []*gtsmodel.FollowRequest{}

	q := r.newFollowQ(&followRequests).
		Where("target_account_id = ?", accountID)

	err := q.Scan(ctx)
	if err != nil {
		return nil, r.conn.ProcessError(err)
	}
	return followRequests, nil
}

func (r *relationshipDB) GetAccountFollows(ctx context.Context, accountID string) ([]*gtsmodel.Follow, db.Error) {
	follows := []*gtsmodel.Follow{}

	q := r.newFollowQ(&follows).
		Where("account_id = ?", accountID)

	err := q.Scan(ctx)
	if err != nil {
		return nil, r.conn.ProcessError(err)
	}
	return follows, nil
}

func (r *relationshipDB) CountAccountFollows(ctx context.Context, accountID string, localOnly bool) (int, db.Error) {
	return r.conn.
		NewSelect().
		Model(&[]*gtsmodel.Follow{}).
		Where("account_id = ?", accountID).
		Count(ctx)
}

func (r *relationshipDB) GetAccountFollowedBy(ctx context.Context, accountID string, localOnly bool) ([]*gtsmodel.Follow, db.Error) {
	follows := []*gtsmodel.Follow{}

	q := r.conn.
		NewSelect().
		Model(&follows)

	if localOnly {
		q = q.ColumnExpr("follow.*").
			Join("JOIN accounts AS a ON follow.account_id = TEXT(a.id)").
			Where("follow.target_account_id = ?", accountID).
			WhereGroup(" AND ", whereEmptyOrNull("a.domain"))
	} else {
		q = q.Where("target_account_id = ?", accountID)
	}

	err := q.Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return nil, r.conn.ProcessError(err)
	}
	return follows, nil
}

func (r *relationshipDB) CountAccountFollowedBy(ctx context.Context, accountID string, localOnly bool) (int, db.Error) {
	return r.conn.
		NewSelect().
		Model(&[]*gtsmodel.Follow{}).
		Where("target_account_id = ?", accountID).
		Count(ctx)
}
