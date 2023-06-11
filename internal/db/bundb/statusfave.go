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
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type statusFaveDB struct {
	conn  *DBConn
	state *state.State
}

func (s *statusFaveDB) GetStatusFave(ctx context.Context, accountID string, statusID string) (*gtsmodel.StatusFave, db.Error) {
	return s.getStatusFave(
		ctx,
		"AccountID.StatusID",
		func(fave *gtsmodel.StatusFave) error {
			return s.conn.
				NewSelect().
				Model(fave).
				Where("? = ?", bun.Ident("account_id"), accountID).
				Where("? = ?", bun.Ident("status_id"), statusID).
				Scan(ctx)
		},
		accountID,
		statusID,
	)
}

func (s *statusFaveDB) GetStatusFaveByID(ctx context.Context, id string) (*gtsmodel.StatusFave, db.Error) {
	return s.getStatusFave(
		ctx,
		"ID",
		func(fave *gtsmodel.StatusFave) error {
			return s.conn.
				NewSelect().
				Model(fave).
				Where("? = ?", bun.Ident("id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (s *statusFaveDB) getStatusFave(ctx context.Context, lookup string, dbQuery func(*gtsmodel.StatusFave) error, keyParts ...any) (*gtsmodel.StatusFave, error) {
	// Fetch status fave from database cache with loader callback
	fave, err := s.state.Caches.GTS.StatusFave().Load(lookup, func() (*gtsmodel.StatusFave, error) {
		var fave gtsmodel.StatusFave

		// Not cached! Perform database query.
		if err := dbQuery(&fave); err != nil {
			return nil, s.conn.ProcessError(err)
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

	// Fetch the status fave author account.
	fave.Account, err = s.state.DB.GetAccountByID(
		gtscontext.SetBarebones(ctx),
		fave.AccountID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting status fave account %q: %w", fave.AccountID, err)
	}

	// Fetch the status fave target account.
	fave.TargetAccount, err = s.state.DB.GetAccountByID(
		gtscontext.SetBarebones(ctx),
		fave.TargetAccountID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting status fave target account %q: %w", fave.TargetAccountID, err)
	}

	// Fetch the status fave target status.
	fave.Status, err = s.state.DB.GetStatusByID(
		gtscontext.SetBarebones(ctx),
		fave.StatusID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting status fave status %q: %w", fave.StatusID, err)
	}

	return fave, nil
}

func (s *statusFaveDB) GetStatusFavesForStatus(ctx context.Context, statusID string) ([]*gtsmodel.StatusFave, db.Error) {
	ids := []string{}

	if err := s.conn.
		NewSelect().
		Table("status_faves").
		Column("id").
		Where("? = ?", bun.Ident("status_id"), statusID).
		Scan(ctx, &ids); err != nil {
		return nil, s.conn.ProcessError(err)
	}

	faves := make([]*gtsmodel.StatusFave, 0, len(ids))

	for _, id := range ids {
		fave, err := s.GetStatusFaveByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting status fave %q: %v", id, err)
			continue
		}

		faves = append(faves, fave)
	}

	return faves, nil
}

func (s *statusFaveDB) PopulateStatusFave(ctx context.Context, statusFave *gtsmodel.StatusFave) error {
	var (
		err  error
		errs = make(gtserror.MultiError, 0, 3)
	)

	if statusFave.Account == nil {
		// StatusFave author is not set, fetch from database.
		statusFave.Account, err = s.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			statusFave.AccountID,
		)
		if err != nil {
			errs.Append(fmt.Errorf("error populating status fave author: %w", err))
		}
	}

	if statusFave.TargetAccount == nil {
		// StatusFave target account is not set, fetch from database.
		statusFave.TargetAccount, err = s.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			statusFave.TargetAccountID,
		)
		if err != nil {
			errs.Append(fmt.Errorf("error populating status fave target account: %w", err))
		}
	}

	if statusFave.Status == nil {
		// StatusFave status is not set, fetch from database.
		statusFave.Status, err = s.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			statusFave.StatusID,
		)
		if err != nil {
			errs.Append(fmt.Errorf("error populating status fave status: %w", err))
		}
	}

	return errs.Combine()
}

func (s *statusFaveDB) PutStatusFave(ctx context.Context, fave *gtsmodel.StatusFave) db.Error {
	return s.state.Caches.GTS.StatusFave().Store(fave, func() error {
		_, err := s.conn.
			NewInsert().
			Model(fave).
			Exec(ctx)
		return s.conn.ProcessError(err)
	})
}

func (s *statusFaveDB) DeleteStatusFaveByID(ctx context.Context, id string) db.Error {
	defer s.state.Caches.GTS.StatusFave().Invalidate("ID", id)

	// Load fave into cache before attempting a delete,
	// as we need it cached in order to trigger the invalidate
	// callback. This in turn invalidates others.
	_, err := s.GetStatusFaveByID(gtscontext.SetBarebones(ctx), id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// not an issue.
			err = nil
		}
		return err
	}

	// Finally delete fave from DB.
	_, err = s.conn.NewDelete().
		Table("status_faves").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx)
	return s.conn.ProcessError(err)
}

func (s *statusFaveDB) DeleteStatusFaves(ctx context.Context, targetAccountID string, originAccountID string) db.Error {
	if targetAccountID == "" && originAccountID == "" {
		return errors.New("DeleteStatusFaves: one of targetAccountID or originAccountID must be set")
	}

	var faveIDs []string

	q := s.conn.
		NewSelect().
		Column("id").
		Table("status_faves")

	if targetAccountID != "" {
		q = q.Where("? = ?", bun.Ident("target_account_id"), targetAccountID)
	}

	if originAccountID != "" {
		q = q.Where("? = ?", bun.Ident("account_id"), originAccountID)
	}

	if _, err := q.Exec(ctx, &faveIDs); err != nil {
		return s.conn.ProcessError(err)
	}

	defer func() {
		// Invalidate all IDs on return.
		for _, id := range faveIDs {
			s.state.Caches.GTS.StatusFave().Invalidate("ID", id)
		}
	}()

	// Load all faves into cache, this *really* isn't great
	// but it is the only way we can ensure we invalidate all
	// related caches correctly (e.g. visibility).
	for _, id := range faveIDs {
		_, err := s.GetStatusFaveByID(ctx, id)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return err
		}
	}

	// Finally delete all from DB.
	_, err := s.conn.NewDelete().
		Table("status_faves").
		Where("? IN (?)", bun.Ident("id"), bun.In(faveIDs)).
		Exec(ctx)
	return s.conn.ProcessError(err)
}

func (s *statusFaveDB) DeleteStatusFavesForStatus(ctx context.Context, statusID string) db.Error {
	// Capture fave IDs in a RETURNING statement.
	var faveIDs []string

	q := s.conn.
		NewSelect().
		Column("id").
		Table("status_faves").
		Where("? = ?", bun.Ident("status_id"), statusID)
	if _, err := q.Exec(ctx, &faveIDs); err != nil {
		return s.conn.ProcessError(err)
	}

	defer func() {
		// Invalidate all IDs on return.
		for _, id := range faveIDs {
			s.state.Caches.GTS.StatusFave().Invalidate("ID", id)
		}
	}()

	// Load all faves into cache, this *really* isn't great
	// but it is the only way we can ensure we invalidate all
	// related caches correctly (e.g. visibility).
	for _, id := range faveIDs {
		_, err := s.GetStatusFaveByID(ctx, id)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return err
		}
	}

	// Finally delete all from DB.
	_, err := s.conn.NewDelete().
		Table("status_faves").
		Where("? IN (?)", bun.Ident("id"), bun.In(faveIDs)).
		Exec(ctx)
	return s.conn.ProcessError(err)
}
