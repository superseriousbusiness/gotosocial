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

package interactionrequests

import (
	"context"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// Accept accepts an interaction request with the given ID,
// on behalf of the given account (whose post it must target).
func (p *Processor) Accept(
	ctx context.Context,
	acct *gtsmodel.Account,
	reqID string,
) (*apimodel.InteractionRequest, gtserror.WithCode) {
	req, err := p.state.DB.GetInteractionRequestByID(ctx, reqID)
	if err != nil {
		err := gtserror.Newf("db error getting interaction request: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if req.TargetAccountID != acct.ID {
		err := gtserror.Newf(
			"interaction request %s does not belong to account %s",
			reqID, acct.ID,
		)
		return nil, gtserror.NewErrorNotFound(err)
	}

	if !req.IsPending() {
		err := gtserror.Newf(
			"interaction request %s has already been handled",
			reqID,
		)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Lock on the interaction req URI to
	// ensure nobody else is modifying it rn.
	unlock := p.state.ProcessingLocks.Lock(req.InteractionURI)
	defer unlock()

	// Mark the request as accepted
	// and generate a URI for it.
	req.AcceptedAt = time.Now()
	req.URI = uris.GenerateURIForAccept(acct.Username, req.ID)
	if err := p.state.DB.UpdateInteractionRequest(
		ctx,
		req,
		"accepted_at",
		"uri",
	); err != nil {
		err := gtserror.Newf("db error updating interaction request: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	switch req.InteractionType {

	case gtsmodel.InteractionLike:
		if errWithCode := p.acceptLike(ctx, req); errWithCode != nil {
			return nil, errWithCode
		}

	case gtsmodel.InteractionReply:
		if errWithCode := p.acceptReply(ctx, req); errWithCode != nil {
			return nil, errWithCode
		}

	case gtsmodel.InteractionAnnounce:
		if errWithCode := p.acceptAnnounce(ctx, req); errWithCode != nil {
			return nil, errWithCode
		}

	default:
		err := gtserror.Newf("unknown interaction type for interaction request %s", reqID)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return the now-accepted req to the caller so
	// they can do something with it if they need to.
	apiReq, err := p.converter.InteractionReqToAPIInteractionReq(
		ctx,
		req,
		acct,
	)
	if err != nil {
		err := gtserror.Newf("error converting interaction request: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiReq, nil
}

// Package-internal convenience
// function to accept a like.
func (p *Processor) acceptLike(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
) gtserror.WithCode {
	// If the Like is missing, that means it's
	// probably already been undone by someone,
	// so there's nothing to actually accept.
	if req.Like == nil {
		err := gtserror.Newf("no Like found for interaction request %s", req.ID)
		return gtserror.NewErrorNotFound(err)
	}

	// Update the Like.
	req.Like.PendingApproval = util.Ptr(false)
	req.Like.PreApproved = false
	req.Like.ApprovedByURI = req.URI
	if err := p.state.DB.UpdateStatusFave(
		ctx,
		req.Like,
		"pending_approval",
		"approved_by_uri",
	); err != nil {
		err := gtserror.Newf("db error updating status fave: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	// Send the accepted request off through the
	// client API processor to handle side effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActivityLike,
		APActivityType: ap.ActivityAccept,
		GTSModel:       req,
		Origin:         req.TargetAccount,
		Target:         req.InteractingAccount,
	})

	return nil
}

// Package-internal convenience
// function to accept a reply.
func (p *Processor) acceptReply(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
) gtserror.WithCode {
	// If the Reply is missing, that means it's
	// probably already been undone by someone,
	// so there's nothing to actually accept.
	if req.Reply == nil {
		err := gtserror.Newf("no Reply found for interaction request %s", req.ID)
		return gtserror.NewErrorNotFound(err)
	}

	// Update the Reply.
	req.Reply.PendingApproval = util.Ptr(false)
	req.Reply.PreApproved = false
	req.Reply.ApprovedByURI = req.URI
	if err := p.state.DB.UpdateStatus(
		ctx,
		req.Reply,
		"pending_approval",
		"approved_by_uri",
	); err != nil {
		err := gtserror.Newf("db error updating status reply: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	// Send the accepted request off through the
	// client API processor to handle side effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityAccept,
		GTSModel:       req,
		Origin:         req.TargetAccount,
		Target:         req.InteractingAccount,
	})

	return nil
}

// Package-internal convenience
// function to accept an announce.
func (p *Processor) acceptAnnounce(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
) gtserror.WithCode {
	// If the Announce is missing, that means it's
	// probably already been undone by someone,
	// so there's nothing to actually accept.
	if req.Reply == nil {
		err := gtserror.Newf("no Announce found for interaction request %s", req.ID)
		return gtserror.NewErrorNotFound(err)
	}

	// Update the Announce.
	req.Announce.PendingApproval = util.Ptr(false)
	req.Announce.PreApproved = false
	req.Announce.ApprovedByURI = req.URI
	if err := p.state.DB.UpdateStatus(
		ctx,
		req.Announce,
		"pending_approval",
		"approved_by_uri",
	); err != nil {
		err := gtserror.Newf("db error updating status announce: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	// Send the accepted request off through the
	// client API processor to handle side effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActivityAnnounce,
		APActivityType: ap.ActivityAccept,
		GTSModel:       req,
		Origin:         req.TargetAccount,
		Target:         req.InteractingAccount,
	})

	return nil
}
