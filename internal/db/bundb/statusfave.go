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
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

type statusFaveDB struct {
	db    *bun.DB
	state *state.State
}

func (s *statusFaveDB) GetStatusFave(ctx context.Context, accountID string, statusID string) (*gtsmodel.StatusFave, error) {
	return s.getStatusFave(
		ctx,
		"AccountID,StatusID",
		func(fave *gtsmodel.StatusFave) error {
			return s.db.
				NewSelect().
				Model(fave).
				Where("status_fave.account_id = ?", accountID).
				Where("status_fave.status_id = ?", statusID).

				// Our old code actually allowed a status to
				// be faved multiple times by the same author,
				// so limit our query + order to fetch latest.
				Order("status_fave.id DESC"). // our IDs are timestamped
				Limit(1).
				Scan(ctx)
		},
		accountID,
		statusID,
	)
}

func (s *statusFaveDB) GetStatusFaveByID(ctx context.Context, id string) (*gtsmodel.StatusFave, error) {
	return s.getStatusFave(
		ctx,
		"ID",
		func(fave *gtsmodel.StatusFave) error {
			return s.db.
				NewSelect().
				Model(fave).
				Where("? = ?", bun.Ident("id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (s *statusFaveDB) GetStatusFaveByURI(ctx context.Context, uri string) (*gtsmodel.StatusFave, error) {
	return s.getStatusFave(
		ctx,
		"URI",
		func(fave *gtsmodel.StatusFave) error {
			return s.db.
				NewSelect().
				Model(fave).
				Where("? = ?", bun.Ident("uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (s *statusFaveDB) getStatusFave(ctx context.Context, lookup string, dbQuery func(*gtsmodel.StatusFave) error, keyParts ...any) (*gtsmodel.StatusFave, error) {
	// Fetch status fave from database cache with loader callback
	fave, err := s.state.Caches.DB.StatusFave.LoadOne(lookup, func() (*gtsmodel.StatusFave, error) {
		var fave gtsmodel.StatusFave

		// Not cached! Perform database query.
		if err := dbQuery(&fave); err != nil {
			return nil, err
		}

		return &fave, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return fave, nil
	}

	// Populate the status favourite model.
	if err := s.PopulateStatusFave(ctx, fave); err != nil {
		return nil, fmt.Errorf("error(s) populating status fave: %w", err)
	}

	return fave, nil
}

func (s *statusFaveDB) GetStatusFaves(ctx context.Context, statusID string) ([]*gtsmodel.StatusFave, error) {
	// Fetch the status fave IDs for status.
	faveIDs, err := s.getStatusFaveIDs(ctx, statusID)
	if err != nil {
		return nil, err
	}

	// Load all fave IDs via cache loader callbacks.
	faves, err := s.state.Caches.DB.StatusFave.LoadIDs("ID",
		faveIDs,
		func(uncached []string) ([]*gtsmodel.StatusFave, error) {
			// Preallocate expected length of uncached faves.
			faves := make([]*gtsmodel.StatusFave, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) fave IDs.
			if err := s.db.NewSelect().
				Model(&faves).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return faves, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the statuses by their
	// IDs to ensure in correct order.
	getID := func(f *gtsmodel.StatusFave) string { return f.ID }
	xslices.OrderBy(faves, faveIDs, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return faves, nil
	}

	// Populate all loaded faves, removing those we fail to
	// populate (removes needing so many nil checks everywhere).
	faves = slices.DeleteFunc(faves, func(fave *gtsmodel.StatusFave) bool {
		if err := s.PopulateStatusFave(ctx, fave); err != nil {
			log.Errorf(ctx, "error populating fave %s: %v", fave.ID, err)
			return true
		}
		return false
	})

	return faves, nil
}

func (s *statusFaveDB) IsStatusFavedBy(ctx context.Context, statusID string, accountID string) (bool, error) {
	fave, err := s.GetStatusFave(ctx, accountID, statusID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return false, err
	}
	return (fave != nil), nil
}

func (s *statusFaveDB) CountStatusFaves(ctx context.Context, statusID string) (int, error) {
	faveIDs, err := s.getStatusFaveIDs(ctx, statusID)
	return len(faveIDs), err
}

func (s *statusFaveDB) getStatusFaveIDs(ctx context.Context, statusID string) ([]string, error) {
	return s.state.Caches.DB.StatusFaveIDs.Load(statusID, func() ([]string, error) {
		var faveIDs []string

		// Status fave IDs not in cache, perform DB query!
		if err := s.db.
			NewSelect().
			Table("status_faves").
			Column("id").
			Where("? = ?", bun.Ident("status_id"), statusID).
			Scan(ctx, &faveIDs); err != nil {
			return nil, err
		}

		return faveIDs, nil
	})
}

func (s *statusFaveDB) PopulateStatusFave(ctx context.Context, statusFave *gtsmodel.StatusFave) error {
	var (
		err  error
		errs = gtserror.NewMultiError(3)
	)

	if statusFave.Account == nil {
		// StatusFave author is not set, fetch from database.
		statusFave.Account, err = s.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			statusFave.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating status fave author: %w", err)
		}
	}

	if statusFave.TargetAccount == nil {
		// StatusFave target account is not set, fetch from database.
		statusFave.TargetAccount, err = s.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			statusFave.TargetAccountID,
		)
		if err != nil {
			errs.Appendf("error populating status fave target account: %w", err)
		}
	}

	if statusFave.Status == nil {
		// StatusFave status is not set, fetch from database.
		statusFave.Status, err = s.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			statusFave.StatusID,
		)
		if err != nil {
			errs.Appendf("error populating status fave status: %w", err)
		}
	}

	return errs.Combine()
}

func (s *statusFaveDB) PutStatusFave(ctx context.Context, fave *gtsmodel.StatusFave) error {
	return s.state.Caches.DB.StatusFave.Store(fave, func() error {
		_, err := s.db.
			NewInsert().
			Model(fave).
			Exec(ctx)
		return err
	})
}

func (s *statusFaveDB) UpdateStatusFave(ctx context.Context, fave *gtsmodel.StatusFave, columns ...string) error {
	fave.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column,
		// ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	// Update the status fave model in the database.
	return s.state.Caches.DB.StatusFave.Store(fave, func() error {
		_, err := s.db.
			NewUpdate().
			Model(fave).
			Where("? = ?", bun.Ident("status_fave.id"), fave.ID).
			Column(columns...).
			Exec(ctx)
		return err
	})
}

func (s *statusFaveDB) DeleteStatusFaveByID(ctx context.Context, id string) error {
	var statusID string

	// Perform DELETE on status fave,
	// returning the status ID it was for.
	if _, err := s.db.NewDelete().
		Table("status_faves").
		Where("id = ?", id).
		Returning("status_id").
		Exec(ctx, &statusID); err != nil {
		if err == sql.ErrNoRows {
			// Not an issue, only due
			// to us doing a RETURNING.
			err = nil
		}
		return err
	}

	if statusID != "" {
		// Invalidate any cached status faves for this status.
		s.state.Caches.DB.StatusFave.Invalidate("ID", id)

		// Invalidate any cached status fave IDs for this status.
		s.state.Caches.DB.StatusFaveIDs.Invalidate(statusID)
	}

	return nil
}

func (s *statusFaveDB) DeleteStatusFaves(ctx context.Context, targetAccountID string, originAccountID string) error {
	if targetAccountID == "" && originAccountID == "" {
		return errors.New("DeleteStatusFaves: one of targetAccountID or originAccountID must be set")
	}

	var statusIDs []string

	// Prepare DELETE query returning
	// the deleted faves for status IDs.
	q := s.db.NewDelete().
		Table("status_faves").
		Returning("status_id")

	if targetAccountID != "" {
		q = q.Where("? = ?", bun.Ident("target_account_id"), targetAccountID)
	}

	if originAccountID != "" {
		q = q.Where("? = ?", bun.Ident("account_id"), originAccountID)
	}

	// Execute query, store favourited status IDs.
	if _, err := q.Exec(ctx, &statusIDs); err != nil {
		if err == sql.ErrNoRows {
			// Not an issue, only due
			// to us doing a RETURNING.
			err = nil
		}
		return err
	}

	// Deduplicate determined status IDs.
	statusIDs = xslices.Deduplicate(statusIDs)

	// Invalidate any cached status faves for this status ID.
	s.state.Caches.DB.StatusFave.InvalidateIDs("ID", statusIDs)

	// Invalidate any cached status fave IDs for this status ID.
	s.state.Caches.DB.StatusFaveIDs.Invalidate(statusIDs...)

	return nil
}

func (s *statusFaveDB) DeleteStatusFavesForStatus(ctx context.Context, statusID string) error {
	// Delete all status faves for status.
	if _, err := s.db.NewDelete().
		Table("status_faves").
		Where("status_id = ?", statusID).
		Exec(ctx); err != nil {
		return err
	}

	// Invalidate any cached status faves for this status.
	s.state.Caches.DB.StatusFave.Invalidate("ID", statusID)

	// Invalidate any cached status fave IDs for this status.
	s.state.Caches.DB.StatusFaveIDs.Invalidate(statusID)

	return nil
}
