/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package bundb

import (
	"context"
	"fmt"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type reportDB struct {
	conn  *DBConn
	state *state.State
}

func (r *reportDB) newReportQ(report interface{}) *bun.SelectQuery {
	return r.conn.NewSelect().Model(report)
}

func (r *reportDB) GetReportByID(ctx context.Context, id string) (*gtsmodel.Report, db.Error) {
	return r.getReport(
		ctx,
		"ID",
		func(report *gtsmodel.Report) error {
			return r.newReportQ(report).Where("? = ?", bun.Ident("report.id"), id).Scan(ctx)
		},
		id,
	)
}

func (r *reportDB) getReport(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Report) error, keyParts ...any) (*gtsmodel.Report, db.Error) {
	// Fetch report from database cache with loader callback
	report, err := r.state.Caches.GTS.Report().Load(lookup, func() (*gtsmodel.Report, error) {
		var report gtsmodel.Report

		// Not cached! Perform database query
		if err := dbQuery(&report); err != nil {
			return nil, r.conn.ProcessError(err)
		}

		return &report, nil
	}, keyParts...)
	if err != nil {
		// error already processed
		return nil, err
	}

	// Set the report author account
	report.Account, err = r.state.DB.GetAccountByID(ctx, report.AccountID)
	if err != nil {
		return nil, fmt.Errorf("error getting report account: %w", err)
	}

	// Set the report target account
	report.TargetAccount, err = r.state.DB.GetAccountByID(ctx, report.TargetAccountID)
	if err != nil {
		return nil, fmt.Errorf("error getting report target account: %w", err)
	}

	if len(report.StatusIDs) > 0 {
		// Fetch reported statuses
		report.Statuses, err = r.state.DB.GetStatuses(ctx, report.StatusIDs)
		if err != nil {
			return nil, fmt.Errorf("error getting status mentions: %w", err)
		}
	}

	if report.ActionTakenByAccountID != "" {
		// Set the report action taken by account
		report.ActionTakenByAccount, err = r.state.DB.GetAccountByID(ctx, report.ActionTakenByAccountID)
		if err != nil {
			return nil, fmt.Errorf("error getting report action taken by account: %w", err)
		}
	}

	return report, nil
}

func (r *reportDB) PutReport(ctx context.Context, report *gtsmodel.Report) db.Error {
	return r.state.Caches.GTS.Report().Store(report, func() error {
		_, err := r.conn.NewInsert().Model(report).Exec(ctx)
		return r.conn.ProcessError(err)
	})
}

func (r *reportDB) UpdateReport(ctx context.Context, report *gtsmodel.Report, columns ...string) (*gtsmodel.Report, db.Error) {
	// Update the report's last-updated
	report.UpdatedAt = time.Now()
	if len(columns) != 0 {
		columns = append(columns, "updated_at")
	}

	if _, err := r.conn.
		NewUpdate().
		Model(report).
		Where("? = ?", bun.Ident("report.id"), report.ID).
		Column(columns...).
		Exec(ctx); err != nil {
		return nil, r.conn.ProcessError(err)
	}

	r.state.Caches.GTS.Report().Invalidate("ID", report.ID)
	return report, nil
}

func (r *reportDB) DeleteReportByID(ctx context.Context, id string) db.Error {
	if _, err := r.conn.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("reports"), bun.Ident("report")).
		Where("? = ?", bun.Ident("report.id"), id).
		Exec(ctx); err != nil {
		return r.conn.ProcessError(err)
	}

	r.state.Caches.GTS.Report().Invalidate("ID", id)
	return nil
}
