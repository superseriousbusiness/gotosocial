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

package report

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// Get returns the user view of a moderation report, with the given id.
func (p *Processor) Get(ctx context.Context, account *gtsmodel.Account, id string) (*apimodel.Report, gtserror.WithCode) {
	report, err := p.state.DB.GetReportByID(ctx, id)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	if report.AccountID != account.ID {
		err = fmt.Errorf("report with id %s does not belong to account %s", report.ID, account.ID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	apiReport, err := p.converter.ReportToAPIReport(ctx, report)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting report to api: %s", err))
	}

	return apiReport, nil
}

// GetMultiple returns reports created by the given account,
// filtered according to the provided parameters.
func (p *Processor) GetMultiple(
	ctx context.Context,
	account *gtsmodel.Account,
	resolved *bool,
	targetAccountID string,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	reports, err := p.state.DB.GetReports(
		ctx,
		resolved,
		account.ID,
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
		item, err := p.converter.ReportToAPIReport(ctx, r)
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
	if targetAccountID != "" {
		query.Set(apiutil.TargetAccountIDKey, targetAccountID)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/reports",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: query,
	}), nil
}
