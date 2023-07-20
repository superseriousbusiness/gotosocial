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
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
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

func (r *relationshipDB) GetAccountFollowRequests(ctx context.Context, accountID string) ([]*gtsmodel.FollowRequest, error) {
	followReqIDs, err := r.getAccountFollowRequestIDs(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return r.GetFollowRequestsByIDs(ctx, followReqIDs)
}

func (r *relationshipDB) GetAccountFollowRequesting(ctx context.Context, accountID string) ([]*gtsmodel.FollowRequest, error) {
	followReqIDs, err := r.getAccountFollowRequestingIDs(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return r.GetFollowRequestsByIDs(ctx, followReqIDs)
}

func (r *relationshipDB) GetAccountBlocks(ctx context.Context, accountID string, page *paging.Pager) ([]*gtsmodel.Block, error) {
	// Load block IDs from cache with database loader callback.
	blockIDs, err := r.state.Caches.GTS.BlockIDs().LoadRange(accountID, func() ([]string, error) {
		var blockIDs []string

		// Block IDs not in cache, perform DB query!
		q := newSelectBlocks(r.conn, accountID)
		if _, err := q.Exec(ctx, &blockIDs); err != nil {
			return nil, r.conn.ProcessError(err)
		}

		return blockIDs, nil
	}, func(blockIDs []string) []string {
		// Filter blockIDs to given paging.
		return page.PageDesc(blockIDs)
	})
	if err != nil {
		return nil, err
	}

	// Convert these IDs to full block objects.
	return r.GetBlocksByIDs(ctx, blockIDs)
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

func (r *relationshipDB) CountAccountFollowRequests(ctx context.Context, accountID string) (int, error) {
	n, err := newSelectFollowRequests(r.db, accountID).Count(ctx)
	return n, r.db.ProcessError(err)
}

func (r *relationshipDB) CountAccountFollowRequesting(ctx context.Context, accountID string) (int, error) {
	n, err := newSelectFollowRequesting(r.db, accountID).Count(ctx)
	return n, r.db.ProcessError(err)
}

func (r *relationshipDB) getAccountFollowIDs(ctx context.Context, accountID string) ([]string, error) {
	return r.state.Caches.GTS.FollowIDs().Load(">"+accountID, func() ([]string, error) {
		var followIDs []string

		// Follow IDs not in cache, perform DB query!
		q := newSelectFollows(r.conn, accountID)
		if _, err := q.Exec(ctx, &followIDs); err != nil {
			return nil, r.conn.ProcessError(err)
		}

		return followIDs, nil
	})
}

func (r *relationshipDB) getAccountLocalFollowIDs(ctx context.Context, accountID string) ([]string, error) {
	return r.state.Caches.GTS.FollowIDs().Load("l>"+accountID, func() ([]string, error) {
		var followIDs []string

		// Follow IDs not in cache, perform DB query!
		q := newSelectLocalFollows(r.conn, accountID)
		if _, err := q.Exec(ctx, &followIDs); err != nil {
			return nil, r.conn.ProcessError(err)
		}

		return followIDs, nil
	})
}

func (r *relationshipDB) getAccountFollowerIDs(ctx context.Context, accountID string) ([]string, error) {
	return r.state.Caches.GTS.FollowIDs().Load("<"+accountID, func() ([]string, error) {
		var followIDs []string

		// Follow IDs not in cache, perform DB query!
		q := newSelectFollowers(r.conn, accountID)
		if _, err := q.Exec(ctx, &followIDs); err != nil {
			return nil, r.conn.ProcessError(err)
		}

		return followIDs, nil
	})
}

func (r *relationshipDB) getAccountLocalFollowerIDs(ctx context.Context, accountID string) ([]string, error) {
	return r.state.Caches.GTS.FollowIDs().Load("<"+accountID, func() ([]string, error) {
		var followIDs []string

		// Follow IDs not in cache, perform DB query!
		q := newSelectLocalFollowers(r.conn, accountID)
		if _, err := q.Exec(ctx, &followIDs); err != nil {
			return nil, r.conn.ProcessError(err)
		}

		return followIDs, nil
	})
}

func (r *relationshipDB) getAccountFollowRequestIDs(ctx context.Context, accountID string) ([]string, error) {
	return r.state.Caches.GTS.FollowRequestIDs().Load(">"+accountID, func() ([]string, error) {
		var followReqIDs []string

		// Follow request IDs not in cache, perform DB query!
		q := newSelectFollowRequests(r.conn, accountID)
		if _, err := q.Exec(ctx, &followReqIDs); err != nil {
			return nil, r.conn.ProcessError(err)
		}

		return followReqIDs, nil
	})
}

func (r *relationshipDB) getAccountFollowRequestingIDs(ctx context.Context, accountID string) ([]string, error) {
	return r.state.Caches.GTS.FollowRequestIDs().Load("<"+accountID, func() ([]string, error) {
		var followReqIDs []string

		// Follow request IDs not in cache, perform DB query!
		q := newSelectFollowRequesting(r.conn, accountID)
		if _, err := q.Exec(ctx, &followReqIDs); err != nil {
			return nil, r.conn.ProcessError(err)
		}

		return followReqIDs, nil
	})
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
		Where("? = ? AND ? IN (?)",
			bun.Ident("account_id"),
			accountID,
			bun.Ident("target_account_id"),
			conn.NewSelect().
				Table("accounts").
				Column("id").
				Where("? IS NULL", bun.Ident("domain")),
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
		Where("? = ? AND ? IN (?)",
			bun.Ident("target_account_id"),
			accountID,
			bun.Ident("account_id"),
			conn.NewSelect().
				Table("accounts").
				Column("id").
				Where("? IS NULL", bun.Ident("domain")),
		).
		OrderExpr("? DESC", bun.Ident("updated_at"))
}

// newSelectBlocks ...
func newSelectBlocks(conn *DBConn, accountID string) *bun.SelectQuery {
	return conn.NewSelect().
		TableExpr("?", bun.Ident("blocks")).
		ColumnExpr("?", bun.Ident("?")).
		Where("? = ?", bun.Ident("account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("updated_at"))
}
