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
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// StatusesGet fetches a number of statuses (in time descending order) from the
// target account, filtered by visibility according to the requesting account.
func (p *Processor) StatusesGet(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	targetAccountID string,
	limit int,
	excludeReplies bool,
	excludeReblogs bool,
	maxID string,
	minID string,
	pinned bool,
	mediaOnly bool,
	publicOnly bool,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	if requestingAccount != nil {
		blocked, err := p.state.DB.IsEitherBlocked(ctx, requestingAccount.ID, targetAccountID)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}

		if blocked {
			// Block exists between accounts.
			// Just return empty statuses.
			return util.EmptyPageableResponse(), nil
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

	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(statuses)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	var (
		items = make([]interface{}, 0, count)

		// Set next + prev values before filtering and API
		// converting, so caller can still page properly.
		nextMaxIDValue = statuses[count-1].ID
		prevMinIDValue = statuses[0].ID
	)

	// Filtering + serialization process is the same for
	// both pinned status queries and 'normal' ones.
	filtered, err := p.visFilter.StatusesVisible(ctx, requestingAccount, statuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	filters, err := p.state.DB.GetFiltersForAccountID(ctx, requestingAccount.ID)
	if err != nil {
		err = gtserror.Newf("couldn't retrieve filters for account %s: %w", requestingAccount.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	for _, s := range filtered {
		// Convert filtered statuses to API statuses.
		item, err := p.converter.StatusToAPIStatus(ctx, s, requestingAccount, statusfilter.FilterContextAccount, filters, nil)
		if err != nil {
			log.Errorf(ctx, "error convering to api status: %v", err)
			continue
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
		Path:           "/api/v1/accounts/" + targetAccountID + "/statuses",
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

// WebStatusesGet fetches a number of statuses (in descending order)
// from the given account. It selects only statuses which are suitable
// for showing on the public web profile of an account.
func (p *Processor) WebStatusesGet(
	ctx context.Context,
	targetAccountID string,
	maxID string,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	account, err := p.state.DB.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			err := fmt.Errorf("account %s not found in the db, not getting web statuses for it", targetAccountID)
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	if account.Domain != "" {
		err := fmt.Errorf("account %s was not a local account, not getting web statuses for it", targetAccountID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	statuses, err := p.state.DB.GetAccountWebStatuses(ctx, account, 10, maxID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(statuses)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	var (
		items = make([]interface{}, 0, count)

		// Set next value before API converting,
		// so caller can still page properly.
		nextMaxIDValue = statuses[count-1].ID
	)

	for _, s := range statuses {
		// Convert fetched statuses to web view statuses.
		item, err := p.converter.StatusToWebStatus(ctx, s)
		if err != nil {
			log.Errorf(ctx, "error convering to web status: %v", err)
			continue
		}
		items = append(items, item)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "/@" + account.Username,
		NextMaxIDValue: nextMaxIDValue,
	})
}

// WebStatusesGetPinned returns web versions of pinned statuses.
func (p *Processor) WebStatusesGetPinned(
	ctx context.Context,
	targetAccountID string,
) ([]*apimodel.WebStatus, gtserror.WithCode) {
	statuses, err := p.state.DB.GetAccountPinnedStatuses(ctx, targetAccountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	webStatuses := make([]*apimodel.WebStatus, 0, len(statuses))
	for _, status := range statuses {
		// Ensure visible via the web.
		visible, err := p.visFilter.StatusVisible(ctx, nil, status)
		if err != nil {
			log.Errorf(ctx, "error checking status visibility: %v", err)
			continue
		}

		if !visible {
			// Don't serve.
			continue
		}

		webStatus, err := p.converter.StatusToWebStatus(ctx, status)
		if err != nil {
			log.Errorf(ctx, "error convering to web status: %v", err)
			continue
		}

		// Normally when viewed via the API, 'pinned' is
		// only true if the *viewing account* has pinned
		// the status being viewed. For web statuses,
		// however, we still want to be able to indicate
		// a pinned status, so bodge this in here.
		webStatus.Pinned = true

		webStatuses = append(webStatuses, webStatus)
	}

	return webStatuses, nil
}
