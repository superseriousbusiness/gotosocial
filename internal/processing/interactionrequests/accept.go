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
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// Accept accepts an interaction request with the given ID,
// on behalf of the requester (whose post it must target).
func (p *Processor) Accept(
	ctx context.Context,
	requester *gtsmodel.Account,
	intReqID string,
) (*apimodel.InteractionApproval, gtserror.WithCode) {
	req, err := p.state.DB.GetInteractionRequestByID(ctx, intReqID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting interaction request: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if req == nil {
		err := gtserror.Newf("interaction request %s not found", intReqID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	if req.TargetAccountID != requester.ID {
		err := gtserror.Newf(
			"interaction request %s does not belong to account %s",
			intReqID, requester.ID,
		)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Lock on the interaction req URI to
	// ensure nobody else is modifying it rn.
	unlock := p.state.ProcessingLocks.Lock(req.InteractionURI)
	defer unlock()

	// Delete the request from the db; we have it in
	// memory now + we don't need it in the db anymore.
	if err := p.state.DB.DeleteInteractionRequestByID(ctx, req.ID); err != nil {
		err := gtserror.Newf("db error deleting interaction request: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Derive approval + error depending on
	// the type of interaction being approved.
	var (
		approval    *gtsmodel.InteractionApproval
		errWithCode gtserror.WithCode
	)

	switch req.InteractionType {

	case gtsmodel.InteractionLike:
		approval, errWithCode = p.acceptLike(ctx, req)

	case gtsmodel.InteractionReply:
		approval, errWithCode = p.acceptReply(ctx, req)

	case gtsmodel.InteractionAnnounce:
		approval, errWithCode = p.acceptAnnounce(ctx, req)

	default:
		err := gtserror.Newf("unknown interaction type for interaction request %s", intReqID)
		errWithCode = gtserror.NewErrorInternalError(err)
	}

	if errWithCode != nil {
		return nil, errWithCode
	}

	// Return the approval to the caller so they
	// can do something with it if they need to.
	apiApproval, err := p.converter.InteractionApprovalToAPIInteractionApproval(
		ctx,
		approval,
		requester,
	)
	if err != nil {
		err := gtserror.Newf("error converting interaction approval: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiApproval, nil
}

// Package-internal convenience function to accept a like.
func (p *Processor) acceptLike(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
) (*gtsmodel.InteractionApproval, gtserror.WithCode) {
	if req.Like == nil {
		// Like undone? Race condition?
		// Nothing we can do anyway.
		err := gtserror.New("req.Like was nil, nothing to accept")
		errWithCode := gtserror.NewErrorNotFound(err)
		return nil, errWithCode
	}

	// Mark fave as approved and store
	// a new interactionApproval for it.
	approval, err := p.c.ApproveFave(ctx, req.Like)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Send the approval off through the client
	// API processor to handle side effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActivityLike,
		APActivityType: ap.ActivityAccept,
		GTSModel:       approval,
		Origin:         approval.Account,
		Target:         approval.InteractingAccount,
	})

	return approval, nil
}

// Package-internal convenience function to accept a reply.
func (p *Processor) acceptReply(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
) (*gtsmodel.InteractionApproval, gtserror.WithCode) {
	if req.Reply == nil {
		// Reply deleted? Race condition?
		// Nothing we can do anyway.
		err := gtserror.New("req.Reply was nil, nothing to accept")
		errWithCode := gtserror.NewErrorNotFound(err)
		return nil, errWithCode
	}

	// Mark reply as approved and store
	// a new interactionApproval for it.
	approval, err := p.c.ApproveReply(ctx, req.Reply)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Send the approval off through the client
	// API processor to handle side effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityAccept,
		GTSModel:       approval,
		Origin:         approval.Account,
		Target:         approval.InteractingAccount,
	})

	return approval, nil
}

// Package-internal convenience function to accept an announce.
func (p *Processor) acceptAnnounce(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
) (*gtsmodel.InteractionApproval, gtserror.WithCode) {
	if req.Announce == nil {
		// Announce undone? Race condition?
		// Nothing we can do anyway.
		err := gtserror.New("req.Announce was nil, nothing to accept")
		errWithCode := gtserror.NewErrorNotFound(err)
		return nil, errWithCode
	}

	// Mark announce as approved and store
	// a new interactionApproval for it.
	approval, err := p.c.ApproveAnnounce(ctx, req.Announce)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Send the approval off through the client
	// API processor to handle side effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActivityAnnounce,
		APActivityType: ap.ActivityAccept,
		GTSModel:       approval,
		Origin:         approval.Account,
		Target:         approval.InteractingAccount,
	})

	return approval, nil
}
