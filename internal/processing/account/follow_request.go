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

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// FollowRequestAccept handles the accepting of a follow request from the sourceAccountID to the requestingAccount (the currently authorized account).
func (p *Processor) FollowRequestAccept(ctx context.Context, requestingAccount *gtsmodel.Account, sourceAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	follow, err := p.state.DB.AcceptFollowRequest(ctx, sourceAccountID, requestingAccount.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	if follow.Account != nil {
		// Only enqueue work in the case we have a request creating account stored.
		// NOTE: due to how AcceptFollowRequest works, the inverse shouldn't be possible.
		p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityAccept,
			GTSModel:       follow,
			OriginAccount:  follow.Account,
			TargetAccount:  follow.TargetAccount,
		})
	}

	return p.RelationshipGet(ctx, requestingAccount, sourceAccountID)
}

// FollowRequestReject handles the rejection of a follow request from the sourceAccountID to the requestingAccount (the currently authorized account).
func (p *Processor) FollowRequestReject(ctx context.Context, requestingAccount *gtsmodel.Account, sourceAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	followRequest, err := p.state.DB.GetFollowRequest(ctx, sourceAccountID, requestingAccount.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	err = p.state.DB.RejectFollowRequest(ctx, sourceAccountID, requestingAccount.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	if followRequest.Account != nil {
		// Only enqueue work in the case we have a request creating account stored.
		// NOTE: due to how GetFollowRequest works, the inverse shouldn't be possible.
		p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityReject,
			GTSModel:       followRequest,
			OriginAccount:  followRequest.Account,
			TargetAccount:  followRequest.TargetAccount,
		})
	}

	return p.RelationshipGet(ctx, requestingAccount, sourceAccountID)
}

// FollowRequestsGet fetches a list of the accounts that are follow requesting the given requestingAccount (the currently authorized account).
func (p *Processor) FollowRequestsGet(ctx context.Context, requestingAccount *gtsmodel.Account, page *paging.Page) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Fetch follow requests targeting the given requesting account model.
	followRequests, err := p.state.DB.GetAccountFollowRequests(ctx, requestingAccount.ID, page)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check for empty response.
	count := len(followRequests)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	var (
		items = make([]interface{}, 0, count)

		// Set next + prev values before filtering and API
		// converting, so caller can still page properly.
		nextMaxIDValue = followRequests[count-1].ID
		prevMinIDValue = followRequests[0].ID
	)

	// Convert database account models to API account models.
	for _, followRequest := range followRequests {
		apiAcct, err := p.tc.AccountToAPIAccountPublic(ctx, followRequest.Account)
		if err != nil {
			log.Errorf(ctx, "error convering to public api account: %v", err)
			continue
		}
		items = append(items, apiAcct)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/follow_requests",
		Next:  page.Next(nextMaxIDValue),
		Prev:  page.Next(prevMinIDValue),
	}), nil
}
