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
	"golang.org/x/exp/slices"
)

type relationshipDB struct {
	db    *WrappedDB
	state *state.State
}

func (r *relationshipDB) GetRelationship(ctx context.Context, requestingAccount string, targetAccount string) (*gtsmodel.Relationship, error) {
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

	// retrieve a note by the requesting account on the target account, if there is one
	note, err := r.GetNote(
		gtscontext.SetBarebones(ctx),
		requestingAccount,
		targetAccount,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, fmt.Errorf("GetRelationship: error fetching note: %w", err)
	}
	if note != nil {
		rel.Note = note.Comment
	}

	return &rel, nil
}

func (r *relationshipDB) GetAccountFollows(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error) {
	followIDs, err := r.getAccountFollowIDs(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return r.GetFollowsByIDs(ctx, followIDs)
}

func (r *relationshipDB) GetAccountLocalFollows(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error) {
	followIDs, err := r.getAccountLocalFollowIDs(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return r.GetFollowsByIDs(ctx, followIDs)
}

func (r *relationshipDB) GetAccountFollowers(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error) {
	followerIDs, err := r.getAccountFollowerIDs(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return r.GetFollowsByIDs(ctx, followerIDs)
}

func (r *relationshipDB) GetAccountLocalFollowers(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error) {
	followerIDs, err := r.getAccountLocalFollowerIDs(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return r.GetFollowsByIDs(ctx, followerIDs)
}

func (r *relationshipDB) CountAccountFollows(ctx context.Context, accountID string) (int, error) {
	followIDs, err := r.getAccountFollowIDs(ctx, accountID)
	return len(followIDs), r.conn.ProcessError(err)
}

func (r *relationshipDB) CountAccountLocalFollows(ctx context.Context, accountID string) (int, error) {
	followIDs, err := r.getAccountLocalFollowIDs(ctx, accountID)
	return len(followIDs), r.conn.ProcessError(err)
}

func (r *relationshipDB) CountAccountFollowers(ctx context.Context, accountID string) (int, error) {
	followerIDs, err := r.getAccountFollowerIDs(ctx, accountID)
	return len(followerIDs), r.conn.ProcessError(err)
}

func (r *relationshipDB) CountAccountLocalFollowers(ctx context.Context, accountID string) (int, error) {
	followerIDs, err := r.getAccountLocalFollowerIDs(ctx, accountID)
	return len(followerIDs), r.conn.ProcessError(err)
}

func (r *relationshipDB) GetAccountFollowRequests(ctx context.Context, accountID string) ([]*gtsmodel.FollowRequest, error) {
	var followReqIDs []string
	if err := newSelectFollowRequests(r.db, accountID).
		Scan(ctx, &followReqIDs); err != nil {
		return nil, r.db.ProcessError(err)
	}
	return r.GetFollowRequestsByIDs(ctx, followReqIDs)
}

func (r *relationshipDB) GetAccountFollowRequesting(ctx context.Context, accountID string) ([]*gtsmodel.FollowRequest, error) {
	var followReqIDs []string
	if err := newSelectFollowRequesting(r.db, accountID).
		Scan(ctx, &followReqIDs); err != nil {
		return nil, r.db.ProcessError(err)
	}
	return r.GetFollowRequestsByIDs(ctx, followReqIDs)
}

func (r *relationshipDB) CountAccountFollowRequests(ctx context.Context, accountID string) (int, error) {
	n, err := newSelectFollowRequests(r.db, accountID).Count(ctx)
	return n, r.db.ProcessError(err)
}

func (r *relationshipDB) CountAccountFollowRequesting(ctx context.Context, accountID string) (int, error) {
	n, err := newSelectFollowRequesting(r.db, accountID).Count(ctx)
	return n, r.db.ProcessError(err)
}

func (r *relationshipDB) getAccountFollowIDs(ctx context.Context, accountID string) ([]string, error) {
	// Generate cache key.
	key := ">" + accountID

	// Look for follow IDs list in cache under this key.
	followIDs, ok := r.state.Caches.GTS.FollowIDs().Get(key)

	if !ok {
		// None found, perform database query.
		if _, err := r.conn.NewSelect().
			Table("follows").
			Column("id").
			Where("? = ?", bun.Ident("account_id"), accountID).
			OrderExpr("? DESC", bun.Ident("updated_at")).
			Exec(ctx, &followIDs); err != nil {
			return nil, r.conn.ProcessError(err)
		}

		// Store this result in the cache under key.
		r.state.Caches.GTS.FollowIDs().Set(key, followIDs)
	}

	// Return clone of result for safety.
	return slices.Clone(followIDs), nil
}

func (r *relationshipDB) getAccountLocalFollowIDs(ctx context.Context, accountID string) ([]string, error) {
	// Generate cache key.
	key := "l>" + accountID

	// Look for follow IDs list in cache under this key.
	followIDs, ok := r.state.Caches.GTS.FollowIDs().Get(key)

	if !ok {
		// None found, perform database query.
		if _, err := r.conn.NewSelect().
			Table("follows").
			Column("id").
			Where("? = ? AND ? IN (?)",
				bun.Ident("account_id"),
				accountID,
				bun.Ident("target_account_id"),
				r.conn.NewSelect().
					Table("accounts").
					Column("id").
					Where("? IS NULL", bun.Ident("domain")),
			).
			OrderExpr("? DESC", bun.Ident("updated_at")).
			Exec(ctx, &followIDs); err != nil {
			return nil, r.conn.ProcessError(err)
		}

		// Store this result in the cache under key.
		r.state.Caches.GTS.FollowIDs().Set(key, followIDs)
	}

	// Return clone of result for safety.
	return slices.Clone(followIDs), nil
}

func (r *relationshipDB) getAccountFollowerIDs(ctx context.Context, accountID string) ([]string, error) {
	// Generate cache key.
	key := "<" + accountID

	// Look for follow IDs list in cache under this key.
	followerIDs, ok := r.state.Caches.GTS.FollowIDs().Get(key)

	if !ok {
		// None found, perform database query.
		if _, err := r.conn.NewSelect().
			Table("follows").
			Column("id").
			Where("? = ?", bun.Ident("target_account_id"), accountID).
			OrderExpr("? DESC", bun.Ident("updated_at")).
			Exec(ctx, &followerIDs); err != nil {
			return nil, r.conn.ProcessError(err)
		}

		// Store this result in the cache under key.
		r.state.Caches.GTS.FollowIDs().Set(key, followerIDs)
	}

	// Return clone of result for safety.
	return slices.Clone(followerIDs), nil
}

func (r *relationshipDB) getAccountLocalFollowerIDs(ctx context.Context, accountID string) ([]string, error) {
	// Generate cache key.
	key := "l<" + accountID

	// Look for follow IDs list in cache under this key.
	followerIDs, ok := r.state.Caches.GTS.FollowIDs().Get(key)

	if !ok {
		// None found, perform database query.
		if _, err := r.conn.NewSelect().
			Table("follows").
			Column("id").
			Where("? = ? AND ? IN (?)",
				bun.Ident("target_account_id"),
				accountID,
				bun.Ident("account_id"),
				r.conn.NewSelect().
					Table("accounts").
					Column("id").
					Where("? IS NULL", bun.Ident("domain")),
			).
			OrderExpr("? DESC", bun.Ident("updated_at")).
			Exec(ctx, &followerIDs); err != nil {
			return nil, r.conn.ProcessError(err)
		}

		// Store this result in the cache under key.
		r.state.Caches.GTS.FollowIDs().Set(key, followerIDs)
	}

	// Return clone of result for safety.
	return slices.Clone(followerIDs), nil
}

// newSelectFollowRequests returns a new select query for all rows in the follow_requests table with target_account_id = accountID.
func newSelectFollowRequests(db *WrappedDB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		TableExpr("?", bun.Ident("follow_requests")).
		ColumnExpr("?", bun.Ident("id")).
		Where("? = ?", bun.Ident("target_account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("updated_at"))
}

// newSelectFollowRequesting returns a new select query for all rows in the follow_requests table with account_id = accountID.
func newSelectFollowRequesting(db *WrappedDB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		TableExpr("?", bun.Ident("follow_requests")).
		ColumnExpr("?", bun.Ident("id")).
		Where("? = ?", bun.Ident("target_account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("updated_at"))
}
