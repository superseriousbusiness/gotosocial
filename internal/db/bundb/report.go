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
	"github.com/uptrace/bun"
)

type reportDB struct {
	db    *bun.DB
	state *state.State
}

func (r *reportDB) newReportQ(report interface{}) *bun.SelectQuery {
	return r.db.NewSelect().Model(report)
}

func (r *reportDB) GetReportByID(ctx context.Context, id string) (*gtsmodel.Report, error) {
	return r.getReport(
		ctx,
		"ID",
		func(report *gtsmodel.Report) error {
			return r.newReportQ(report).Where("? = ?", bun.Ident("report.id"), id).Scan(ctx)
		},
		id,
	)
}

func (r *reportDB) GetReports(ctx context.Context, resolved *bool, accountID string, targetAccountID string, page *paging.Page) ([]*gtsmodel.Report, error) {
	var (
		// Get paging params.
		minID = page.GetMin()
		maxID = page.GetMax()
		limit = page.GetLimit()
		order = page.GetOrder()

		// Make educated guess for slice size
		reportIDs = make([]string, 0, limit)
	)

	q := r.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("reports"), bun.Ident("report")).
		// Select only IDs from table.
		Column("report.id")

	if resolved != nil {
		i := bun.Ident("report.action_taken_by_account_id")
		if *resolved {
			q = q.Where("? IS NOT NULL", i)
		} else {
			q = q.Where("? IS NULL", i)
		}
	}

	if accountID != "" {
		q = q.Where("? = ?", bun.Ident("report.account_id"), accountID)
	}

	if targetAccountID != "" {
		q = q.Where("? = ?", bun.Ident("report.target_account_id"), targetAccountID)
	}

	// Return only reports with id
	// lower than provided maxID.
	if maxID != "" {
		q = q.Where("? < ?", bun.Ident("report.id"), maxID)
	}

	// Return only reports with id
	// greater than provided minID.
	if minID != "" {
		q = q.Where("? > ?", bun.Ident("report.id"), minID)
	}

	if limit > 0 {
		// Limit amount of
		// reports returned.
		q = q.Limit(limit)
	}

	if order == paging.OrderAscending {
		// Page up.
		q = q.OrderExpr("? ASC", bun.Ident("report.id"))
	} else {
		// Page down.
		q = q.OrderExpr("? DESC", bun.Ident("report.id"))
	}

	if err := q.Scan(ctx, &reportIDs); err != nil {
		return nil, err
	}

	// Catch case of no reports early
	if len(reportIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	// If we're paging up, we still want reports
	// to be sorted by ID desc, so reverse ids slice.
	if order == paging.OrderAscending {
		slices.Reverse(reportIDs)
	}

	// Allocate return slice (will be at most len reportIDs)
	reports := make([]*gtsmodel.Report, 0, len(reportIDs))
	for _, id := range reportIDs {
		report, err := r.GetReportByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting report %q: %v", id, err)
			continue
		}

		// Append to return slice
		reports = append(reports, report)
	}

	return reports, nil
}

func (r *reportDB) getReport(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Report) error, keyParts ...any) (*gtsmodel.Report, error) {
	// Fetch report from database cache with loader callback
	report, err := r.state.Caches.DB.Report.LoadOne(lookup, func() (*gtsmodel.Report, error) {
		var report gtsmodel.Report

		// Not cached! Perform database query
		if err := dbQuery(&report); err != nil {
			return nil, err
		}

		return &report, nil
	}, keyParts...)
	if err != nil {
		// error already processed
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return report, nil
	}

	if err := r.state.DB.PopulateReport(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

func (r *reportDB) PopulateReport(ctx context.Context, report *gtsmodel.Report) error {
	var (
		err  error
		errs = gtserror.NewMultiError(4)
	)

	if report.Account == nil {
		// Report account is not set, fetch from the database.
		report.Account, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			report.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating report account: %w", err)
		}
	}

	if report.TargetAccount == nil {
		// Report target account is not set, fetch from the database.
		report.TargetAccount, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			report.TargetAccountID,
		)
		if err != nil {
			errs.Appendf("error populating report target account: %w", err)
		}
	}

	if l := len(report.StatusIDs); l > 0 && l != len(report.Statuses) {
		// Report target statuses not set, fetch from the database.
		report.Statuses, err = r.state.DB.GetStatusesByIDs(
			gtscontext.SetBarebones(ctx),
			report.StatusIDs,
		)
		if err != nil {
			errs.Appendf("error populating report statuses: %w", err)
		}
	}

	if l := len(report.RuleIDs); l > 0 && l != len(report.Rules) {
		// Report target rules not set, fetch from the database.

		for _, v := range report.RuleIDs {
			rule, err := r.state.DB.GetRuleByID(ctx, v)
			if err != nil {
				errs.Appendf("error populating report rules: %w", err)
			} else {
				report.Rules = append(report.Rules, rule)
			}
		}
	}

	if report.ActionTakenByAccountID != "" &&
		report.ActionTakenByAccount == nil {
		// Report action account is not set, fetch from the database.
		report.ActionTakenByAccount, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			report.ActionTakenByAccountID,
		)
		if err != nil {
			errs.Appendf("error populating report action taken by account: %w", err)
		}
	}

	return errs.Combine()
}

func (r *reportDB) PutReport(ctx context.Context, report *gtsmodel.Report) error {
	return r.state.Caches.DB.Report.Store(report, func() error {
		_, err := r.db.NewInsert().Model(report).Exec(ctx)
		return err
	})
}

func (r *reportDB) UpdateReport(ctx context.Context, report *gtsmodel.Report, columns ...string) error {
	// Update the report's last-updated
	report.UpdatedAt = time.Now()
	if len(columns) != 0 {
		columns = append(columns, "updated_at")
	}

	return r.state.Caches.DB.Report.Store(report, func() error {
		_, err := r.db.
			NewUpdate().
			Model(report).
			Where("? = ?", bun.Ident("report.id"), report.ID).
			Column(columns...).
			Exec(ctx)
		return err
	})
}

func (r *reportDB) DeleteReportByID(ctx context.Context, id string) error {
	// Delete the report from DB.
	if _, err := r.db.NewDelete().
		TableExpr("? AS ?", bun.Ident("reports"), bun.Ident("report")).
		Where("? = ?", bun.Ident("report.id"), id).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate any cached report model by ID.
	r.state.Caches.DB.Report.Invalidate("ID", id)

	return nil
}
