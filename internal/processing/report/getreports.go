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

package report

import (
	"context"
	"fmt"
	"strconv"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) ReportsGet(ctx context.Context, account *gtsmodel.Account, resolved *bool, targetAccountID string, maxID string, sinceID string, minID string, limit int) (*apimodel.PageableResponse, gtserror.WithCode) {
	reports, err := p.db.GetReports(ctx, resolved, account.ID, targetAccountID, maxID, sinceID, minID, limit)
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
		item, err := p.tc.ReportToAPIReport(ctx, r)
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
	if targetAccountID != "" {
		extraQueryParams = append(extraQueryParams, "target_account_id="+targetAccountID)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:            items,
		Path:             "/api/v1/reports",
		NextMaxIDValue:   nextMaxIDValue,
		PrevMinIDValue:   prevMinIDValue,
		Limit:            limit,
		ExtraQueryParams: extraQueryParams,
	})
}
