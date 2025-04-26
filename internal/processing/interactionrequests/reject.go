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
)

// Reject rejects an interaction request with the given ID,
// on behalf of the given account (whose post it must target).
func (p *Processor) Reject(
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

	// Mark the request as rejected
	// and generate a URI for it.
	req.RejectedAt = time.Now()
	req.URI = uris.GenerateURIForReject(acct.Username, req.ID)
	if err := p.state.DB.UpdateInteractionRequest(
		ctx,
		req,
		"rejected_at",
		"uri",
	); err != nil {
		err := gtserror.Newf("db error updating interaction request: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	switch req.InteractionType {

	case gtsmodel.InteractionLike:
		// Send the rejected request off through the
		// client API processor to handle side effects.
		p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
			APObjectType:   ap.ActivityLike,
			APActivityType: ap.ActivityReject,
			GTSModel:       req,
			Origin:         req.TargetAccount,
			Target:         req.InteractingAccount,
		})

	case gtsmodel.InteractionReply:
		// Send the rejected request off through the
		// client API processor to handle side effects.
		p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityReject,
			GTSModel:       req,
			Origin:         req.TargetAccount,
			Target:         req.InteractingAccount,
		})

	case gtsmodel.InteractionAnnounce:
		// Send the rejected request off through the
		// client API processor to handle side effects.
		p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
			APObjectType:   ap.ActivityAnnounce,
			APActivityType: ap.ActivityReject,
			GTSModel:       req,
			Origin:         req.TargetAccount,
			Target:         req.InteractingAccount,
		})

	default:
		err := gtserror.Newf("unknown interaction type for interaction request %s", reqID)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return the now-rejected req to the caller so
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
