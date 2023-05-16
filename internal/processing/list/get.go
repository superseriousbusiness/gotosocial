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

package list

import (
	"context"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Get returns the api model of one list with the given ID.
func (p *Processor) Get(ctx context.Context, account *gtsmodel.Account, id string) (*apimodel.List, gtserror.WithCode) {
	list, err := p.state.DB.GetListByID(
		// Use barebones ctx; no embedded
		// structs necessary for simple GET.
		gtscontext.SetBarebones(ctx),
		id,
	)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	if list.AccountID != account.ID {
		err = fmt.Errorf("list with id %s does not belong to account %s", list.ID, account.ID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	apiList, err := p.tc.ListToAPIList(ctx, list)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting list to api: %s", err))
	}

	return apiList, nil
}

// GetMultiple returns multiple lists created by the given account, sorted by list ID DESC (newest first).
func (p *Processor) GetMultiple(ctx context.Context, account *gtsmodel.Account, id string) ([]*apimodel.List, gtserror.WithCode) {
	lists, err := p.state.DB.GetListsForAccountID(
		// Use barebones ctx; no embedded
		// structs necessary for simple GET.
		gtscontext.SetBarebones(ctx),
		account.ID,
	)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiLists := make([]*apimodel.List, 0, len(lists))
	for _, list := range lists {
		apiList, err := p.tc.ListToAPIList(ctx, list)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting list to api: %s", err))
		}

		apiLists = append(apiLists, apiList)
	}

	return apiLists, nil
}

// // GetMultiple returns multiple reports created by the given account, filtered according to the provided parameters.
// func (p *Processor) GetMultiple(ctx context.Context, account *gtsmodel.Account) (*apimodel.PageableResponse, gtserror.WithCode) {
// 	reports, err := p.state.DB.GetReports(ctx, resolved, account.ID, targetAccountID, maxID, sinceID, minID, limit)
// 	if err != nil {
// 		if err == db.ErrNoEntries {
// 			return util.EmptyPageableResponse(), nil
// 		}
// 		return nil, gtserror.NewErrorInternalError(err)
// 	}

// 	count := len(reports)
// 	items := make([]interface{}, 0, count)
// 	nextMaxIDValue := ""
// 	prevMinIDValue := ""
// 	for i, r := range reports {
// 		item, err := p.tc.ReportToAPIReport(ctx, r)
// 		if err != nil {
// 			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting report to api: %s", err))
// 		}

// 		if i == count-1 {
// 			nextMaxIDValue = item.ID
// 		}

// 		if i == 0 {
// 			prevMinIDValue = item.ID
// 		}

// 		items = append(items, item)
// 	}

// 	extraQueryParams := []string{}
// 	if resolved != nil {
// 		extraQueryParams = append(extraQueryParams, "resolved="+strconv.FormatBool(*resolved))
// 	}
// 	if targetAccountID != "" {
// 		extraQueryParams = append(extraQueryParams, "target_account_id="+targetAccountID)
// 	}

// 	return util.PackagePageableResponse(util.PageableResponseParams{
// 		Items:            items,
// 		Path:             "/api/v1/reports",
// 		NextMaxIDValue:   nextMaxIDValue,
// 		PrevMinIDValue:   prevMinIDValue,
// 		Limit:            limit,
// 		ExtraQueryParams: extraQueryParams,
// 	})
// }
