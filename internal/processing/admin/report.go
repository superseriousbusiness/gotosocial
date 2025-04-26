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

package admin

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// ReportsGet returns reports stored on this
// instance, with the given parameters.
func (p *Processor) ReportsGet(
	ctx context.Context,
	account *gtsmodel.Account,
	resolved *bool,
	accountID string,
	targetAccountID string,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	reports, err := p.state.DB.GetReports(
		ctx,
		resolved,
		accountID,
		targetAccountID,
		page,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(reports)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := reports[count-1].ID
	hi := reports[0].ID

	// Convert each report to API model.
	items := make([]interface{}, 0, count)
	for _, r := range reports {
		item, err := p.converter.ReportToAdminAPIReport(ctx, r, account)
		if err != nil {
			err := fmt.Errorf("error converting report to api: %s", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		items = append(items, item)
	}

	// Assemble next/prev page queries.
	query := make(url.Values, 3)
	if resolved != nil {
		query.Set(apiutil.ResolvedKey, strconv.FormatBool(*resolved))
	}
	if accountID != "" {
		query.Set(apiutil.AccountIDKey, accountID)
	}
	if targetAccountID != "" {
		query.Set(apiutil.TargetAccountIDKey, targetAccountID)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/admin/reports",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: query,
	}), nil
}

// ReportGet returns one report, with the given ID.
func (p *Processor) ReportGet(ctx context.Context, account *gtsmodel.Account, id string) (*apimodel.AdminReport, gtserror.WithCode) {
	report, err := p.state.DB.GetReportByID(ctx, id)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	apimodelReport, err := p.converter.ReportToAdminAPIReport(ctx, report, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apimodelReport, nil
}

// ReportResolve marks a report with the given id as resolved,
// and stores the provided actionTakenComment (if not null).
// If the report creator is from this instance, an email will
// be sent to them to let them know that the report is resolved.
func (p *Processor) ReportResolve(ctx context.Context, account *gtsmodel.Account, id string, actionTakenComment *string) (*apimodel.AdminReport, gtserror.WithCode) {
	report, err := p.state.DB.GetReportByID(ctx, id)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	columns := []string{
		"action_taken_at",
		"action_taken_by_account_id",
	}

	report.ActionTakenAt = time.Now()
	report.ActionTakenByAccountID = account.ID

	if actionTakenComment != nil {
		report.ActionTaken = *actionTakenComment
		columns = append(columns, "action_taken")
	}

	err = p.state.DB.UpdateReport(ctx, report, columns...)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Process side effects of closing the report.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActivityFlag,
		APActivityType: ap.ActivityUpdate,
		GTSModel:       report,
		Origin:         account,
		Target:         report.Account,
	})

	apimodelReport, err := p.converter.ReportToAdminAPIReport(ctx, report, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apimodelReport, nil
}
