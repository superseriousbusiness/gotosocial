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

package account

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// StatusesGet fetches a number of statuses (in time descending order) from the given account, filtered by visibility for
// the account given in authed.
func (p *Processor) StatusesGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string, limit int, excludeReplies bool, excludeReblogs bool, maxID string, minID string, pinned bool, mediaOnly bool, publicOnly bool) (*apimodel.PageableResponse, gtserror.WithCode) {
	if requestingAccount != nil {
		if blocked, err := p.state.DB.IsEitherBlocked(ctx, requestingAccount.ID, targetAccountID); err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		} else if blocked {
			err := errors.New("block exists between accounts")
			return nil, gtserror.NewErrorNotFound(err)
		}
	}

	var (
		statuses []*gtsmodel.Status
		err      error
	)
	if pinned {
		// Get *ONLY* pinned statuses.
		statuses, err = p.state.DB.GetAccountPinnedStatuses(ctx, targetAccountID)
	} else {
		// Get account statuses which *may* include pinned ones.
		statuses, err = p.state.DB.GetAccountStatuses(ctx, targetAccountID, limit, excludeReplies, excludeReblogs, maxID, minID, mediaOnly, publicOnly)
	}
	if err != nil {
		if err == db.ErrNoEntries {
			return util.EmptyPageableResponse(), nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Filtering + serialization process is the same for either pinned status queries or 'normal' ones.
	filtered, err := p.filter.StatusesVisible(ctx, requestingAccount, statuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(filtered)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	items := make([]interface{}, 0, count)
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

	if pinned {
		// We don't page on pinned status responses,
		// so we can save some work + just return items.
		return &apimodel.PageableResponse{
			Items: items,
		}, nil
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
			fmt.Sprintf("pinned=%t", pinned),
			fmt.Sprintf("only_media=%t", mediaOnly),
			fmt.Sprintf("only_public=%t", publicOnly),
		},
	})
}

// WebStatusesGet fetches a number of statuses (in descending order) from the given account. It selects only
// statuses which are suitable for showing on the public web profile of an account.
func (p *Processor) WebStatusesGet(ctx context.Context, targetAccountID string, maxID string) (*apimodel.PageableResponse, gtserror.WithCode) {
	acct, err := p.state.DB.GetAccountByID(ctx, targetAccountID)
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

	statuses, err := p.state.DB.GetAccountWebStatuses(ctx, targetAccountID, 10, maxID)
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
