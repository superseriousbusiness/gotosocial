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

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type relationshipDB struct {
	conn  *DBConn
	state *state.State
}

// func (r *relationshipDB) IsBlocked(ctx context.Context, account1 string, account2 string, eitherDirection bool) (bool, db.Error) {
// 	// Look for a block in direction of account1->account2
// 	block1, err := r.GetBlock(
// 		gtscontext.SetBarebones(ctx),
// 		account1,
// 		account2,
// 	)
// 	if err != nil && !errors.Is(err, db.ErrNoEntries) {
// 		return false, err
// 	}

// 	if block1 != nil {
// 		// account1 blocks account2
// 		return true, nil
// 	} else if !eitherDirection {
// 		// Don't check for mutli-directional
// 		return false, nil
// 	}

// 	// Look for a block in direction of account2->account1
// 	block2, err := r.GetBlock(
// 		gtscontext.SetBarebones(ctx),
// 		account2,
// 		account1,
// 	)
// 	if err != nil && !errors.Is(err, db.ErrNoEntries) {
// 		return false, err
// 	}

// 	return (block2 != nil), nil
// }

// func (r *relationshipDB) GetBlock(ctx context.Context, account1 string, account2 string) (*gtsmodel.Block, db.Error) {
// 	return r.getBlock(
// 		ctx,
// 		"AccountID.TargetAccountID",
// 		func(block *gtsmodel.Block) error {
// 			return r.conn.NewSelect().Model(block).
// 				Where("? = ?", bun.Ident("block.account_id"), account1).
// 				Where("? = ?", bun.Ident("block.target_account_id"), account2).
// 				Scan(ctx)
// 		},
// 		account1, account2,
// 	)
// }

// func (r *relationshipDB) GetBlockByID(ctx context.Context, id string) (*gtsmodel.Block, error) {
// 	return r.getBlock(
// 		ctx,
// 		"ID",
// 		func(block *gtsmodel.Block) error {
// 			return r.conn.NewSelect().Model(block).
// 				Where("? = ?", bun.Ident("block.id"), id).
// 				Scan(ctx)
// 		},
// 		id,
// 	)
// }

// func (r *relationshipDB) GetBlockByURI(ctx context.Context, uri string) (*gtsmodel.Block, error) {
// 	return r.getBlock(
// 		ctx,
// 		"URI",
// 		func(block *gtsmodel.Block) error {
// 			return r.conn.NewSelect().Model(block).
// 				Where("? = ?", bun.Ident("block.uri"), uri).
// 				Scan(ctx)
// 		},
// 		uri,
// 	)
// }

// func (r *relationshipDB) getBlock(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Block) error, keyParts ...any) (*gtsmodel.Block, db.Error) {
// 	// Fetch block from database cache with loader callback
// 	block, err := r.state.Caches.GTS.Block().Load(lookup, func() (*gtsmodel.Block, error) {
// 		var block gtsmodel.Block

// 		if err := dbQuery(&block); err != nil {
// 			return nil, r.conn.ProcessError(err)
// 		}

// 		return &block, nil
// 	}, keyParts...)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if gtscontext.Barebones(ctx) {
// 		// no need to fully populate.
// 		return block, nil
// 	}

// 	// Set the block originating account
// 	block.Account, err = r.state.DB.GetAccountByID(
// 		gtscontext.SetBarebones(ctx),
// 		block.AccountID,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("error populating block origin account: %w", err)
// 	}

// 	// Set the block target account
// 	block.TargetAccount, err = r.state.DB.GetAccountByID(
// 		gtscontext.SetBarebones(ctx),
// 		block.TargetAccountID,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("error populating block target account: %w", err)
// 	}

// 	return block, nil
// }

// func (r *relationshipDB) PutBlock(ctx context.Context, block *gtsmodel.Block) db.Error {
// 	// Store the block in the database and set in cache.
// 	err := r.state.Caches.GTS.Block().Store(block, func() error {
// 		_, err := r.conn.NewInsert().Model(block).Exec(ctx)
// 		return r.conn.ProcessError(err)
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	// Invalid accounts from all visibility lookups.
// 	r.state.Caches.Visibility.Invalidate("ItemID", block.AccountID)
// 	r.state.Caches.Visibility.Invalidate("ItemID", block.TargetAccountID)
// 	r.state.Caches.Visibility.Invalidate("RequesterID", block.AccountID)
// 	r.state.Caches.Visibility.Invalidate("RequesterID", block.TargetAccountID)

// 	return nil
// }

// func (r *relationshipDB) DeleteBlockByID(ctx context.Context, id string) db.Error {
// 	// First, look for block with ID.
// 	block, err := r.GetBlockByID(
// 		gtscontext.SetBarebones(ctx), id,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	// Delete from database.
// 	if _, err := r.conn.
// 		NewDelete().
// 		TableExpr("? AS ?", bun.Ident("blocks"), bun.Ident("block")).
// 		Where("? = ?", bun.Ident("block.id"), id).
// 		Exec(ctx); err != nil {
// 		return r.conn.ProcessError(err)
// 	}

// 	// Invalid block from database lookups.
// 	r.state.Caches.GTS.Block().Invalidate("ID", id)

// 	// Invalid accounts from all visibility lookups.
// 	r.state.Caches.Visibility.Invalidate("ItemID", block.AccountID)
// 	r.state.Caches.Visibility.Invalidate("ItemID", block.TargetAccountID)
// 	r.state.Caches.Visibility.Invalidate("RequesterID", block.AccountID)
// 	r.state.Caches.Visibility.Invalidate("RequesterID", block.TargetAccountID)

// 	return nil
// }

// func (r *relationshipDB) DeleteBlockByURI(ctx context.Context, uri string) db.Error {
// 	// First, look for block with URI.
// 	block, err := r.GetBlockByID(
// 		gtscontext.SetBarebones(ctx), uri,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	// Delete from database.
// 	if _, err := r.conn.
// 		NewDelete().
// 		TableExpr("? AS ?", bun.Ident("blocks"), bun.Ident("block")).
// 		Where("? = ?", bun.Ident("block.uri"), uri).
// 		Exec(ctx); err != nil {
// 		return r.conn.ProcessError(err)
// 	}

// 	// Invalid block from database lookups.
// 	r.state.Caches.GTS.Block().Invalidate("URI", uri)

// 	// Invalid accounts from all visibility lookups.
// 	r.state.Caches.Visibility.Invalidate("ItemID", block.AccountID)
// 	r.state.Caches.Visibility.Invalidate("ItemID", block.TargetAccountID)
// 	r.state.Caches.Visibility.Invalidate("RequesterID", block.AccountID)
// 	r.state.Caches.Visibility.Invalidate("RequesterID", block.TargetAccountID)

// 	return nil
// }

// func (r *relationshipDB) DeleteBlocksByOriginAccountID(ctx context.Context, originAccountID string) db.Error {
// 	blockIDs := []string{}

// 	q := r.conn.
// 		NewSelect().
// 		TableExpr("? AS ?", bun.Ident("blocks"), bun.Ident("block")).
// 		Column("block.id").
// 		Where("? = ?", bun.Ident("block.account_id"), originAccountID)

// 	if err := q.Scan(ctx, &blockIDs); err != nil {
// 		return r.conn.ProcessError(err)
// 	}

// 	for _, blockID := range blockIDs {
// 		if err := r.DeleteBlockByID(ctx, blockID); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (r *relationshipDB) DeleteBlocksByTargetAccountID(ctx context.Context, targetAccountID string) db.Error {
// 	blockIDs := []string{}

// 	q := r.conn.
// 		NewSelect().
// 		TableExpr("? AS ?", bun.Ident("blocks"), bun.Ident("block")).
// 		Column("block.id").
// 		Where("? = ?", bun.Ident("block.target_account_id"), targetAccountID)

// 	if err := q.Scan(ctx, &blockIDs); err != nil {
// 		return r.conn.ProcessError(err)
// 	}

// 	for _, blockID := range blockIDs {
// 		if err := r.DeleteBlockByID(ctx, blockID); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// 	// make sure account 2 follows account 1
// 	f2, err := r.IsFollowing(ctx, account2, account1)
// 	if err != nil {
// 		return false, err
// 	}

// 	return f1 && f2, nil
// }

// func (r *relationshipDB) AcceptFollowRequest(ctx context.Context, originAccountID string, targetAccountID string) (*gtsmodel.Follow, db.Error) {
// 	// Get original follow request.
// 	var followRequestID string
// 	if err := r.conn.
// 		NewSelect().
// 		TableExpr("? AS ?", bun.Ident("follow_requests"), bun.Ident("follow_request")).
// 		Column("follow_request.id").
// 		Where("? = ?", bun.Ident("follow_request.account_id"), originAccountID).
// 		Where("? = ?", bun.Ident("follow_request.target_account_id"), targetAccountID).
// 		Scan(ctx, &followRequestID); err != nil {
// 		return nil, r.conn.ProcessError(err)
// 	}

// 	followRequest, err := r.getFollowRequest(ctx, followRequestID)
// 	if err != nil {
// 		return nil, r.conn.ProcessError(err)
// 	}

// 	// Create a new follow to 'replace'
// 	// the original follow request with.
// 	follow := &gtsmodel.Follow{
// 		ID:              followRequest.ID,
// 		AccountID:       originAccountID,
// 		Account:         followRequest.Account,
// 		TargetAccountID: targetAccountID,
// 		TargetAccount:   followRequest.TargetAccount,
// 		URI:             followRequest.URI,
// 	}

// 	// If the follow already exists, just
// 	// replace the URI with the new one.
// 	if _, err := r.conn.
// 		NewInsert().
// 		Model(follow).
// 		On("CONFLICT (?,?) DO UPDATE set ? = ?", bun.Ident("account_id"), bun.Ident("target_account_id"), bun.Ident("uri"), follow.URI).
// 		Exec(ctx); err != nil {
// 		return nil, r.conn.ProcessError(err)
// 	}

// 	// Delete original follow request.
// 	if _, err := r.conn.
// 		NewDelete().
// 		TableExpr("? AS ?", bun.Ident("follow_requests"), bun.Ident("follow_request")).
// 		Where("? = ?", bun.Ident("follow_request.id"), followRequest.ID).
// 		Exec(ctx); err != nil {
// 		return nil, r.conn.ProcessError(err)
// 	}

// 	// Delete original follow request notification.
// 	if err := r.deleteFollowRequestNotif(ctx, originAccountID, targetAccountID); err != nil {
// 		return nil, err
// 	}

// 	// Invalid accounts from all visibility lookups.
// 	r.state.Caches.Visibility.Invalidate("ItemID", originAccountID)
// 	r.state.Caches.Visibility.Invalidate("ItemID", targetAccountID)
// 	r.state.Caches.Visibility.Invalidate("RequesterID", originAccountID)
// 	r.state.Caches.Visibility.Invalidate("RequesterID", targetAccountID)

// 	// return the new follow
// 	return follow, nil
// }

// func (r *relationshipDB) RejectFollowRequest(ctx context.Context, originAccountID string, targetAccountID string) (*gtsmodel.FollowRequest, db.Error) {
// 	// Get original follow request.
// 	var followRequestID string
// 	if err := r.conn.
// 		NewSelect().
// 		TableExpr("? AS ?", bun.Ident("follow_requests"), bun.Ident("follow_request")).
// 		Column("follow_request.id").
// 		Where("? = ?", bun.Ident("follow_request.account_id"), originAccountID).
// 		Where("? = ?", bun.Ident("follow_request.target_account_id"), targetAccountID).
// 		Scan(ctx, &followRequestID); err != nil {
// 		return nil, r.conn.ProcessError(err)
// 	}

// 	followRequest, err := r.getFollowRequest(ctx, followRequestID)
// 	if err != nil {
// 		return nil, r.conn.ProcessError(err)
// 	}

// 	// Delete original follow request.
// 	if _, err := r.conn.
// 		NewDelete().
// 		TableExpr("? AS ?", bun.Ident("follow_requests"), bun.Ident("follow_request")).
// 		Where("? = ?", bun.Ident("follow_request.id"), followRequest.ID).
// 		Exec(ctx); err != nil {
// 		return nil, r.conn.ProcessError(err)
// 	}

// 	// Delete original follow request notification.
// 	if err := r.deleteFollowRequestNotif(ctx, originAccountID, targetAccountID); err != nil {
// 		return nil, err
// 	}

// 	// Invalid accounts from all visibility lookups.
// 	r.state.Caches.Visibility.Invalidate("ItemID", originAccountID)
// 	r.state.Caches.Visibility.Invalidate("ItemID", targetAccountID)
// 	r.state.Caches.Visibility.Invalidate("RequesterID", originAccountID)
// 	r.state.Caches.Visibility.Invalidate("RequesterID", targetAccountID)

// 	// return the deleted follow request
// 	return followRequest, nil
// }

// func (r *relationshipDB) deleteFollowRequestNotif(ctx context.Context, originAccountID string, targetAccountID string) db.Error {
// 	var id string
// 	if err := r.conn.
// 		NewSelect().
// 		TableExpr("? AS ?", bun.Ident("notifications"), bun.Ident("notification")).
// 		Column("notification.id").
// 		Where("? = ?", bun.Ident("notification.origin_account_id"), originAccountID).
// 		Where("? = ?", bun.Ident("notification.target_account_id"), targetAccountID).
// 		Where("? = ?", bun.Ident("notification.notification_type"), gtsmodel.NotificationFollowRequest).
// 		Limit(1). // There should only be one!
// 		Scan(ctx, &id); err != nil {
// 		err = r.conn.ProcessError(err)
// 		if errors.Is(err, db.ErrNoEntries) {
// 			// If no entries, the notif didn't
// 			// exist anyway so nothing to do here.
// 			return nil
// 		}
// 		// Return on real error.
// 		return err
// 	}

// 	return r.state.DB.DeleteNotification(ctx, id)
// }

// func (r *relationshipDB) getFollow(ctx context.Context, id string) (*gtsmodel.Follow, db.Error) {
// 	follow := &gtsmodel.Follow{}

// 	err := r.conn.
// 		NewSelect().
// 		Model(follow).
// 		Where("? = ?", bun.Ident("follow.id"), id).
// 		Scan(ctx)
// 	if err != nil {
// 		return nil, r.conn.ProcessError(err)
// 	}

// 	follow.Account, err = r.state.DB.GetAccountByID(ctx, follow.AccountID)
// 	if err != nil {
// 		log.Errorf(ctx, "error getting follow account %q: %v", follow.AccountID, err)
// 	}

// 	follow.TargetAccount, err = r.state.DB.GetAccountByID(ctx, follow.TargetAccountID)
// 	if err != nil {
// 		log.Errorf(ctx, "error getting follow target account %q: %v", follow.TargetAccountID, err)
// 	}

// 	return follow, nil
// }

// func (r *relationshipDB) GetLocalFollowersIDs(ctx context.Context, targetAccountID string) ([]string, db.Error) {
// 	accountIDs := []string{}

// 	// Select only the account ID of each follow.
// 	q := r.conn.
// 		NewSelect().
// 		TableExpr("? AS ?", bun.Ident("follows"), bun.Ident("follow")).
// 		ColumnExpr("? AS ?", bun.Ident("follow.account_id"), bun.Ident("account_id")).
// 		Where("? = ?", bun.Ident("follow.target_account_id"), targetAccountID)

// 	// Join on accounts table to select only
// 	// those with NULL domain (local accounts).
// 	q = q.
// 		Join("JOIN ? AS ? ON ? = ?",
// 			bun.Ident("accounts"),
// 			bun.Ident("account"),
// 			bun.Ident("follow.account_id"),
// 			bun.Ident("account.id"),
// 		).
// 		Where("? IS NULL", bun.Ident("account.domain"))

// 	// We don't *really* need to order these,
// 	// but it makes it more consistent to do so.
// 	q = q.Order("account_id DESC")

// 	if err := q.Scan(ctx, &accountIDs); err != nil {
// 		return nil, r.conn.ProcessError(err)
// 	}

// 	return accountIDs, nil
// }

// func (r *relationshipDB) GetFollows(ctx context.Context, accountID string, targetAccountID string) ([]*gtsmodel.Follow, db.Error) {
// 	ids := []string{}

// 	q := r.conn.
// 		NewSelect().
// 		TableExpr("? AS ?", bun.Ident("follows"), bun.Ident("follow")).
// 		Column("follow.id").
// 		Order("follow.updated_at DESC")

// 	if accountID != "" {
// 		q = q.Where("? = ?", bun.Ident("follow.account_id"), accountID)
// 	}

// 	if targetAccountID != "" {
// 		q = q.Where("? = ?", bun.Ident("follow.target_account_id"), targetAccountID)
// 	}

// 	if err := q.Scan(ctx, &ids); err != nil {
// 		return nil, r.conn.ProcessError(err)
// 	}

// 	follows := make([]*gtsmodel.Follow, 0, len(ids))
// 	for _, id := range ids {
// 		follow, err := r.getFollow(ctx, id)
// 		if err != nil {
// 			log.Errorf(ctx, "error getting follow %q: %v", id, err)
// 			continue
// 		}

// 		follows = append(follows, follow)
// 	}

// 	return follows, nil
// }

// func (r *relationshipDB) CountFollows(ctx context.Context, accountID string, targetAccountID string) (int, db.Error) {
// 	q := r.conn.
// 		NewSelect().
// 		TableExpr("? AS ?", bun.Ident("follows"), bun.Ident("follow")).
// 		Column("follow.id")

// 	if accountID != "" {
// 		q = q.Where("? = ?", bun.Ident("follow.account_id"), accountID)
// 	}

// 	if targetAccountID != "" {
// 		q = q.Where("? = ?", bun.Ident("follow.target_account_id"), targetAccountID)
// 	}

// 	return q.Count(ctx)
// }

// func (r *relationshipDB) getFollowRequest(ctx context.Context, id string) (*gtsmodel.FollowRequest, db.Error) {
// 	followRequest := &gtsmodel.FollowRequest{}

// 	err := r.conn.
// 		NewSelect().
// 		Model(followRequest).
// 		Where("? = ?", bun.Ident("follow_request.id"), id).
// 		Scan(ctx)
// 	if err != nil {
// 		return nil, r.conn.ProcessError(err)
// 	}

// 	followRequest.Account, err = r.state.DB.GetAccountByID(ctx, followRequest.AccountID)
// 	if err != nil {
// 		log.Errorf(ctx, "error getting follow request account %q: %v", followRequest.AccountID, err)
// 	}

// 	followRequest.TargetAccount, err = r.state.DB.GetAccountByID(ctx, followRequest.TargetAccountID)
// 	if err != nil {
// 		log.Errorf(ctx, "error getting follow request target account %q: %v", followRequest.TargetAccountID, err)
// 	}

// 	return followRequest, nil
// }

// func (r *relationshipDB) GetFollowRequests(ctx context.Context, accountID string, targetAccountID string) ([]*gtsmodel.FollowRequest, db.Error) {
// 	ids := []string{}

// 	q := r.conn.
// 		NewSelect().
// 		TableExpr("? AS ?", bun.Ident("follow_requests"), bun.Ident("follow_request")).
// 		Column("follow_request.id")

// 	if accountID != "" {
// 		q = q.Where("? = ?", bun.Ident("follow_request.account_id"), accountID)
// 	}

// 	if targetAccountID != "" {
// 		q = q.Where("? = ?", bun.Ident("follow_request.target_account_id"), targetAccountID)
// 	}

// 	if err := q.Scan(ctx, &ids); err != nil {
// 		return nil, r.conn.ProcessError(err)
// 	}

// 	followRequests := make([]*gtsmodel.FollowRequest, 0, len(ids))
// 	for _, id := range ids {
// 		followRequest, err := r.getFollowRequest(ctx, id)
// 		if err != nil {
// 			log.Errorf(ctx, "error getting follow request %q: %v", id, err)
// 			continue
// 		}

// 		followRequests = append(followRequests, followRequest)
// 	}

// 	return followRequests, nil
// }

// func (r *relationshipDB) CountFollowRequests(ctx context.Context, accountID string, targetAccountID string) (int, db.Error) {
// 	q := r.conn.
// 		NewSelect().
// 		TableExpr("? AS ?", bun.Ident("follow_requests"), bun.Ident("follow_request")).
// 		Column("follow_request.id").
// 		Order("follow_request.updated_at DESC")

// 	if accountID != "" {
// 		q = q.Where("? = ?", bun.Ident("follow_request.account_id"), accountID)
// 	}

// 	if targetAccountID != "" {
// 		q = q.Where("? = ?", bun.Ident("follow_request.target_account_id"), targetAccountID)
// 	}

// 	return q.Count(ctx)
// }

// func (r *relationshipDB) Unfollow(ctx context.Context, originAccountID string, targetAccountID string) (string, db.Error) {
// 	uri := new(string)

// 	_, err := r.conn.
// 		NewDelete().
// 		TableExpr("? AS ?", bun.Ident("follows"), bun.Ident("follow")).
// 		Where("? = ?", bun.Ident("follow.target_account_id"), targetAccountID).
// 		Where("? = ?", bun.Ident("follow.account_id"), originAccountID).
// 		Returning("?", bun.Ident("uri")).Exec(ctx, uri)

// 	// Only return proper errors.
// 	if err = r.conn.ProcessError(err); err != db.ErrNoEntries {
// 		return *uri, err
// 	}

// 	return *uri, nil
// }

// func (r *relationshipDB) UnfollowRequest(ctx context.Context, originAccountID string, targetAccountID string) (string, db.Error) {
// 	uri := new(string)

// 	_, err := r.conn.
// 		NewDelete().
// 		TableExpr("? AS ?", bun.Ident("follow_requests"), bun.Ident("follow_request")).
// 		Where("? = ?", bun.Ident("follow_request.target_account_id"), targetAccountID).
// 		Where("? = ?", bun.Ident("follow_request.account_id"), originAccountID).
// 		Returning("?", bun.Ident("uri")).Exec(ctx, uri)

// 	// Only return proper errors.
// 	if err = r.conn.ProcessError(err); err != db.ErrNoEntries {
// 		return *uri, err
// 	}

// 	return *uri, nil
// }

func (r *relationshipDB) GetRelationship(ctx context.Context, requestingAccount string, targetAccount string) (*gtsmodel.Relationship, db.Error) {
	var rel gtsmodel.Relationship
	rel.ID = targetAccount

	// check if the requesting follows the target
	follow, err := r.GetFollow(
		gtscontext.SetBarebones(ctx),
		requestingAccount,
		targetAccount,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, fmt.Errorf("GetRelationship: error fetching follow: %w", err)
	}

	if follow != nil {
		// follow exists so we can fill these fields out...
		rel.Following = true
		rel.ShowingReblogs = *follow.ShowReblogs
		rel.Notifying = *follow.Notify
	}

	// check if the target follows the requesting
	rel.FollowedBy, err = r.IsFollowing(ctx,
		targetAccount,
		requestingAccount,
	)
	if err != nil {
		return nil, fmt.Errorf("GetRelationship: error checking followedBy: %w", err)
	}

	// check if requesting has follow requested target
	rel.Requested, err = r.IsFollowRequested(ctx,
		requestingAccount,
		targetAccount,
	)
	if err != nil {
		return nil, fmt.Errorf("GetRelationship: error checking requested: %w", err)
	}

	// check if the requesting account is blocking the target account
	rel.Blocking, err = r.IsBlocked(ctx, requestingAccount, targetAccount)
	if err != nil {
		return nil, fmt.Errorf("GetRelationship: error checking blocking: %w", err)
	}

	// check if the requesting account is blocked by the target account
	rel.BlockedBy, err = r.IsBlocked(ctx, targetAccount, requestingAccount)
	if err != nil {
		return nil, fmt.Errorf("GetRelationship: error checking blockedBy: %w", err)
	}

	return &rel, nil
}

func (r *relationshipDB) GetAccountFollows(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error) {
	var followIDs []string
	if err := newSelectFollows(r.conn, accountID).
		Scan(ctx, &followIDs); err != nil {
		return nil, r.conn.ProcessError(err)
	}
	return r.GetFollowsByIDs(ctx, followIDs)
}

func (r *relationshipDB) GetAccountLocalFollows(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error) {
	var followIDs []string
	if err := newSelectLocalFollows(r.conn, accountID).
		Scan(ctx, &followIDs); err != nil {
		return nil, r.conn.ProcessError(err)
	}
	return r.GetFollowsByIDs(ctx, followIDs)
}

func (r *relationshipDB) GetAccountFollowers(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error) {
	var followIDs []string
	if err := newSelectFollowers(r.conn, accountID).
		Scan(ctx, &followIDs); err != nil {
		return nil, r.conn.ProcessError(err)
	}
	return r.GetFollowsByIDs(ctx, followIDs)
}

func (r *relationshipDB) GetAccountLocalFollowers(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error) {
	var followIDs []string
	if err := newSelectLocalFollowers(r.conn, accountID).
		Scan(ctx, &followIDs); err != nil {
		return nil, r.conn.ProcessError(err)
	}
	return r.GetFollowsByIDs(ctx, followIDs)
}

func (r *relationshipDB) CountAccountFollows(ctx context.Context, accountID string) (int, error) {
	n, err := newSelectFollows(r.conn, accountID).Count(ctx)
	return n, r.conn.ProcessError(err)
}

func (r *relationshipDB) CountAccountLocalFollows(ctx context.Context, accountID string) (int, error) {
	n, err := newSelectLocalFollows(r.conn, accountID).Count(ctx)
	return n, r.conn.ProcessError(err)
}

func (r *relationshipDB) CountAccountFollowers(ctx context.Context, accountID string) (int, error) {
	n, err := newSelectFollowers(r.conn, accountID).Count(ctx)
	return n, r.conn.ProcessError(err)
}

func (r *relationshipDB) CountAccountLocalFollowers(ctx context.Context, accountID string) (int, error) {
	n, err := newSelectLocalFollowers(r.conn, accountID).Count(ctx)
	return n, r.conn.ProcessError(err)
}

func (r *relationshipDB) GetAccountFollowRequests(ctx context.Context, accountID string) ([]*gtsmodel.FollowRequest, error) {
	var followReqIDs []string
	if err := newSelectFollowRequests(r.conn, accountID).
		Scan(ctx, &followReqIDs); err != nil {
		return nil, r.conn.ProcessError(err)
	}
	return r.GetFollowRequestsByIDs(ctx, followReqIDs)
}

func (r *relationshipDB) GetAccountFollowRequesting(ctx context.Context, accountID string) ([]*gtsmodel.FollowRequest, error) {
	var followReqIDs []string
	if err := newSelectFollowRequesting(r.conn, accountID).
		Scan(ctx, &followReqIDs); err != nil {
		return nil, r.conn.ProcessError(err)
	}
	return r.GetFollowRequestsByIDs(ctx, followReqIDs)
}

func (r *relationshipDB) CountAccountFollowRequests(ctx context.Context, accountID string) (int, error) {
	n, err := newSelectFollowRequests(r.conn, accountID).Count(ctx)
	return n, r.conn.ProcessError(err)
}

func (r *relationshipDB) CountAccountFollowRequesting(ctx context.Context, accountID string) (int, error) {
	n, err := newSelectFollowRequesting(r.conn, accountID).Count(ctx)
	return n, r.conn.ProcessError(err)
}

// newSelectFollowRequests returns a new select query for all rows in the follow_requests table with target_account_id = accountID.
func newSelectFollowRequests(conn *DBConn, accountID string) *bun.SelectQuery {
	return conn.NewSelect().
		TableExpr("?", bun.Ident("follow_requests")).
		ColumnExpr("?", bun.Ident("id")).
		Where("? = ?", bun.Ident("target_account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("updated_at"))
}

// newSelectFollowRequesting returns a new select query for all rows in the follow_requests table with account_id = accountID.
func newSelectFollowRequesting(conn *DBConn, accountID string) *bun.SelectQuery {
	return conn.NewSelect().
		TableExpr("?", bun.Ident("follow_requests")).
		ColumnExpr("?", bun.Ident("id")).
		Where("? = ?", bun.Ident("target_account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("updated_at"))
}

// newSelectFollows returns a new select query for all rows in the follows table with account_id = accountID.
func newSelectFollows(conn *DBConn, accountID string) *bun.SelectQuery {
	return conn.NewSelect().
		Table("follows").
		Column("id").
		Where("? = ?", bun.Ident("account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("updated_at"))
}

// newSelectLocalFollows returns a new select query for all rows in the follows table with
// account_id = accountID where the corresponding account ID has a NULL domain (i.e. is local).
func newSelectLocalFollows(conn *DBConn, accountID string) *bun.SelectQuery {
	return conn.NewSelect().
		Table("follows").
		Column("id").
		Where("? = ? AND ? IN ( SELECT ? FROM ? WHERE ? IS NULL )",
			bun.Ident("account_id"),
			accountID,
			bun.Ident("target_account_id"),
			bun.Ident("id"),
			bun.Ident("accounts"),
			bun.Ident("domain"),
		).
		OrderExpr("? DESC", bun.Ident("updated_at"))
}

// newSelectFollowers returns a new select query for all rows in the follows table with target_account_id = accountID.
func newSelectFollowers(conn *DBConn, accountID string) *bun.SelectQuery {
	return conn.NewSelect().
		Table("follows").
		Column("id").
		Where("? = ?", bun.Ident("target_account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("updated_at"))
}

// newSelectLocalFollowers returns a new select query for all rows in the follows table with
// target_account_id = accountID where the corresponding account ID has a NULL domain (i.e. is local).
func newSelectLocalFollowers(conn *DBConn, accountID string) *bun.SelectQuery {
	return conn.NewSelect().
		Table("follows").
		Column("id").
		Where("? = ? AND ? IN ( SELECT ? FROM ? WHERE ? IS NULL )",
			bun.Ident("target_account_id"),
			accountID,
			bun.Ident("account_id"),
			bun.Ident("id"),
			bun.Ident("accounts"),
			bun.Ident("domain"),
		).
		OrderExpr("? DESC", bun.Ident("updated_at"))
}
