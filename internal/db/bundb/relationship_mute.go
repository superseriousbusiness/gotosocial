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

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func (r *relationshipDB) IsMuted(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, error) {
	mute, err := r.GetMute(
		gtscontext.SetBarebones(ctx),
		sourceAccountID,
		targetAccountID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return false, err
	}
	return mute != nil, nil
}

func (r *relationshipDB) GetMuteByID(ctx context.Context, id string) (*gtsmodel.UserMute, error) {
	return r.getMute(
		ctx,
		"ID",
		func(mute *gtsmodel.UserMute) error {
			return r.db.NewSelect().Model(mute).
				Where("? = ?", bun.Ident("id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (r *relationshipDB) GetMute(
	ctx context.Context,
	sourceAccountID string,
	targetAccountID string,
) (*gtsmodel.UserMute, error) {
	return r.getMute(
		ctx,
		"AccountID,TargetAccountID",
		func(mute *gtsmodel.UserMute) error {
			return r.db.NewSelect().Model(mute).
				Where("? = ?", bun.Ident("account_id"), sourceAccountID).
				Where("? = ?", bun.Ident("target_account_id"), targetAccountID).
				Scan(ctx)
		},
		sourceAccountID,
		targetAccountID,
	)
}

func (r *relationshipDB) CountAccountMutes(ctx context.Context, accountID string) (int, error) {
	muteIDs, err := r.getAccountMuteIDs(ctx, accountID, nil)
	return len(muteIDs), err
}

func (r *relationshipDB) getMutesByIDs(ctx context.Context, ids []string) ([]*gtsmodel.UserMute, error) {
	// Load all mutes IDs via cache loader callbacks.
	mutes, err := r.state.Caches.DB.UserMute.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.UserMute, error) {
			// Preallocate expected length of uncached mutes.
			mutes := make([]*gtsmodel.UserMute, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := r.db.NewSelect().
				Model(&mutes).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return mutes, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the mutes by their
	// IDs to ensure in correct order.
	getID := func(b *gtsmodel.UserMute) string { return b.ID }
	xslices.OrderBy(mutes, ids, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return mutes, nil
	}

	// Populate all loaded mutes, removing those we fail to
	// populate (removes needing so many nil checks everywhere).
	mutes = slices.DeleteFunc(mutes, func(mute *gtsmodel.UserMute) bool {
		if err := r.populateMute(ctx, mute); err != nil {
			log.Errorf(ctx, "error populating mute %s: %v", mute.ID, err)
			return true
		}
		return false
	})

	return mutes, nil
}

func (r *relationshipDB) getMute(
	ctx context.Context,
	lookup string,
	dbQuery func(*gtsmodel.UserMute) error,
	keyParts ...any,
) (*gtsmodel.UserMute, error) {
	// Fetch mute from cache with loader callback
	mute, err := r.state.Caches.DB.UserMute.LoadOne(lookup, func() (*gtsmodel.UserMute, error) {
		var mute gtsmodel.UserMute

		// Not cached! Perform database query
		if err := dbQuery(&mute); err != nil {
			return nil, err
		}

		return &mute, nil
	}, keyParts...)
	if err != nil {
		// already processe
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return mute, nil
	}

	if err := r.populateMute(ctx, mute); err != nil {
		return nil, err
	}

	return mute, nil
}

func (r *relationshipDB) populateMute(ctx context.Context, mute *gtsmodel.UserMute) error {
	var (
		errs gtserror.MultiError
		err  error
	)

	if mute.Account == nil {
		// Mute origin account is not set, fetch from database.
		mute.Account, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			mute.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating mute account: %w", err)
		}
	}

	if mute.TargetAccount == nil {
		// Mute target account is not set, fetch from database.
		mute.TargetAccount, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			mute.TargetAccountID,
		)
		if err != nil {
			errs.Appendf("error populating mute target account: %w", err)
		}
	}

	return errs.Combine()
}

func (r *relationshipDB) PutMute(ctx context.Context, mute *gtsmodel.UserMute) error {
	return r.state.Caches.DB.UserMute.Store(mute, func() error {
		_, err := NewUpsert(r.db).Model(mute).Constraint("id").Exec(ctx)
		return err
	})
}

func (r *relationshipDB) DeleteMuteByID(ctx context.Context, id string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted gtsmodel.UserMute

	// Delete mute with given ID,
	// returning the deleted models.
	if _, err := r.db.NewDelete().
		Model(&deleted).
		Where("? = ?", bun.Ident("id"), id).
		Returning("?", bun.Ident("account_id")).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate cached mute with ID, manually
	// call invalidate hook in case not cached.
	r.state.Caches.DB.UserMute.Invalidate("ID", id)
	r.state.Caches.OnInvalidateUserMute(&deleted)

	return nil
}

func (r *relationshipDB) DeleteAccountMutes(ctx context.Context, accountID string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted []*gtsmodel.UserMute

	// Delete all mutes either from
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
		Returning("?",
			bun.Ident("account_id"),
		).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate all account's incoming / outoing user mutes.
	r.state.Caches.DB.UserMute.Invalidate("AccountID", accountID)
	r.state.Caches.DB.UserMute.Invalidate("TargetAccountID", accountID)

	// In case not all user mutes were in
	// cache, manually call invalidate hooks.
	for _, block := range deleted {
		r.state.Caches.OnInvalidateUserMute(block)
	}

	return nil
}

func (r *relationshipDB) GetAccountMutes(
	ctx context.Context,
	accountID string,
	page *paging.Page,
) ([]*gtsmodel.UserMute, error) {
	muteIDs, err := r.getAccountMuteIDs(ctx, accountID, page)
	if err != nil {
		return nil, err
	}
	return r.getMutesByIDs(ctx, muteIDs)
}

func (r *relationshipDB) getAccountMuteIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(&r.state.Caches.DB.UserMuteIDs, accountID, page, func() ([]string, error) {
		var muteIDs []string

		// Mute IDs not in cache. Perform DB query.
		if _, err := r.db.
			NewSelect().
			TableExpr("?", bun.Ident("user_mutes")).
			ColumnExpr("?", bun.Ident("id")).
			Where("? = ?", bun.Ident("account_id"), accountID).
			WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
				var notYetExpiredSQL string
				switch r.db.Dialect().Name() {
				case dialect.SQLite:
					notYetExpiredSQL = "? > DATE('now')"
				case dialect.PG:
					notYetExpiredSQL = "? > NOW()"
				default:
					log.Panicf(nil, "db conn %s was neither pg nor sqlite", r.db)
				}
				return q.
					Where("? IS NULL", bun.Ident("expires_at")).
					WhereOr(notYetExpiredSQL, bun.Ident("expires_at"))
			}).
			OrderExpr("? DESC", bun.Ident("id")).
			Exec(ctx, &muteIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return muteIDs, nil
	})
}
