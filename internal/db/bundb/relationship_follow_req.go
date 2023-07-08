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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/uptrace/bun"
)

func (r *relationshipDB) GetFollowRequestByID(ctx context.Context, id string) (*gtsmodel.FollowRequest, error) {
	return r.getFollowRequest(
		ctx,
		"ID",
		func(followReq *gtsmodel.FollowRequest) error {
			return r.conn.NewSelect().
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
			return r.conn.NewSelect().
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
		"AccountID.TargetAccountID",
		func(followReq *gtsmodel.FollowRequest) error {
			return r.conn.NewSelect().
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
	// Preallocate slice of expected length.
	followReqs := make([]*gtsmodel.FollowRequest, 0, len(ids))

	for _, id := range ids {
		// Fetch follow request model for this ID.
		followReq, err := r.GetFollowRequestByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting follow request %q: %v", id, err)
			continue
		}

		// Append to return slice.
		followReqs = append(followReqs, followReq)
	}

	return followReqs, nil
}

func (r *relationshipDB) IsFollowRequested(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, db.Error) {
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
	followReq, err := r.state.Caches.GTS.FollowRequest().Load(lookup, func() (*gtsmodel.FollowRequest, error) {
		var followReq gtsmodel.FollowRequest

		// Not cached! Perform database query
		if err := dbQuery(&followReq); err != nil {
			return nil, r.conn.ProcessError(err)
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

	// Set the follow request source account
	followReq.Account, err = r.state.DB.GetAccountByID(
		gtscontext.SetBarebones(ctx),
		followReq.AccountID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting follow request source account: %w", err)
	}

	// Set the follow request target account
	followReq.TargetAccount, err = r.state.DB.GetAccountByID(
		gtscontext.SetBarebones(ctx),
		followReq.TargetAccountID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting follow request target account: %w", err)
	}

	return followReq, nil
}

func (r *relationshipDB) PutFollowRequest(ctx context.Context, follow *gtsmodel.FollowRequest) error {
	return r.state.Caches.GTS.FollowRequest().Store(follow, func() error {
		_, err := r.conn.NewInsert().Model(follow).Exec(ctx)
		return r.conn.ProcessError(err)
	})
}

func (r *relationshipDB) UpdateFollowRequest(ctx context.Context, followRequest *gtsmodel.FollowRequest, columns ...string) error {
	followRequest.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	return r.state.Caches.GTS.FollowRequest().Store(followRequest, func() error {
		if _, err := r.conn.NewUpdate().
			Model(followRequest).
			Where("? = ?", bun.Ident("follow_request.id"), followRequest.ID).
			Column(columns...).
			Exec(ctx); err != nil {
			return r.conn.ProcessError(err)
		}

		return nil
	})
}

func (r *relationshipDB) AcceptFollowRequest(ctx context.Context, sourceAccountID string, targetAccountID string) (*gtsmodel.Follow, db.Error) {
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

	if err := r.state.Caches.GTS.Follow().Store(follow, func() error {
		// If the follow already exists, just
		// replace the URI with the new one.
		_, err := r.conn.
			NewInsert().
			Model(follow).
			On("CONFLICT (?,?) DO UPDATE set ? = ?", bun.Ident("account_id"), bun.Ident("target_account_id"), bun.Ident("uri"), follow.URI).
			Exec(ctx)
		return r.conn.ProcessError(err)
	}); err != nil {
		return nil, err
	}

	// Invalidate follow request from cache lookups on return.
	defer r.state.Caches.GTS.FollowRequest().Invalidate("ID", followReq.ID)

	// Delete original follow request.
	if _, err := r.conn.
		NewDelete().
		Table("follow_requests").
		Where("? = ?", bun.Ident("id"), followReq.ID).
		Exec(ctx); err != nil {
		return nil, r.conn.ProcessError(err)
	}

	// Delete original follow request notification
	if err := r.state.DB.DeleteNotifications(ctx, []string{
		string(gtsmodel.NotificationFollowRequest),
	}, targetAccountID, sourceAccountID); err != nil {
		return nil, err
	}

	return follow, nil
}

func (r *relationshipDB) RejectFollowRequest(ctx context.Context, sourceAccountID string, targetAccountID string) db.Error {
	// Delete follow request first.
	if err := r.DeleteFollowRequest(ctx, sourceAccountID, targetAccountID); err != nil {
		return err
	}

	// Delete follow request notification
	return r.state.DB.DeleteNotifications(ctx, []string{
		string(gtsmodel.NotificationFollowRequest),
	}, targetAccountID, sourceAccountID)
}

func (r *relationshipDB) DeleteFollowRequest(ctx context.Context, sourceAccountID string, targetAccountID string) error {
	defer r.state.Caches.GTS.FollowRequest().Invalidate("AccountID.TargetAccountID", sourceAccountID, targetAccountID)

	// Load followreq into cache before attempting a delete,
	// as we need it cached in order to trigger the invalidate
	// callback. This in turn invalidates others.
	follow, err := r.GetFollowRequest(
		gtscontext.SetBarebones(ctx),
		sourceAccountID,
		targetAccountID,
	)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// Already gone.
			return nil
		}
		return err
	}

	// Finally delete followreq from DB.
	_, err = r.conn.NewDelete().
		Table("follow_requests").
		Where("? = ?", bun.Ident("id"), follow.ID).
		Exec(ctx)
	return r.conn.ProcessError(err)
}

func (r *relationshipDB) DeleteFollowRequestByID(ctx context.Context, id string) error {
	defer r.state.Caches.GTS.FollowRequest().Invalidate("ID", id)

	// Load followreq into cache before attempting a delete,
	// as we need it cached in order to trigger the invalidate
	// callback. This in turn invalidates others.
	_, err := r.GetFollowRequestByID(gtscontext.SetBarebones(ctx), id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// not an issue.
			err = nil
		}
		return err
	}

	// Finally delete followreq from DB.
	_, err = r.conn.NewDelete().
		Table("follow_requests").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx)
	return r.conn.ProcessError(err)
}

func (r *relationshipDB) DeleteFollowRequestByURI(ctx context.Context, uri string) error {
	defer r.state.Caches.GTS.FollowRequest().Invalidate("URI", uri)

	// Load followreq into cache before attempting a delete,
	// as we need it cached in order to trigger the invalidate
	// callback. This in turn invalidates others.
	_, err := r.GetFollowRequestByURI(gtscontext.SetBarebones(ctx), uri)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// not an issue.
			err = nil
		}
		return err
	}

	// Finally delete followreq from DB.
	_, err = r.conn.NewDelete().
		Table("follow_requests").
		Where("? = ?", bun.Ident("uri"), uri).
		Exec(ctx)
	return r.conn.ProcessError(err)
}

func (r *relationshipDB) DeleteAccountFollowRequests(ctx context.Context, accountID string) error {
	var followReqIDs []string

	// Get full list of IDs.
	if _, err := r.conn.
		NewSelect().
		Column("id").
		Table("follow_requestss").
		WhereOr("? = ? OR ? = ?",
			bun.Ident("account_id"),
			accountID,
			bun.Ident("target_account_id"),
			accountID,
		).
		Exec(ctx, &followReqIDs); err != nil {
		return r.conn.ProcessError(err)
	}

	defer func() {
		// Invalidate all IDs on return.
		for _, id := range followReqIDs {
			r.state.Caches.GTS.FollowRequest().Invalidate("ID", id)
		}
	}()

	// Load all followreqs into cache, this *really* isn't
	// great but it is the only way we can ensure we invalidate
	// all related caches correctly (e.g. visibility).
	for _, id := range followReqIDs {
		_, err := r.GetFollowRequestByID(ctx, id)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return err
		}
	}

	// Finally delete all from DB.
	_, err := r.conn.NewDelete().
		Table("follow_requests").
		Where("? IN (?)", bun.Ident("id"), bun.In(followReqIDs)).
		Exec(ctx)
	return r.conn.ProcessError(err)
}
