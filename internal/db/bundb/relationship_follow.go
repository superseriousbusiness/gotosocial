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
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/uptrace/bun"
)

func (r *relationshipDB) GetFollowByID(ctx context.Context, id string) (*gtsmodel.Follow, error) {
	return r.getFollow(
		ctx,
		"ID",
		func(follow *gtsmodel.Follow) error {
			return r.conn.NewSelect().
				Model(follow).
				Where("? = ?", bun.Ident("id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (r *relationshipDB) GetFollowByURI(ctx context.Context, uri string) (*gtsmodel.Follow, error) {
	return r.getFollow(
		ctx,
		"URI",
		func(follow *gtsmodel.Follow) error {
			return r.conn.NewSelect().
				Model(follow).
				Where("? = ?", bun.Ident("uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (r *relationshipDB) GetFollow(ctx context.Context, sourceAccountID string, targetAccountID string) (*gtsmodel.Follow, error) {
	return r.getFollow(
		ctx,
		"AccountID.TargetAccountID",
		func(follow *gtsmodel.Follow) error {
			return r.conn.NewSelect().
				Model(follow).
				Where("? = ?", bun.Ident("account_id"), sourceAccountID).
				Where("? = ?", bun.Ident("target_account_id"), targetAccountID).
				Scan(ctx)
		},
		sourceAccountID,
		targetAccountID,
	)
}

func (r *relationshipDB) GetFollowsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Follow, error) {
	// Preallocate slice of expected length.
	follows := make([]*gtsmodel.Follow, 0, len(ids))

	for _, id := range ids {
		// Fetch follow model for this ID.
		follow, err := r.GetFollowByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting follow %q: %v", id, err)
			continue
		}

		// Append to return slice.
		follows = append(follows, follow)
	}

	return follows, nil
}

func (r *relationshipDB) IsFollowing(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, db.Error) {
	follow, err := r.GetFollow(
		gtscontext.SetBarebones(ctx),
		sourceAccountID,
		targetAccountID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return false, err
	}
	return (follow != nil), nil
}

func (r *relationshipDB) IsMutualFollowing(ctx context.Context, accountID1 string, accountID2 string) (bool, db.Error) {
	// make sure account 1 follows account 2
	f1, err := r.IsFollowing(ctx,
		accountID1,
		accountID2,
	)
	if !f1 /* f1 = false when err != nil */ {
		return false, err
	}

	// make sure account 2 follows account 1
	f2, err := r.IsFollowing(ctx,
		accountID2,
		accountID1,
	)
	if !f2 /* f2 = false when err != nil */ {
		return false, err
	}

	return true, nil
}

func (r *relationshipDB) getFollow(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Follow) error, keyParts ...any) (*gtsmodel.Follow, error) {
	// Fetch follow from database cache with loader callback
	follow, err := r.state.Caches.GTS.Follow().Load(lookup, func() (*gtsmodel.Follow, error) {
		var follow gtsmodel.Follow

		// Not cached! Perform database query
		if err := dbQuery(&follow); err != nil {
			return nil, r.conn.ProcessError(err)
		}

		return &follow, nil
	}, keyParts...)
	if err != nil {
		// error already processed
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return follow, nil
	}

	if err := r.state.DB.PopulateFollow(ctx, follow); err != nil {
		return nil, err
	}

	return follow, nil
}

func (r *relationshipDB) PopulateFollow(ctx context.Context, follow *gtsmodel.Follow) error {
	var (
		err  error
		errs = make(gtserror.MultiError, 0, 2)
	)

	if follow.Account == nil {
		// Follow account is not set, fetch from the database.
		follow.Account, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			follow.AccountID,
		)
		if err != nil {
			errs.Append(fmt.Errorf("error populating follow account: %w", err))
		}
	}

	if follow.TargetAccount == nil {
		// Follow target account is not set, fetch from the database.
		follow.TargetAccount, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			follow.TargetAccountID,
		)
		if err != nil {
			errs.Append(fmt.Errorf("error populating follow target account: %w", err))
		}
	}

	return errs.Combine()
}

func (r *relationshipDB) PutFollow(ctx context.Context, follow *gtsmodel.Follow) error {
	return r.state.Caches.GTS.Follow().Store(follow, func() error {
		_, err := r.conn.NewInsert().Model(follow).Exec(ctx)
		return r.conn.ProcessError(err)
	})
}

func (r *relationshipDB) UpdateFollow(ctx context.Context, follow *gtsmodel.Follow, columns ...string) error {
	follow.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	return r.state.Caches.GTS.Follow().Store(follow, func() error {
		if _, err := r.conn.NewUpdate().
			Model(follow).
			Where("? = ?", bun.Ident("follow.id"), follow.ID).
			Column(columns...).
			Exec(ctx); err != nil {
			return r.conn.ProcessError(err)
		}

		return nil
	})
}

func (r *relationshipDB) deleteFollow(ctx context.Context, id string) error {
	// Delete the follow itself using the given ID.
	if _, err := r.conn.NewDelete().
		Table("follows").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil {
		return r.conn.ProcessError(err)
	}

	// Delete every list entry that used this followID.
	if err := r.state.DB.DeleteListEntriesForFollowID(ctx, id); err != nil {
		return fmt.Errorf("deleteFollow: error deleting list entries: %w", err)
	}

	return nil
}

func (r *relationshipDB) DeleteFollowByID(ctx context.Context, id string) error {
	defer r.state.Caches.GTS.Follow().Invalidate("ID", id)

	// Load follow into cache before attempting a delete,
	// as we need it cached in order to trigger the invalidate
	// callback. This in turn invalidates others.
	follow, err := r.GetFollowByID(gtscontext.SetBarebones(ctx), id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// Already gone.
			return nil
		}
		return err
	}

	// Finally delete follow from DB.
	return r.deleteFollow(ctx, follow.ID)
}

func (r *relationshipDB) DeleteFollowByURI(ctx context.Context, uri string) error {
	defer r.state.Caches.GTS.Follow().Invalidate("URI", uri)

	// Load follow into cache before attempting a delete,
	// as we need it cached in order to trigger the invalidate
	// callback. This in turn invalidates others.
	follow, err := r.GetFollowByURI(gtscontext.SetBarebones(ctx), uri)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// Already gone.
			return nil
		}
		return err
	}

	// Finally delete follow from DB.
	return r.deleteFollow(ctx, follow.ID)
}

func (r *relationshipDB) DeleteAccountFollows(ctx context.Context, accountID string) error {
	var followIDs []string

	// Get full list of IDs.
	if _, err := r.conn.
		NewSelect().
		Column("id").
		Table("follows").
		WhereOr("? = ? OR ? = ?",
			bun.Ident("account_id"),
			accountID,
			bun.Ident("target_account_id"),
			accountID,
		).
		Exec(ctx, &followIDs); err != nil {
		return r.conn.ProcessError(err)
	}

	defer func() {
		// Invalidate all IDs on return.
		for _, id := range followIDs {
			r.state.Caches.GTS.Follow().Invalidate("ID", id)
		}
	}()

	// Load all follows into cache, this *really* isn't great
	// but it is the only way we can ensure we invalidate all
	// related caches correctly (e.g. visibility).
	for _, id := range followIDs {
		follow, err := r.GetFollowByID(ctx, id)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return err
		}

		// Delete each follow from DB.
		if err := r.deleteFollow(ctx, follow.ID); err != nil && !errors.Is(err, db.ErrNoEntries) {
			return err
		}
	}

	return nil
}
