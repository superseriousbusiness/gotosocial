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

package federatingdb

import (
	"context"
	"errors"
	"net/http"

	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

/*
	The code in this file handles the three types of "polite"
	interaction requests currently recognized by GoToSocial:
	LikeRequest, ReplyRequest, and AnnounceRequest.

	A request looks a bit like this, note the requested
	interaction itself is nested in the "instrument" property:

	{
	  "@context": [
	    "https://www.w3.org/ns/activitystreams",
	    "https://gotosocial.org/ns"
	  ],
	  "type": "LikeRequest",
	  "id": "https://example.com/users/bob/interaction_requests/likes/12345",
	  "actor": "https://example.com/users/bob",
	  "object": "https://example.com/users/alice/statuses/1",
	  "to": "https://example.com/users/alice",
	  "instrument": {
	    "type": "Like",
	    "id": "https://example.com/users/bob/likes/12345",
	    "object": "https://example.com/users/alice/statuses/1",
	    "attributedTo": "https://example.com/users/bob",
	    "to": [
	      "https://www.w3.org/ns/activitystreams#Public",
	      "https://example.com/users/alice"
	    ]
	  }
	}

	Because each of the interaction types are a bit different,
	they're unfortunately also parsed and stored differently:
	LikeRequests have the Like checked first, here, against
	the interaction policy of the target status, whereas
	AnnounceRequests and ReplyRequests have the interaction
	checked against the interaction policy of the target
	status asynchronously, in the FromFediAPI processor.

	It may be possible to dick about with the logic a bit and
	shuffle the checks all here, or all in the processor, but
	that's a job for future refactoring by future tobi/kimbe.
*/

// partialInteractionRequest represents a
// partially-parsed interaction request
// returned from the util function parseInteractionReq.
type partialInteractionRequest struct {
	intRequestURI string
	requesting    *gtsmodel.Account
	receiving     *gtsmodel.Account
	object        *gtsmodel.Status
	instrument    vocab.Type
}

// parseIntReq does some first-pass parsing
// of the given InteractionRequestable (LikeRequest,
// ReplyRequest, AnnounceRequest), checking stuff like:
//
//   - interaction request has a single object
//   - interaction request object is a status
//   - object status belongs to receiving account
//   - interaction request has a single instrument
//
// It returns a partialInteractionRequest struct,
// or an error if something goes wrong.
func (f *DB) parseInteractionRequest(ctx context.Context, intRequest ap.InteractionRequestable) (*partialInteractionRequest, error) {

	// Get and stringify the ID/URI of interaction request once,
	// and mark this particular activity as handled in ID cache.
	intRequestURI := ap.GetJSONLDId(intRequest).String()
	f.activityIDs.Set(intRequestURI, struct{}{})

	// Extract relevant values from passed ctx.
	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		// Already processed.
		return nil, nil
	}

	requesting := activityContext.requestingAcct
	receiving := activityContext.receivingAcct

	if requesting.IsMoving() {
		// A Moving account
		// can't do this.
		return nil, nil
	}

	if receiving.IsMoving() {
		// Moving accounts can't
		// do anything with interaction
		// requests, so ignore it.
		return nil, nil
	}

	// Make sure we have a single
	// object of the interaction request.
	objectIRIs := ap.GetObjectIRIs(intRequest)
	if l := len(objectIRIs); l != 1 {
		return nil, gtserror.NewfWithCode(
			http.StatusBadRequest,
			"invalid object len %d, wanted 1", l,
		)
	}

	// Extract the status URI str.
	statusIRI := objectIRIs[0]
	statusIRIStr := statusIRI.String()

	// Fetch status by given URI from the database.
	status, err := f.state.DB.GetStatusByURI(ctx, statusIRIStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.Newf("db error getting object status %s: %w", statusIRIStr, err)
	}

	// Ensure received by correct account.
	if status.AccountID != receiving.ID {
		return nil, gtserror.NewfWithCode(
			http.StatusForbidden,
			"receiver %s is not owner of interaction-requested status",
			receiving.URI,
		)
	}

	// Ensure we have the expected one instrument.
	instruments := ap.ExtractInstruments(intRequest)
	if l := len(instruments); l != 1 {
		return nil, gtserror.NewfWithCode(
			http.StatusBadRequest,
			"invalid instrument len %d, wanted 1", l,
		)
	}

	// Instrument should be a
	// type and not just an IRI.
	instrument := instruments[0].GetType()
	if instrument == nil {
		return nil, gtserror.NewWithCode(
			http.StatusBadRequest,
			"instrument was not vocab.Type",
		)
	}

	// Check the instrument is an approveable type.
	approvable, ok := instrument.(ap.WithApprovedBy)
	if !ok {
		return nil, gtserror.NewWithCode(
			http.StatusBadRequest,
			"instrument was not Approvable",
		)
	}

	// Ensure that `approvedBy` isn't already set.
	if u := ap.GetApprovedBy(approvable); u != nil {
		return nil, gtserror.NewfWithCode(
			http.StatusBadRequest,
			"instrument claims to already be approvedBy %s",
			u.String(),
		)
	}

	return &partialInteractionRequest{
		intRequestURI: intRequestURI,
		requesting:    requesting,
		receiving:     receiving,
		object:        status,
		instrument:    instrument,
	}, nil
}

func (f *DB) LikeRequest(ctx context.Context, likeReq vocab.GoToSocialLikeRequest) error {
	log.DebugKV(ctx, "LikeRequest", serialize{likeReq})

	// Parse out base level interaction request information.
	partial, err := f.parseInteractionRequest(ctx, likeReq)
	if err != nil {
		return err
	}

	// Ensure the instrument vocab.Type is Likeable.
	likeable, ok := ap.ToLikeable(partial.instrument)
	if !ok {
		return gtserror.NewWithCode(
			http.StatusBadRequest,
			"could not parse instrument to Likeable",
		)
	}

	// Convert received AS like type to internal fave model.
	fave, err := f.converter.ASLikeToFave(ctx, likeable)
	if err != nil {
		err := gtserror.Newf("error converting from AS type: %w", err)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	// Ensure fave enacted by correct account.
	if fave.AccountID != partial.requesting.ID {
		return gtserror.NewfWithCode(
			http.StatusForbidden,
			"requester %s is not expected actor %s",
			partial.requesting.URI, fave.Account.URI,
		)
	}

	// Ensure fave received by correct account.
	if fave.TargetAccountID != partial.receiving.ID {
		return gtserror.NewfWithCode(
			http.StatusForbidden,
			"receiver %s is not expected %s",
			partial.receiving.URI, fave.TargetAccount.URI,
		)
	}

	// Ensure this is a valid Like target for requester.
	policyResult, err := f.intFilter.StatusLikeable(ctx,
		partial.requesting,
		fave.Status,
	)
	if err != nil {
		return gtserror.Newf(
			"error seeing if status %s is likeable: %w",
			fave.Status.URI, err,
		)
	} else if policyResult.Forbidden() {
		return gtserror.NewWithCode(
			http.StatusForbidden,
			"requester does not have permission to Like status",
		)
	}

	// Policy result is either automatic or manual
	// approval, so store the interaction request.
	intReq := &gtsmodel.InteractionRequest{
		ID:                    id.NewULID(),
		TargetStatusID:        fave.StatusID,
		TargetStatus:          fave.Status,
		TargetAccountID:       fave.TargetAccountID,
		TargetAccount:         fave.TargetAccount,
		InteractingAccountID:  fave.AccountID,
		InteractingAccount:    fave.Account,
		InteractionRequestURI: partial.intRequestURI,
		InteractionURI:        fave.URI,
		InteractionType:       gtsmodel.InteractionLike,
		Polite:                util.Ptr(true),
		Like:                  fave,
	}
	switch err := f.state.DB.PutInteractionRequest(ctx, intReq); {
	case err == nil:
		// No problem.

	case errors.Is(err, db.ErrAlreadyExists):
		// Already processed this, race condition? Just warn + return.
		log.Warnf(ctx, "received duplicate interaction request: %s", partial.intRequestURI)
		return nil

	default:
		// Proper DB error.
		return gtserror.Newf(
			"db error storing interaction request %s",
			partial.intRequestURI,
		)
	}

	// Int req is now stored.
	//
	// Set some fields on the
	// pending fave and store it.
	fave.ID = id.NewULID()
	fave.PendingApproval = util.Ptr(true)
	fave.PreApproved = policyResult.AutomaticApproval()
	if err := f.state.DB.PutStatusFave(ctx, fave); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			// The fave already exists in the
			// database, which means we've already
			// handled side effects. We can just
			// return nil here and be done with it.
			return nil
		}

		return gtserror.Newf("error inserting %s into db: %w", fave.URI, err)
	}

	// Further processing will be carried out
	// asynchronously, and our caller will return 202 Accepted.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APActivityType: ap.ActivityCreate,
		APObjectType:   ap.ActivityLikeRequest,
		GTSModel:       intReq,
		Receiving:      partial.receiving,
		Requesting:     partial.requesting,
	})

	return nil
}

func (f *DB) ReplyRequest(ctx context.Context, replyReq vocab.GoToSocialReplyRequest) error {
	log.DebugKV(ctx, "ReplyRequest", serialize{replyReq})

	// Parse out base level interaction request information.
	partial, err := f.parseInteractionRequest(ctx, replyReq)
	if err != nil {
		return err
	}

	// Ensure the instrument vocab.Type is Statusable.
	statusable, ok := ap.ToStatusable(partial.instrument)
	if !ok {
		return gtserror.NewWithCode(
			http.StatusBadRequest,
			"could not parse instrument to Statusable",
		)
	}

	// Check for spam / relevance.
	ok, err = f.statusableOK(
		ctx,
		partial.receiving,
		partial.requesting,
		statusable,
	)
	if err != nil {
		// Error already
		// wrapped.
		return err
	}

	if !ok {
		// Not relevant / spam.
		// Already logged.
		return nil
	}

	// Statusable must reply to something.
	inReplyToURIs := ap.GetInReplyTo(statusable)
	if l := len(inReplyToURIs); l != 1 {
		return gtserror.NewfWithCode(
			http.StatusBadRequest,
			"expected inReplyTo length 1, got %d", l,
		)
	}
	inReplyToURI := inReplyToURIs[0]
	inReplyToURIStr := inReplyToURI.String()

	// Make sure we have the status this interaction reply encompasses.
	inReplyTo, err := f.state.DB.GetStatusByURI(ctx, inReplyToURIStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("db error getting inReplyTo status %s: %w", inReplyToURIStr, err)
	}

	// Check status exists.
	if inReplyTo == nil {
		log.Warnf(ctx, "received ReplyRequest for non-existent status: %s", inReplyToURIStr)
		return nil
	}

	// Make sure the parent status is owned by receiver.
	if inReplyTo.AccountURI != partial.receiving.URI {
		return gtserror.NewfWithCode(
			http.StatusBadRequest,
			"inReplyTo status %s not owned by receiving account %s",
			inReplyToURIStr, partial.receiving.URI,
		)
	}

	// Extract the attributed to (i.e. author) URI of status.
	attributedToURI, err := ap.ExtractAttributedToURI(statusable)
	if err != nil {
		err := gtserror.Newf("invalid status attributedTo value: %w", err)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	// Ensure status author is account of requester.
	attributedToURIStr := attributedToURI.String()
	if attributedToURIStr != partial.requesting.URI {
		return gtserror.NewfWithCode(
			http.StatusBadRequest,
			"status attributedTo %s not requesting account %s",
			inReplyToURIStr, partial.requesting.URI,
		)
	}

	// Create a pending interaction request in the database.
	// This request will be handled further by the processor.
	intReq := &gtsmodel.InteractionRequest{
		ID:                    id.NewULID(),
		TargetStatusID:        inReplyTo.ID,
		TargetStatus:          inReplyTo,
		TargetAccountID:       inReplyTo.AccountID,
		TargetAccount:         inReplyTo.Account,
		InteractingAccountID:  partial.requesting.ID,
		InteractingAccount:    partial.requesting,
		InteractionRequestURI: partial.intRequestURI,
		InteractionURI:        ap.GetJSONLDId(statusable).String(),
		InteractionType:       gtsmodel.InteractionReply,
		Polite:                util.Ptr(true),
		Reply:                 nil, // Not settable yet.
	}
	switch err := f.state.DB.PutInteractionRequest(ctx, intReq); {
	case err == nil:
		// No problem.

	case errors.Is(err, db.ErrAlreadyExists):
		// Already processed this, race condition? Just warn + return.
		log.Warnf(ctx, "received duplicate interaction request: %s", partial.intRequestURI)
		return nil

	default:
		// Proper DB error.
		return gtserror.Newf(
			"db error storing interaction request %s",
			partial.intRequestURI,
		)
	}

	// Further processing will be carried out
	// asynchronously, return 202 Accepted.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APActivityType: ap.ActivityCreate,
		APObjectType:   ap.ActivityReplyRequest,
		GTSModel:       intReq,
		APObject:       statusable,
		Receiving:      partial.receiving,
		Requesting:     partial.requesting,
	})

	return nil
}

func (f *DB) AnnounceRequest(ctx context.Context, announceReq vocab.GoToSocialAnnounceRequest) error {
	log.DebugKV(ctx, "AnnounceRequest", serialize{announceReq})

	// Parse out base level interaction request information.
	partial, err := f.parseInteractionRequest(ctx, announceReq)
	if err != nil {
		return err
	}

	// Ensure the instrument vocab.Type is Announceable.
	announceable, ok := ap.ToAnnounceable(partial.instrument)
	if !ok {
		return gtserror.NewWithCode(
			http.StatusBadRequest,
			"could not parse instrument to Announceable",
		)
	}

	// Convert received AS Announce type to internal boost wrapper model.
	boost, new, err := f.converter.ASAnnounceToStatus(ctx, announceable)
	if err != nil {
		err := gtserror.Newf("error converting from AS type: %w", err)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	if !new {
		// We already have this announce, just return.
		log.Warnf(ctx, "received AnnounceRequest for existing announce: %s", boost.URI)
		return nil
	}

	// Fetch origin status that this boost is targetting from database.
	targetStatus, err := f.state.DB.GetStatusByURI(ctx, boost.BoostOfURI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf(
			"db error getting announce object %s: %w",
			boost.BoostOfURI, err,
		)
	}

	if targetStatus == nil {
		// Status doesn't seem to exist, just drop this AnnounceRequest.
		log.Warnf(ctx, "received AnnounceRequest for non-existent status %s", boost.BoostOfURI)
		return nil
	}

	// Ensure target status is owned by receiving account.
	if targetStatus.AccountID != partial.receiving.ID {
		return gtserror.NewfWithCode(
			http.StatusBadRequest,
			"announce object %s not owned by receiving account %s",
			boost.BoostOfURI, partial.receiving.URI,
		)
	}

	// Ensure announce enacted by correct account.
	if boost.AccountID != partial.requesting.ID {
		return gtserror.NewfWithCode(
			http.StatusForbidden,
			"requester %s is not expected actor %s",
			partial.requesting.URI, boost.Account.URI,
		)
	}

	// Create a pending interaction request in the database.
	// This request will be handled further by the processor.
	intReq := &gtsmodel.InteractionRequest{
		ID:                    id.NewULID(),
		TargetStatusID:        targetStatus.ID,
		TargetStatus:          targetStatus,
		TargetAccountID:       targetStatus.AccountID,
		TargetAccount:         targetStatus.Account,
		InteractingAccountID:  boost.AccountID,
		InteractingAccount:    boost.Account,
		InteractionRequestURI: partial.intRequestURI,
		InteractionURI:        boost.URI,
		InteractionType:       gtsmodel.InteractionAnnounce,
		Polite:                util.Ptr(true),
		Announce:              boost,
	}
	switch err := f.state.DB.PutInteractionRequest(ctx, intReq); {
	case err == nil:
		// No problem.

	case errors.Is(err, db.ErrAlreadyExists):
		// Already processed this, race condition? Just warn + return.
		log.Warnf(ctx, "received duplicate interaction request: %s", partial.intRequestURI)
		return nil

	default:
		// Proper DB error.
		return gtserror.Newf(
			"db error storing interaction request %s",
			partial.intRequestURI,
		)
	}

	// Further processing will be carried out
	// asynchronously, return 202 Accepted.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APActivityType: ap.ActivityCreate,
		APObjectType:   ap.ActivityAnnounceRequest,
		GTSModel:       intReq,
		Receiving:      partial.receiving,
		Requesting:     partial.requesting,
	})

	return nil
}
