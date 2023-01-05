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

package account

import (
	"context"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) StatusesGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string, limit int, excludeReplies bool, excludeReblogs bool, maxID string, minID string, pinnedOnly bool, mediaOnly bool, publicOnly bool) (*apimodel.PageableResponse, gtserror.WithCode) {
	if requestingAccount != nil {
		if blocked, err := p.db.IsBlocked(ctx, requestingAccount.ID, targetAccountID, true); err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		} else if blocked {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("block exists between accounts"))
		}
	}

	statuses, err := p.db.GetAccountStatuses(ctx, targetAccountID, limit, excludeReplies, excludeReblogs, maxID, minID, pinnedOnly, mediaOnly, publicOnly)
	if err != nil {
		if err == db.ErrNoEntries {
			return util.EmptyPageableResponse(), nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	var filtered []*gtsmodel.Status
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
		Items:          items,
		Path:           fmt.Sprintf("/api/v1/accounts/%s/statuses", targetAccountID),
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
		ExtraQueryParams: []string{
			fmt.Sprintf("exclude_replies=%t", excludeReplies),
			fmt.Sprintf("exclude_reblogs=%t", excludeReblogs),
			fmt.Sprintf("pinned_only=%t", pinnedOnly),
			fmt.Sprintf("only_media=%t", mediaOnly),
			fmt.Sprintf("only_public=%t", publicOnly),
		},
	})
}

func (p *processor) WebStatusesGet(ctx context.Context, targetAccountID string, maxID string) (*apimodel.PageableResponse, gtserror.WithCode) {
	acct, err := p.db.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		if err == db.ErrNoEntries {
			err := fmt.Errorf("account %s not found in the db, not getting web statuses for it", targetAccountID)
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	if acct.Domain != "" {
		err := fmt.Errorf("account %s was not a local account, not getting web statuses for it", targetAccountID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	statuses, err := p.db.GetAccountWebStatuses(ctx, targetAccountID, 10, maxID)
	if err != nil {
		if err == db.ErrNoEntries {
			return util.EmptyPageableResponse(), nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(statuses)

	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	items := []interface{}{}
	nextMaxIDValue := ""
	prevMinIDValue := ""
	for i, s := range statuses {
		item, err := p.tc.StatusToAPIStatus(ctx, s, nil)
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
		Path:             "/@" + acct.Username,
		NextMaxIDValue:   nextMaxIDValue,
		PrevMinIDValue:   prevMinIDValue,
		ExtraQueryParams: []string{},
	})
}
