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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type relationshipDB struct {
	db    *bun.DB
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
		return nil, gtserror.Newf("error fetching follow: %w", err)
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
		return nil, gtserror.Newf("error checking followedBy: %w", err)
	}

	// check if requesting has follow requested target
	rel.Requested, err = r.IsFollowRequested(ctx,
		requestingAccount,
		targetAccount,
	)
	if err != nil {
		return nil, gtserror.Newf("error checking requested: %w", err)
	}

	// check if target has follow requested requesting
	rel.RequestedBy, err = r.IsFollowRequested(ctx,
		targetAccount,
		requestingAccount,
	)
	if err != nil {
		return nil, gtserror.Newf("error checking requestedBy: %w", err)
	}

	// check if the requesting account is blocking the target account
	rel.Blocking, err = r.IsBlocked(ctx, requestingAccount, targetAccount)
	if err != nil {
		return nil, gtserror.Newf("error checking blocking: %w", err)
	}

	// check if the requesting account is blocked by the target account
	rel.BlockedBy, err = r.IsBlocked(ctx, targetAccount, requestingAccount)
	if err != nil {
		return nil, gtserror.Newf("error checking blockedBy: %w", err)
	}

	// retrieve a note by the requesting account on the target account, if there is one
	note, err := r.GetNote(
		gtscontext.SetBarebones(ctx),
		requestingAccount,
		targetAccount,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.Newf("error fetching note: %w", err)
	}
	if note != nil {
		rel.Note = note.Comment
	}

	// check if the requesting account is muting the target account
	mute, err := r.GetMute(ctx, requestingAccount, targetAccount)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.Newf("error checking muting: %w", err)
	}
	if mute != nil && !mute.Expired(time.Now()) {
		rel.Muting = true
		rel.MutingNotifications = *mute.Notifications
	}

	return &rel, nil
}

func (r *relationshipDB) GetAccountFollows(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Follow, error) {
	followIDs, err := r.GetAccountFollowIDs(ctx, accountID, page)
	if err != nil {
		return nil, err
	}
	return r.GetFollowsByIDs(ctx, followIDs)
}

func (r *relationshipDB) GetAccountLocalFollows(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error) {
	followIDs, err := r.GetAccountLocalFollowIDs(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return r.GetFollowsByIDs(ctx, followIDs)
}

func (r *relationshipDB) GetAccountFollowers(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Follow, error) {
	followerIDs, err := r.GetAccountFollowerIDs(ctx, accountID, page)
	if err != nil {
		return nil, err
	}
	return r.GetFollowsByIDs(ctx, followerIDs)
}

func (r *relationshipDB) GetAccountLocalFollowers(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error) {
	followerIDs, err := r.GetAccountLocalFollowerIDs(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return r.GetFollowsByIDs(ctx, followerIDs)
}

func (r *relationshipDB) GetAccountFollowRequests(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.FollowRequest, error) {
	followReqIDs, err := r.GetAccountFollowRequestIDs(ctx, accountID, page)
	if err != nil {
		return nil, err
	}
	return r.GetFollowRequestsByIDs(ctx, followReqIDs)
}

func (r *relationshipDB) GetAccountFollowRequesting(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.FollowRequest, error) {
	followReqIDs, err := r.GetAccountFollowRequestingIDs(ctx, accountID, page)
	if err != nil {
		return nil, err
	}
	return r.GetFollowRequestsByIDs(ctx, followReqIDs)
}

func (r *relationshipDB) GetAccountBlocks(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Block, error) {
	blockIDs, err := r.GetAccountBlockIDs(ctx, accountID, page)
	if err != nil {
		return nil, err
	}
	return r.GetBlocksByIDs(ctx, blockIDs)
}

func (r *relationshipDB) CountAccountBlocks(ctx context.Context, accountID string) (int, error) {
	blockIDs, err := r.GetAccountBlockIDs(ctx, accountID, nil)
	return len(blockIDs), err
}

func (r *relationshipDB) GetAccountFollowIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(&r.state.Caches.DB.FollowIDs, ">"+accountID, page, func() ([]string, error) {
		var followIDs []string

		// Follow IDs not in cache, perform DB query!
		q := newSelectFollows(r.db, accountID)
		if _, err := q.Exec(ctx, &followIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return followIDs, nil
	})
}

func (r *relationshipDB) GetAccountLocalFollowIDs(ctx context.Context, accountID string) ([]string, error) {
	return r.state.Caches.DB.FollowIDs.Load("l>"+accountID, func() ([]string, error) {
		var followIDs []string

		// Follow IDs not in cache, perform DB query!
		q := newSelectLocalFollows(r.db, accountID)
		if _, err := q.Exec(ctx, &followIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return followIDs, nil
	})
}

func (r *relationshipDB) GetAccountFollowerIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(&r.state.Caches.DB.FollowIDs, "<"+accountID, page, func() ([]string, error) {
		var followIDs []string

		// Follow IDs not in cache, perform DB query!
		q := newSelectFollowers(r.db, accountID)
		if _, err := q.Exec(ctx, &followIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return followIDs, nil
	})
}

func (r *relationshipDB) GetAccountLocalFollowerIDs(ctx context.Context, accountID string) ([]string, error) {
	return r.state.Caches.DB.FollowIDs.Load("l<"+accountID, func() ([]string, error) {
		var followIDs []string

		// Follow IDs not in cache, perform DB query!
		q := newSelectLocalFollowers(r.db, accountID)
		if _, err := q.Exec(ctx, &followIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return followIDs, nil
	})
}

func (r *relationshipDB) GetAccountFollowRequestIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(&r.state.Caches.DB.FollowRequestIDs, ">"+accountID, page, func() ([]string, error) {
		var followReqIDs []string

		// Follow request IDs not in cache, perform DB query!
		q := newSelectFollowRequests(r.db, accountID)
		if _, err := q.Exec(ctx, &followReqIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return followReqIDs, nil
	})
}

func (r *relationshipDB) GetAccountFollowRequestingIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(&r.state.Caches.DB.FollowRequestIDs, "<"+accountID, page, func() ([]string, error) {
		var followReqIDs []string

		// Follow request IDs not in cache, perform DB query!
		q := newSelectFollowRequesting(r.db, accountID)
		if _, err := q.Exec(ctx, &followReqIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return followReqIDs, nil
	})
}

func (r *relationshipDB) GetAccountBlockIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(&r.state.Caches.DB.BlockIDs, accountID, page, func() ([]string, error) {
		var blockIDs []string

		// Block IDs not in cache, perform DB query!
		q := newSelectBlocks(r.db, accountID)
		if _, err := q.Exec(ctx, &blockIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return blockIDs, nil
	})
}

// newSelectFollowRequests returns a new select query for all rows in the follow_requests table with target_account_id = accountID.
func newSelectFollowRequests(db *bun.DB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		TableExpr("?", bun.Ident("follow_requests")).
		ColumnExpr("?", bun.Ident("id")).
		Where("? = ?", bun.Ident("target_account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("id"))
}

// newSelectFollowRequesting returns a new select query for all rows in the follow_requests table with account_id = accountID.
func newSelectFollowRequesting(db *bun.DB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		TableExpr("?", bun.Ident("follow_requests")).
		ColumnExpr("?", bun.Ident("id")).
		Where("? = ?", bun.Ident("target_account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("id"))
}

// newSelectFollows returns a new select query for all rows in the follows table with account_id = accountID.
func newSelectFollows(db *bun.DB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		Table("follows").
		Column("id").
		Where("? = ?", bun.Ident("account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("created_at"))
}

// newSelectLocalFollows returns a new select query for all rows in the follows table with
// account_id = accountID where the corresponding account ID has a NULL domain (i.e. is local).
func newSelectLocalFollows(db *bun.DB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		Table("follows").
		Column("id").
		Where("? = ? AND ? IN (?)",
			bun.Ident("account_id"),
			accountID,
			bun.Ident("target_account_id"),
			db.NewSelect().
				Table("accounts").
				Column("id").
				Where("? IS NULL", bun.Ident("domain")),
		).
		OrderExpr("? DESC", bun.Ident("created_at"))
}

// newSelectFollowers returns a new select query for all rows in the follows table with target_account_id = accountID.
func newSelectFollowers(db *bun.DB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		Table("follows").
		Column("id").
		Where("? = ?", bun.Ident("target_account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("created_at"))
}

// newSelectLocalFollowers returns a new select query for all rows in the follows table with
// target_account_id = accountID where the corresponding account ID has a NULL domain (i.e. is local).
func newSelectLocalFollowers(db *bun.DB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		Table("follows").
		Column("id").
		Where("? = ? AND ? IN (?)",
			bun.Ident("target_account_id"),
			accountID,
			bun.Ident("account_id"),
			db.NewSelect().
				Table("accounts").
				Column("id").
				Where("? IS NULL", bun.Ident("domain")),
		).
		OrderExpr("? DESC", bun.Ident("created_at"))
}

// newSelectBlocks returns a new select query for all rows in the blocks table with account_id = accountID.
func newSelectBlocks(db *bun.DB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		TableExpr("?", bun.Ident("blocks")).
		ColumnExpr("?", bun.Ident("id")).
		Where("? = ?", bun.Ident("account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("id"))
}
