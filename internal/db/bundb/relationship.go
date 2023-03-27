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
