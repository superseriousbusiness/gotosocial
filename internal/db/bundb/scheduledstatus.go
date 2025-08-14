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

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

type scheduledStatusDB struct {
	db    *bun.DB
	state *state.State
}

func (s *scheduledStatusDB) GetAllScheduledStatuses(ctx context.Context) ([]*gtsmodel.ScheduledStatus, error) {

	var statusIDs []string

	// Select ALL token IDs.
	if err := s.db.NewSelect().
		Table("scheduled_statuses").
		Column("id").
		Scan(ctx, &statusIDs); err != nil {
		return nil, err
	}

	return s.GetScheduledStatusesByIDs(ctx, statusIDs)
}

func (s *scheduledStatusDB) GetScheduledStatusByID(ctx context.Context, id string) (*gtsmodel.ScheduledStatus, error) {
	return s.getScheduledStatus(
		ctx,
		"ID",
		func(scheduledStatus *gtsmodel.ScheduledStatus) error {
			return s.db.
				NewSelect().
				Model(scheduledStatus).
				Where("? = ?", bun.Ident("scheduled_status.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (s *scheduledStatusDB) getScheduledStatus(
	ctx context.Context,
	lookup string,
	dbQuery func(*gtsmodel.ScheduledStatus) error,
	keyParts ...any,
) (*gtsmodel.ScheduledStatus, error) {
	// Fetch scheduled status from database cache with loader callback
	scheduledStatus, err := s.state.Caches.DB.ScheduledStatus.LoadOne(lookup, func() (*gtsmodel.ScheduledStatus, error) {
		var scheduledStatus gtsmodel.ScheduledStatus

		// Not cached! Perform database query
		if err := dbQuery(&scheduledStatus); err != nil {
			return nil, err
		}

		return &scheduledStatus, nil
	}, keyParts...)
	if err != nil {
		// Error already processed.
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return scheduledStatus, nil
	}

	if err := s.PopulateScheduledStatus(ctx, scheduledStatus); err != nil {
		return nil, err
	}

	return scheduledStatus, nil
}

func (s *scheduledStatusDB) PopulateScheduledStatus(ctx context.Context, status *gtsmodel.ScheduledStatus) error {
	var (
		err  error
		errs = gtserror.NewMultiError(1)
	)

	if status.Account == nil {
		status.Account, err = s.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			status.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating scheduled status author account: %w", err)
		}
	}

	if status.Application == nil {
		status.Application, err = s.state.DB.GetApplicationByID(
			gtscontext.SetBarebones(ctx),
			status.ApplicationID,
		)
		if err != nil {
			errs.Appendf("error populating scheduled status application: %w", err)
		}
	}

	if !status.AttachmentsPopulated() {
		// Status attachments are out-of-date with IDs, repopulate.
		status.MediaAttachments, err = s.state.DB.GetAttachmentsByIDs(
			gtscontext.SetBarebones(ctx),
			status.MediaIDs,
		)
		if err != nil {
			errs.Appendf("error populating status attachments: %w", err)
		}
	}

	return errs.Combine()
}

func (s *scheduledStatusDB) GetScheduledStatusesByIDs(ctx context.Context, ids []string) ([]*gtsmodel.ScheduledStatus, error) {
	// Load all scheduled status IDs via cache loader callbacks.
	statuses, err := s.state.Caches.DB.ScheduledStatus.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.ScheduledStatus, error) {
			// Preallocate expected length of uncached scheduled statuses.
			statuses := make([]*gtsmodel.ScheduledStatus, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := s.db.NewSelect().
				Model(&statuses).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return statuses, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the statuses by their
	// IDs to ensure in correct order.
	getID := func(r *gtsmodel.ScheduledStatus) string { return r.ID }
	xslices.OrderBy(statuses, ids, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return statuses, nil
	}

	// Populate all loaded scheduled statuses, removing those we
	// fail to populate (removes needing so many nil checks everywhere).
	statuses = slices.DeleteFunc(statuses, func(scheduledStatus *gtsmodel.ScheduledStatus) bool {
		if err := s.PopulateScheduledStatus(ctx, scheduledStatus); err != nil {
			log.Errorf(ctx, "error populating %s: %v", scheduledStatus.ID, err)
			return true
		}
		return false
	})

	return statuses, nil
}

func (s *scheduledStatusDB) GetScheduledStatusesForAcct(
	ctx context.Context,
	acctID string,
	page *paging.Page,
) ([]*gtsmodel.ScheduledStatus, error) {
	var (
		// Get paging params.
		minID = page.GetMin()
		maxID = page.GetMax()
		limit = page.GetLimit()
		order = page.GetOrder()

		// Make educated guess for slice size
		statusIDs = make([]string, 0, limit)
	)

	// Create the basic select query.
	q := s.db.
		NewSelect().
		Column("id").
		TableExpr(
			"? AS ?",
			bun.Ident("scheduled_statuses"),
			bun.Ident("scheduled_status"),
		)

	// Select scheduled statuses by the account.
	if acctID != "" {
		q = q.Where("? = ?", bun.Ident("account_id"), acctID)
	}

	// Add paging param max ID.
	if maxID != "" {
		q = q.Where("? < ?", bun.Ident("id"), maxID)
	}

	// Add paging param min ID.
	if minID != "" {
		q = q.Where("? > ?", bun.Ident("id"), minID)
	}

	// Add paging param order.
	if order == paging.OrderAscending {
		// Page up.
		q = q.OrderExpr("? ASC", bun.Ident("id"))
	} else {
		// Page down.
		q = q.OrderExpr("? DESC", bun.Ident("id"))
	}

	// Add paging param limit.
	if limit > 0 {
		q = q.Limit(limit)
	}

	// Execute the query and scan into IDs.
	err := q.Scan(ctx, &statusIDs)
	if err != nil {
		return nil, err
	}

	// Catch case of no items early
	if len(statusIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	// If we're paging up, we still want statuses
	// to be sorted by ID desc, so reverse ids slice.
	if order == paging.OrderAscending {
		slices.Reverse(statusIDs)
	}

	// Load all scheduled statuses by their IDs.
	return s.GetScheduledStatusesByIDs(ctx, statusIDs)
}

func (s *scheduledStatusDB) PutScheduledStatus(ctx context.Context, status *gtsmodel.ScheduledStatus) error {
	return s.state.Caches.DB.ScheduledStatus.Store(status, func() error {
		return s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			if _, err := tx.NewInsert().
				Model(status).
				Exec(ctx); err != nil {
				return gtserror.Newf("error inserting scheduled status: %w", err)
			}

			// change the scheduled status ID of the
			// media attachments to the current status
			for _, a := range status.MediaAttachments {
				a.ScheduledStatusID = status.ID
				if _, err := tx.
					NewUpdate().
					Model(a).
					Column("scheduled_status_id").
					Where("? = ?", bun.Ident("media_attachment.id"), a.ID).
					Exec(ctx); err != nil {
					return gtserror.Newf("error updating media: %w", err)
				}
			}

			return nil
		})
	})
}

func (s *scheduledStatusDB) DeleteScheduledStatusByID(ctx context.Context, id string) error {
	var deleted gtsmodel.ScheduledStatus

	// Delete scheduled status
	// from database by its ID.
	if _, err := s.db.NewDelete().
		Model(&deleted).
		Returning("?, ?", bun.Ident("id"), bun.Ident("attachments")).
		Where("? = ?", bun.Ident("scheduled_status.id"), id).
		Exec(ctx); err != nil {
		return err
	}

	// Invalidate cached scheduled status by its ID,
	// manually call invalidate hook in case not cached.
	s.state.Caches.DB.ScheduledStatus.Invalidate("ID", id)
	s.state.Caches.OnInvalidateScheduledStatus(&deleted)

	return nil
}

func (s *scheduledStatusDB) DeleteScheduledStatusesByAccountID(ctx context.Context, accountID string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted []*gtsmodel.ScheduledStatus

	if _, err := s.db.NewDelete().
		Model(&deleted).
		Returning("?, ?", bun.Ident("id"), bun.Ident("attachments")).
		Where("? = ?", bun.Ident("account_id"), accountID).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	for _, deleted := range deleted {
		// Invalidate cached scheduled statuses by ID
		// and related media attachments.
		s.state.Caches.DB.ScheduledStatus.Invalidate("ID", deleted.ID)
		s.state.Caches.OnInvalidateScheduledStatus(deleted)
	}

	return nil
}

func (s *scheduledStatusDB) DeleteScheduledStatusesByApplicationID(ctx context.Context, applicationID string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted []*gtsmodel.ScheduledStatus

	if _, err := s.db.NewDelete().
		Model(&deleted).
		Returning("?, ?", bun.Ident("id"), bun.Ident("attachments")).
		Where("? = ?", bun.Ident("application_id"), applicationID).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	for _, deleted := range deleted {
		// Invalidate cached scheduled statuses by ID
		// and related media attachments.
		s.state.Caches.DB.ScheduledStatus.Invalidate("ID", deleted.ID)
		s.state.Caches.OnInvalidateScheduledStatus(deleted)
	}

	return nil
}

func (s *scheduledStatusDB) UpdateScheduledStatusScheduledDate(ctx context.Context, scheduledStatus *gtsmodel.ScheduledStatus, scheduledAt *time.Time) error {
	return s.state.Caches.DB.ScheduledStatus.Store(scheduledStatus, func() error {
		_, err := s.db.NewUpdate().
			Model(scheduledStatus).
			Where("? = ?", bun.Ident("scheduled_status.id"), scheduledStatus.ID).
			Column("scheduled_at").
			Exec(ctx)
		return err
	})
}

func (s *scheduledStatusDB) GetScheduledStatusesCountForAcct(ctx context.Context, acctID string, scheduledAt *time.Time) (int, error) {
	q := s.db.
		NewSelect().
		Column("id").
		TableExpr(
			"? AS ?",
			bun.Ident("scheduled_statuses"),
			bun.Ident("scheduled_status"),
		).
		Where("? = ?", bun.Ident("account_id"), acctID)

	if scheduledAt != nil {
		startOfDay := time.Date(scheduledAt.Year(), scheduledAt.Month(), scheduledAt.Day(), 0, 0, 0, 0, scheduledAt.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)
		q = q.
			Where("? >= ? AND ? < ?", bun.Ident("scheduled_at"), startOfDay, bun.Ident("scheduled_at"), endOfDay)
	}

	count, err := q.Count(ctx)

	if err != nil {
		return 0, err
	}

	return count, nil
}
