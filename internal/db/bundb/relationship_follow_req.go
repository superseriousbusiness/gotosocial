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
	"slices"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

func (r *relationshipDB) GetFollowRequestByID(ctx context.Context, id string) (*gtsmodel.FollowRequest, error) {
	return r.getFollowRequest(
		ctx,
		"ID",
		func(followReq *gtsmodel.FollowRequest) error {
			return r.db.NewSelect().
				Model(followReq).
				Where("? = ?", bun.Ident("id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (r *relationshipDB) GetFollowRequestByURI(ctx context.Context, uri string) (*gtsmodel.FollowRequest, error) {
	return r.getFollowRequest(
		ctx,
		"URI",
		func(followReq *gtsmodel.FollowRequest) error {
			return r.db.NewSelect().
				Model(followReq).
				Where("? = ?", bun.Ident("uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (r *relationshipDB) GetFollowRequest(ctx context.Context, sourceAccountID string, targetAccountID string) (*gtsmodel.FollowRequest, error) {
	return r.getFollowRequest(
		ctx,
		"AccountID,TargetAccountID",
		func(followReq *gtsmodel.FollowRequest) error {
			return r.db.NewSelect().
				Model(followReq).
				Where("? = ?", bun.Ident("account_id"), sourceAccountID).
				Where("? = ?", bun.Ident("target_account_id"), targetAccountID).
				Scan(ctx)
		},
		sourceAccountID,
		targetAccountID,
	)
}

func (r *relationshipDB) GetFollowRequestsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.FollowRequest, error) {
	// Load all follow IDs via cache loader callbacks.
	follows, err := r.state.Caches.DB.FollowRequest.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.FollowRequest, error) {
			// Preallocate expected length of uncached followReqs.
			follows := make([]*gtsmodel.FollowRequest, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := r.db.NewSelect().
				Model(&follows).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return follows, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the requests by their
	// IDs to ensure in correct order.
	getID := func(f *gtsmodel.FollowRequest) string { return f.ID }
	xslices.OrderBy(follows, ids, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return follows, nil
	}

	// Populate all loaded followreqs, removing those we fail to
	// populate (removes needing so many nil checks everywhere).
	follows = slices.DeleteFunc(follows, func(follow *gtsmodel.FollowRequest) bool {
		if err := r.PopulateFollowRequest(ctx, follow); err != nil {
			log.Errorf(ctx, "error populating follow request %s: %v", follow.ID, err)
			return true
		}
		return false
	})

	return follows, nil
}

func (r *relationshipDB) IsFollowRequested(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, error) {
	followReq, err := r.GetFollowRequest(
		gtscontext.SetBarebones(ctx),
		sourceAccountID,
		targetAccountID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return false, err
	}
	return (followReq != nil), nil
}

func (r *relationshipDB) getFollowRequest(ctx context.Context, lookup string, dbQuery func(*gtsmodel.FollowRequest) error, keyParts ...any) (*gtsmodel.FollowRequest, error) {
	// Fetch follow request from database cache with loader callback
	followReq, err := r.state.Caches.DB.FollowRequest.LoadOne(lookup, func() (*gtsmodel.FollowRequest, error) {
		var followReq gtsmodel.FollowRequest

		// Not cached! Perform database query
		if err := dbQuery(&followReq); err != nil {
			return nil, err
		}

		return &followReq, nil
	}, keyParts...)
	if err != nil {
		// error already processed
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return followReq, nil
	}

	if err := r.state.DB.PopulateFollowRequest(ctx, followReq); err != nil {
		return nil, err
	}

	return followReq, nil
}

func (r *relationshipDB) PopulateFollowRequest(ctx context.Context, follow *gtsmodel.FollowRequest) error {
	var (
		err  error
		errs = gtserror.NewMultiError(2)
	)

	if follow.Account == nil {
		// Follow account is not set, fetch from the database.
		follow.Account, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			follow.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating follow request account: %w", err)
		}
	}

	if follow.TargetAccount == nil {
		// Follow target account is not set, fetch from the database.
		follow.TargetAccount, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			follow.TargetAccountID,
		)
		if err != nil {
			errs.Appendf("error populating follow target request account: %w", err)
		}
	}

	return errs.Combine()
}

func (r *relationshipDB) PutFollowRequest(ctx context.Context, follow *gtsmodel.FollowRequest) error {
	return r.state.Caches.DB.FollowRequest.Store(follow, func() error {
		_, err := r.db.NewInsert().Model(follow).Exec(ctx)
		return err
	})
}

func (r *relationshipDB) UpdateFollowRequest(ctx context.Context, followRequest *gtsmodel.FollowRequest, columns ...string) error {
	followRequest.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	return r.state.Caches.DB.FollowRequest.Store(followRequest, func() error {
		if _, err := r.db.NewUpdate().
			Model(followRequest).
			Where("? = ?", bun.Ident("follow_request.id"), followRequest.ID).
			Column(columns...).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}

func (r *relationshipDB) AcceptFollowRequest(ctx context.Context, sourceAccountID string, targetAccountID string) (*gtsmodel.Follow, error) {
	// Get original follow request.
	followReq, err := r.GetFollowRequest(ctx, sourceAccountID, targetAccountID)
	if err != nil {
		return nil, err
	}

	// Create a new follow to 'replace'
	// the original follow request with.
	follow := &gtsmodel.Follow{
		ID:              followReq.ID,
		AccountID:       sourceAccountID,
		Account:         followReq.Account,
		TargetAccountID: targetAccountID,
		TargetAccount:   followReq.TargetAccount,
		URI:             followReq.URI,
		ShowReblogs:     followReq.ShowReblogs,
		Notify:          followReq.Notify,
	}

	if err := r.state.Caches.DB.Follow.Store(follow, func() error {
		// If the follow already exists, just
		// replace the URI with the new one.
		_, err := r.db.
			NewInsert().
			Model(follow).
			On("CONFLICT (?,?) DO UPDATE set ? = ?", bun.Ident("account_id"), bun.Ident("target_account_id"), bun.Ident("uri"), follow.URI).
			Exec(ctx)
		return err
	}); err != nil {
		return nil, err
	}

	// Delete original follow request.
	if _, err := r.db.
		NewDelete().
		Table("follow_requests").
		Where("? = ?", bun.Ident("id"), followReq.ID).
		Exec(ctx); err != nil {
		return nil, err
	}

	// Delete original follow request notification
	if err := r.state.DB.DeleteNotifications(ctx, []string{
		string(gtsmodel.NotificationFollowRequest),
	}, targetAccountID, sourceAccountID); err != nil {
		return nil, err
	}

	return follow, nil
}

func (r *relationshipDB) RejectFollowRequest(ctx context.Context, sourceAccountID string, targetAccountID string) error {
	// Delete follow request first.
	if err := r.DeleteFollowRequest(ctx, sourceAccountID, targetAccountID); err != nil {
		return err
	}

	// Delete follow request notification
	return r.state.DB.DeleteNotifications(ctx, []string{
		string(gtsmodel.NotificationFollowRequest),
	}, targetAccountID, sourceAccountID)
}

func (r *relationshipDB) DeleteFollowRequest(
	ctx context.Context,
	sourceAccountID string,
	targetAccountID string,
) error {

	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted gtsmodel.FollowRequest
	deleted.AccountID = sourceAccountID
	deleted.TargetAccountID = targetAccountID

	// Delete all follow reqs either
	// from account, or targeting account,
	// returning the deleted models.
	if _, err := r.db.NewDelete().
		Model(&deleted).
		Where("? = ?", bun.Ident("account_id"), sourceAccountID).
		Where("? = ?", bun.Ident("target_account_id"), targetAccountID).
		Returning("?", bun.Ident("id")).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate cached follow with source / target account IDs,
	// manually calling invalidate hook in case it isn't cached.
	r.state.Caches.DB.FollowRequest.Invalidate("AccountID,TargetAccountID",
		sourceAccountID, targetAccountID)
	r.state.Caches.OnInvalidateFollowRequest(&deleted)

	return nil
}

func (r *relationshipDB) DeleteFollowRequestByID(ctx context.Context, id string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted gtsmodel.FollowRequest
	deleted.ID = id

	// Delete follow with given URI,
	// returning the deleted models.
	if _, err := r.db.NewDelete().
		Model(&deleted).
		Where("? = ?", bun.Ident("id"), id).
		Returning("?, ?",
			bun.Ident("account_id"),
			bun.Ident("target_account_id"),
		).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate cached follow with URI, manually
	// call invalidate hook in case not cached.
	r.state.Caches.DB.FollowRequest.Invalidate("ID", id)
	r.state.Caches.OnInvalidateFollowRequest(&deleted)

	return nil
}

func (r *relationshipDB) DeleteFollowRequestByURI(ctx context.Context, uri string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted gtsmodel.FollowRequest

	// Delete follow with given URI,
	// returning the deleted models.
	if _, err := r.db.NewDelete().
		Model(&deleted).
		Where("? = ?", bun.Ident("uri"), uri).
		Returning("?, ?, ?",
			bun.Ident("id"),
			bun.Ident("account_id"),
			bun.Ident("target_account_id"),
		).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate cached follow with URI, manually
	// call invalidate hook in case not cached.
	r.state.Caches.DB.FollowRequest.Invalidate("URI", uri)
	r.state.Caches.OnInvalidateFollowRequest(&deleted)

	return nil
}

func (r *relationshipDB) DeleteAccountFollowRequests(ctx context.Context, accountID string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted []*gtsmodel.FollowRequest

	// Delete all follows either from
	// account, or targeting account,
	// returning the deleted models.
	if _, err := r.db.NewDelete().
		Model(&deleted).
		WhereOr("? = ? OR ? = ?",
			bun.Ident("account_id"),
			accountID,
			bun.Ident("target_account_id"),
			accountID,
		).
		Returning("?, ?, ?",
			bun.Ident("id"),
			bun.Ident("account_id"),
			bun.Ident("target_account_id"),
		).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate all account's incoming / outoing follows requests.
	r.state.Caches.DB.FollowRequest.Invalidate("AccountID", accountID)
	r.state.Caches.DB.FollowRequest.Invalidate("TargetAccountID", accountID)

	// In case not all follow were in
	// cache, manually call invalidate hooks.
	for _, followReq := range deleted {
		r.state.Caches.OnInvalidateFollowRequest(followReq)
	}

	return nil
}
