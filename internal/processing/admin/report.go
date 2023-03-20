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
	"fmt"
	"strconv"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// ReportsGet returns all reports stored on this instance, with the given parameters.
func (p *Processor) ReportsGet(
	ctx context.Context,
	account *gtsmodel.Account,
	resolved *bool,
	accountID string,
	targetAccountID string,
	maxID string,
	sinceID string,
	minID string,
	limit int,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	reports, err := p.state.DB.GetReports(ctx, resolved, accountID, targetAccountID, maxID, sinceID, minID, limit)
	if err != nil {
		if err == db.ErrNoEntries {
			return util.EmptyPageableResponse(), nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(reports)
	items := make([]interface{}, 0, count)
	nextMaxIDValue := ""
	prevMinIDValue := ""
	for i, r := range reports {
		item, err := p.tc.ReportToAdminAPIReport(ctx, r, account)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting report to api: %s", err))
		}

		if i == count-1 {
			nextMaxIDValue = item.ID
		}

		if i == 0 {
			prevMinIDValue = item.ID
		}

		items = append(items, item)
	}

	extraQueryParams := []string{}
	if resolved != nil {
		extraQueryParams = append(extraQueryParams, "resolved="+strconv.FormatBool(*resolved))
	}
	if accountID != "" {
		extraQueryParams = append(extraQueryParams, "account_id="+accountID)
	}
	if targetAccountID != "" {
		extraQueryParams = append(extraQueryParams, "target_account_id="+targetAccountID)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:            items,
		Path:             "/api/v1/admin/reports",
		NextMaxIDValue:   nextMaxIDValue,
		PrevMinIDValue:   prevMinIDValue,
		Limit:            limit,
		ExtraQueryParams: extraQueryParams,
	})
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

	apimodelReport, err := p.tc.ReportToAdminAPIReport(ctx, report, account)
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

	updatedReport, err := p.state.DB.UpdateReport(ctx, report, columns...)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Process side effects of closing the report.
	p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ActivityFlag,
		APActivityType: ap.ActivityUpdate,
		GTSModel:       report,
		OriginAccount:  account,
		TargetAccount:  report.Account,
	})

	apimodelReport, err := p.tc.ReportToAdminAPIReport(ctx, updatedReport, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apimodelReport, nil
}
