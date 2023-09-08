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

	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type relationshipDB struct {
	db    *DB
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

	return &rel, nil
}

func (r *relationshipDB) GetAccountFollows(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Follow, error) {
	followIDs, err := r.getAccountFollowIDs(ctx, accountID, page)
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

func (r *relationshipDB) GetAccountFollowers(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Follow, error) {
	followerIDs, err := r.getAccountFollowerIDs(ctx, accountID, page)
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

func (r *relationshipDB) GetAccountFollowRequests(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.FollowRequest, error) {
	followReqIDs, err := r.getAccountFollowRequestIDs(ctx, accountID, page)
	if err != nil {
		return nil, err
	}
	return r.GetFollowRequestsByIDs(ctx, followReqIDs)
}

func (r *relationshipDB) GetAccountFollowRequesting(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.FollowRequest, error) {
	followReqIDs, err := r.getAccountFollowRequestingIDs(ctx, accountID, page)
	if err != nil {
		return nil, err
	}
	return r.GetFollowRequestsByIDs(ctx, followReqIDs)
}

func (r *relationshipDB) GetAccountBlocks(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Block, error) {
	blockIDs, err := r.getAccountBlockIDs(ctx, accountID, page)
	if err != nil {
		return nil, err
	}
	return r.GetBlocksByIDs(ctx, blockIDs)
}

func (r *relationshipDB) CountAccountFollows(ctx context.Context, accountID string) (int, error) {
	followIDs, err := r.getAccountFollowIDs(ctx, accountID, nil)
	return len(followIDs), err
}

func (r *relationshipDB) CountAccountLocalFollows(ctx context.Context, accountID string) (int, error) {
	followIDs, err := r.getAccountLocalFollowIDs(ctx, accountID)
	return len(followIDs), err
}

func (r *relationshipDB) CountAccountFollowers(ctx context.Context, accountID string) (int, error) {
	followerIDs, err := r.getAccountFollowerIDs(ctx, accountID, nil)
	return len(followerIDs), err
}

func (r *relationshipDB) CountAccountLocalFollowers(ctx context.Context, accountID string) (int, error) {
	followerIDs, err := r.getAccountLocalFollowerIDs(ctx, accountID)
	return len(followerIDs), err
}

func (r *relationshipDB) CountAccountFollowRequests(ctx context.Context, accountID string) (int, error) {
	followReqIDs, err := r.getAccountFollowRequestIDs(ctx, accountID, nil)
	return len(followReqIDs), err
}

func (r *relationshipDB) CountAccountFollowRequesting(ctx context.Context, accountID string) (int, error) {
	followReqIDs, err := r.getAccountFollowRequestingIDs(ctx, accountID, nil)
	return len(followReqIDs), err
}

func (r *relationshipDB) CountAccountBlocks(ctx context.Context, accountID string) (int, error) {
	blockIDs, err := r.getAccountBlockIDs(ctx, accountID, nil)
	return len(blockIDs), err
}

func (r *relationshipDB) getAccountFollowIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(r.state.Caches.GTS.FollowIDs(), ">"+accountID, page, func() ([]string, error) {
		var followIDs []string

		// Follow IDs not in cache, perform DB query!
		q := newSelectFollows(r.db, accountID)
		if _, err := q.Exec(ctx, &followIDs); err != nil {
			return nil, err
		}

		return followIDs, nil
	})
}

func (r *relationshipDB) getAccountLocalFollowIDs(ctx context.Context, accountID string) ([]string, error) {
	return r.state.Caches.GTS.FollowIDs().Load("l>"+accountID, func() ([]string, error) {
		var followIDs []string

		// Follow IDs not in cache, perform DB query!
		q := newSelectLocalFollows(r.db, accountID)
		if _, err := q.Exec(ctx, &followIDs); err != nil {
			return nil, err
		}

		return followIDs, nil
	})
}

func (r *relationshipDB) getAccountFollowerIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(r.state.Caches.GTS.FollowIDs(), "<"+accountID, page, func() ([]string, error) {
		var followIDs []string

		// Follow IDs not in cache, perform DB query!
		q := newSelectFollowers(r.db, accountID)
		if _, err := q.Exec(ctx, &followIDs); err != nil {
			return nil, err
		}

		return followIDs, nil
	})
}

func (r *relationshipDB) getAccountLocalFollowerIDs(ctx context.Context, accountID string) ([]string, error) {
	return r.state.Caches.GTS.FollowIDs().Load("l<"+accountID, func() ([]string, error) {
		var followIDs []string

		// Follow IDs not in cache, perform DB query!
		q := newSelectLocalFollowers(r.db, accountID)
		if _, err := q.Exec(ctx, &followIDs); err != nil {
			return nil, err
		}

		return followIDs, nil
	})
}

func (r *relationshipDB) getAccountFollowRequestIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(r.state.Caches.GTS.FollowRequestIDs(), ">"+accountID, page, func() ([]string, error) {
		var followReqIDs []string

		// Follow request IDs not in cache, perform DB query!
		q := newSelectFollowRequests(r.db, accountID)
		if _, err := q.Exec(ctx, &followReqIDs); err != nil {
			return nil, err
		}

		return followReqIDs, nil
	})
}

func (r *relationshipDB) getAccountFollowRequestingIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(r.state.Caches.GTS.FollowRequestIDs(), "<"+accountID, page, func() ([]string, error) {
		var followReqIDs []string

		// Follow request IDs not in cache, perform DB query!
		q := newSelectFollowRequesting(r.db, accountID)
		if _, err := q.Exec(ctx, &followReqIDs); err != nil {
			return nil, err
		}

		return followReqIDs, nil
	})
}

func (r *relationshipDB) getAccountBlockIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(r.state.Caches.GTS.BlockIDs(), accountID, page, func() ([]string, error) {
		var blockIDs []string

		// Block IDs not in cache, perform DB query!
		q := newSelectBlocks(r.db, accountID)
		if _, err := q.Exec(ctx, &blockIDs); err != nil {
			return nil, err
		}

		return blockIDs, nil
	})
}

// loadPagedIDs loads a page of IDs from given SliceCache by `key`, resorting to `load` function if required. Uses `page` to sort + page resulting IDs.
// NOTE: IDs returned from `cache` / `load` MUST be in descending order, otherwise paging will not work correctly / return things out of order.
func loadPagedIDs(cache *cache.SliceCache[string], key string, page *paging.Page, load func() ([]string, error)) ([]string, error) {
	// Check cache for IDs, else load.
	ids, err := cache.Load(key, load)
	if err != nil {
		return nil, err
	}

	// Our cached / selected bIDs are
	// ALWAYS stored in descending order.
	// Depending on the paging requested
	// this may be an unexpected order.
	if !page.GetOrder().Ascending() {
		ids = paging.Reverse(ids)
	}

	// Page the resulting IDs.
	ids = page.Page(ids)

	return ids, nil
}

// newSelectFollowRequests returns a new select query for all rows in the follow_requests table with target_account_id = accountID.
func newSelectFollowRequests(db *DB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		TableExpr("?", bun.Ident("follow_requests")).
		ColumnExpr("?", bun.Ident("id")).
		Where("? = ?", bun.Ident("target_account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("id"))
}

// newSelectFollowRequesting returns a new select query for all rows in the follow_requests table with account_id = accountID.
func newSelectFollowRequesting(db *DB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		TableExpr("?", bun.Ident("follow_requests")).
		ColumnExpr("?", bun.Ident("id")).
		Where("? = ?", bun.Ident("target_account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("id"))
}

// newSelectFollows returns a new select query for all rows in the follows table with account_id = accountID.
func newSelectFollows(db *DB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		Table("follows").
		Column("id").
		Where("? = ?", bun.Ident("account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("id"))
}

// newSelectLocalFollows returns a new select query for all rows in the follows table with
// account_id = accountID where the corresponding account ID has a NULL domain (i.e. is local).
func newSelectLocalFollows(db *DB, accountID string) *bun.SelectQuery {
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
		OrderExpr("? DESC", bun.Ident("id"))
}

// newSelectFollowers returns a new select query for all rows in the follows table with target_account_id = accountID.
func newSelectFollowers(db *DB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		Table("follows").
		Column("id").
		Where("? = ?", bun.Ident("target_account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("id"))
}

// newSelectLocalFollowers returns a new select query for all rows in the follows table with
// target_account_id = accountID where the corresponding account ID has a NULL domain (i.e. is local).
func newSelectLocalFollowers(db *DB, accountID string) *bun.SelectQuery {
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
		OrderExpr("? DESC", bun.Ident("id"))
}

// newSelectBlocks returns a new select query for all rows in the blocks table with account_id = accountID.
func newSelectBlocks(db *DB, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		TableExpr("?", bun.Ident("blocks")).
		ColumnExpr("?", bun.Ident("?")).
		Where("? = ?", bun.Ident("account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("id"))
}
