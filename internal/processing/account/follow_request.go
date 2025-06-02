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

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
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
		p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityAccept,
			GTSModel:       follow,
			Origin:         follow.Account,
			Target:         follow.TargetAccount,
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
		p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityReject,
			GTSModel:       followRequest,
			Origin:         followRequest.Account,
			Target:         followRequest.TargetAccount,
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

	// Get the lowest and highest
	// ID values, used for paging.
	lo := followRequests[count-1].ID
	hi := followRequests[0].ID

	// Func to fetch follow source at index.
	getIdx := func(i int) *gtsmodel.Account {
		return followRequests[i].Account
	}

	// Get a filtered slice of public API account models.
	items := p.c.GetVisibleAPIAccountsPaged(ctx,
		requestingAccount,
		getIdx,
		count,
	)

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/follow_requests",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}

// OutgoingFollowRequestsGet fetches a list of the accounts with a pending follow request originating from the given requestingAccount (the currently authorized account).
func (p *Processor) OutgoingFollowRequestsGet(ctx context.Context, requestingAccount *gtsmodel.Account, page *paging.Page) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Fetch follow requests originating from the given requesting account model.
	followRequests, err := p.state.DB.GetAccountFollowRequesting(ctx, requestingAccount.ID, page)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check for empty response.
	count := len(followRequests)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := followRequests[count-1].ID
	hi := followRequests[0].ID

	// Func to fetch follow source at index.
	getIdx := func(i int) *gtsmodel.Account {
		return followRequests[i].TargetAccount
	}

	// Get a filtered slice of public API account models.
	items := p.c.GetVisibleAPIAccountsPaged(ctx,
		requestingAccount,
		getIdx,
		count,
	)

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/follow_requests/outgoing",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}
