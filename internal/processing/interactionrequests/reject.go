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

// Reject rejects an interaction request with the given ID,
// on behalf of the requester (whose post it must target).
func (p *Processor) Reject(
	ctx context.Context,
	requester *gtsmodel.Account,
	intReqID string,
) (*apimodel.InteractionRejection, gtserror.WithCode) {
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

	// Derive rejection + error depending on
	// the type of interaction being rejected.
	var (
		rejection   *gtsmodel.InteractionRejection
		errWithCode gtserror.WithCode
	)

	switch req.InteractionType {

	case gtsmodel.InteractionLike:
		rejection, errWithCode = p.rejectLike(ctx, req)

	case gtsmodel.InteractionReply:
		rejection, errWithCode = p.rejectReply(ctx, req)

	case gtsmodel.InteractionAnnounce:
		rejection, errWithCode = p.rejectAnnounce(ctx, req)

	default:
		err := gtserror.Newf("unknown interaction type for interaction request %s", intReqID)
		errWithCode = gtserror.NewErrorInternalError(err)
	}

	if errWithCode != nil {
		return nil, errWithCode
	}

	// Return the rejection to the caller so they
	// can do something with it if they need to.
	apiRejection, err := p.converter.InteractionRejectionToAPIInteractionRejection(
		ctx,
		rejection,
		requester,
	)
	if err != nil {
		err := gtserror.Newf("error converting interaction rejection: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiRejection, nil
}

// Package-internal convenience function to reject a like.
func (p *Processor) rejectLike(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
) (*gtsmodel.InteractionRejection, gtserror.WithCode) {
	if req.Like == nil {
		// Like undone? Race condition?
		// Nothing we can do anyway.
		err := gtserror.New("req.Like was nil, nothing to reject")
		errWithCode := gtserror.NewErrorNotFound(err)
		return nil, errWithCode
	}

	// Mark fave as rejected and store
	// a new interactionRejection for it.
	rejection, err := p.c.RejectFave(ctx, req.Like)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Send the rejection off through the client
	// API processor to handle side effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActivityLike,
		APActivityType: ap.ActivityReject,
		GTSModel:       rejection,
		Origin:         rejection.Account,
		Target:         rejection.InteractingAccount,
	})

	return rejection, nil
}

// Package-internal convenience function to reject a reply.
func (p *Processor) rejectReply(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
) (*gtsmodel.InteractionRejection, gtserror.WithCode) {
	if req.Reply == nil {
		// Reply deleted? Race condition?
		// Nothing we can do anyway.
		err := gtserror.New("req.Reply was nil, nothing to reject")
		errWithCode := gtserror.NewErrorNotFound(err)
		return nil, errWithCode
	}

	// Mark reply as rejected and store
	// a new interactionRejection for it.
	rejection, err := p.c.RejectReply(ctx, req.Reply)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Send the rejection off through the client
	// API processor to handle side effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityReject,
		GTSModel:       rejection,
		Origin:         rejection.Account,
		Target:         rejection.InteractingAccount,
	})

	return rejection, nil
}

// Package-internal convenience function to reject an announce.
func (p *Processor) rejectAnnounce(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
) (*gtsmodel.InteractionRejection, gtserror.WithCode) {
	if req.Announce == nil {
		// Announce undone? Race condition?
		// Nothing we can do anyway.
		err := gtserror.New("req.Announce was nil, nothing to reject")
		errWithCode := gtserror.NewErrorNotFound(err)
		return nil, errWithCode
	}

	// Mark announce as rejected and store
	// a new interactionRejection for it.
	rejection, err := p.c.RejectAnnounce(ctx, req.Announce)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Send the rejection off through the client
	// API processor to handle side effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActivityAnnounce,
		APActivityType: ap.ActivityReject,
		GTSModel:       rejection,
		Origin:         rejection.Account,
		Target:         rejection.InteractingAccount,
	})

	return rejection, nil
}
