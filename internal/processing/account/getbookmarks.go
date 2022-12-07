/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package account

import (
	"context"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) BookmarksGet(ctx context.Context, requestingAccount *gtsmodel.Account, limit int, maxID string, minID string) (*apimodel.PageableResponse, gtserror.WithCode) {
	if requestingAccount == nil {
		return nil, gtserror.NewErrorForbidden(fmt.Errorf("cannot retrieve bookmarks without a requesting account"))
	}

	statuses, err := p.db.GetBookmarks(ctx, requestingAccount.ID, limit, maxID, minID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	filtered := make([]*gtsmodel.Status, 0, len(statuses))
	for _, s := range statuses {
		visible, err := p.filter.StatusVisible(ctx, s, requestingAccount)
		if err == nil && visible {
			filtered = append(filtered, s)
		}
	}

	count := len(filtered)

	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	items := []interface{}{}
	nextMaxIDValue := ""
	prevMinIDValue := ""
	for i, s := range filtered {
		item, err := p.tc.StatusToAPIStatus(ctx, s, requestingAccount)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status to api: %s", err))
		}

		if i == count-1 {
			nextMaxIDValue = item.GetID()
		}

		if i == 0 {
			prevMinIDValue = item.GetID()
		}

		items = append(items, item)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:            items,
		Path:             "/api/v1/bookmarks",
		NextMaxIDValue:   nextMaxIDValue,
		PrevMinIDValue:   prevMinIDValue,
		Limit:            limit,
		ExtraQueryParams: []string{},
	})
}
