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

// firstPassIntReq represents a partially-parsed
// interaction request returned from the util
// function parseInteractionReq.
type firstPassIntReq struct {
	intReqURI  string
	requesting *gtsmodel.Account
	receiving  *gtsmodel.Account
	object     *gtsmodel.Status
	instrument vocab.Type
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
// It returns a firstPassIntReq struct, or an error
// if something goes wrong.
func (f *DB) parseIntReq(ctx context.Context, intReq ap.InteractionRequestable) (*firstPassIntReq, error) {

	// Get and stringify the
	// ID/URI of the int req once.
	intReqURI := ap.GetJSONLDId(intReq).String()

	// Mark activity as handled.
	f.activityIDs.Set(intReqURI, struct{}{})

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
	objectIRIs := ap.GetObjectIRIs(intReq)
	if l := len(objectIRIs); l != 1 {
		return nil, gtserror.NewfWithCode(
			http.StatusBadRequest,
			"invalid object len %d, wanted 1", l,
		)
	}
	statusIRI := objectIRIs[0]
	statusIRIStr := statusIRI.String()

	// Object should be a status.
	status, err := f.state.DB.GetStatusByURI(ctx, statusIRIStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting object status %s: %w", statusIRIStr, err)
		return nil, err
	}

	// Ensure int req received by correct account.
	if status.AccountID != receiving.ID {
		err := gtserror.NewfWithCode(
			http.StatusForbidden,
			"receiver %s is not owner of interaction-requested status",
			receiving.URI,
		)
		return nil, err
	}

	// We should have one instrument.
	instruments := ap.ExtractInstruments(intReq)
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
			"instrument was not a type",
		)
	}

	// Check the instrument is
	// something that can be approved.
	approvable, ok := instrument.(ap.WithApprovedBy)
	if !ok {
		return nil, gtserror.NewWithCode(
			http.StatusBadRequest,
			"instrument was not an Approvable",
		)
	}

	// Make sure `approvedBy` isn't
	// already set on the instrument.
	if u := ap.GetApprovedBy(approvable); u != nil {
		return nil, gtserror.NewfWithCode(
			http.StatusBadRequest,
			"instrument claims to already be approvedBy %s",
			u.String(),
		)
	}

	return &firstPassIntReq{
		intReqURI:  ap.GetJSONLDId(intReq).String(),
		requesting: requesting,
		receiving:  receiving,
		object:     status,
		instrument: instrument,
	}, nil
}

func (f *DB) LikeRequest(ctx context.Context, likeReq vocab.GoToSocialLikeRequest) error {
	log.DebugKV(ctx, "LikeRequest", serialize{likeReq})

	// Parse out basic interaction request stuff.
	fpir, err := f.parseIntReq(ctx, likeReq)
	if err != nil {
		return err
	}

	// Parse instrument vocab.Type to Like.
	if fpir.instrument.GetTypeName() != ap.ActivityLike {
		return gtserror.NewWithCode(
			http.StatusBadRequest,
			"instrument of LikeRequest was not a Like",
		)
	}

	likeable, ok := fpir.instrument.(vocab.ActivityStreamsLike)
	if !ok {
		return gtserror.NewWithCode(
			http.StatusBadRequest,
			"could not parse instrument of LikeRequest to Like",
		)
	}

	// Convert received AS like type to internal fave model.
	fave, err := f.converter.ASLikeToFave(ctx, likeable)
	if err != nil {
		err := gtserror.Newf("error converting from AS type: %w", err)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	// Ensure fave enacted by correct account.
	if fave.AccountID != fpir.requesting.ID {
		return gtserror.NewfWithCode(
			http.StatusForbidden,
			"requester %s is not expected actor %s",
			fpir.requesting.URI, fave.Account.URI,
		)
	}

	// Ensure fave received by correct account.
	if fave.TargetAccountID != fpir.receiving.ID {
		err := gtserror.NewfWithCode(
			http.StatusForbidden,
			"receiver %s is not expected %s",
			fpir.receiving.URI, fave.TargetAccount.URI,
		)
		return err
	}

	// Ensure not invalid Like target for requester.
	policyCheckResult, err := f.intFilter.StatusLikeable(ctx,
		fpir.requesting,
		fave.Status,
	)
	if err != nil {
		return gtserror.Newf(
			"error seeing if status %s is likeable: %w",
			fave.Status.URI, err,
		)
	}

	if policyCheckResult.Forbidden() {
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
		InteractionRequestURI: fpir.intReqURI,
		InteractionURI:        fave.URI,
		InteractionType:       gtsmodel.InteractionLike,
		Like:                  fave,
	}

	switch err := f.state.DB.PutInteractionRequest(ctx, intReq); {
	case err == nil:
		// No problem.

	case errors.Is(err, db.ErrAlreadyExists):
		// Already processed this, race
		// condition? Just warn + return.
		log.Warnf(ctx,
			"avoided storing duplicate interaction request %s",
			fpir.intReqURI,
		)
		return nil

	default:
		// Proper DB error.
		return gtserror.Newf(
			"db error storing interaction request %s",
			fpir.intReqURI,
		)
	}

	// Int req is now stored.
	//
	// Set some fields on the
	// pending fave and store it.
	fave.ID = id.NewULID()
	fave.PendingApproval = util.Ptr(true)
	fave.PreApproved = policyCheckResult.AutomaticApproval()

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
	// asynchronously, return 202 Accepted.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APActivityType: ap.ActivityCreate,
		APObjectType:   ap.ActivityLikeRequest,
		GTSModel:       intReq,
		Receiving:      fpir.receiving,
		Requesting:     fpir.requesting,
	})

	return nil
}

func (f *DB) ReplyRequest(ctx context.Context, replyReq vocab.GoToSocialReplyRequest) error {
	log.DebugKV(ctx, "ReplyRequest", serialize{replyReq})

	// Parse out basic interaction request stuff.
	fpir, err := f.parseIntReq(ctx, replyReq)
	if err != nil {
		return err
	}

	// Parse instrument vocab.Type to Statusable.
	statusable, ok := ap.ToStatusable(fpir.instrument)
	if !ok {
		return gtserror.NewWithCode(
			http.StatusBadRequest,
			"could not parse instrument of ReplyRequest to Statusable",
		)
	}

	// Check for spam / relevance.
	ok, err = f.statusableOK(
		ctx,
		fpir.receiving,
		fpir.requesting,
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

	// Make sure we have the status this replies to.
	inReplyTo, err := f.state.DB.GetStatusByURI(ctx, inReplyToURIStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf(
			"db error getting inReplyTo status %s: %w",
			inReplyToURIStr, err,
		)
	}

	if inReplyTo == nil {
		// Status doesn't seem to exist,
		// just drop this ReplyRequest.
		log.Debugf(ctx,
			"got ReplyRequest for non-existent status %s",
			inReplyToURIStr,
		)
		return nil
	}

	// Make sure replied-to status is owned
	// by receiver / target of the ReplyRequest.
	if inReplyTo.AccountURI != fpir.receiving.URI {
		return gtserror.NewfWithCode(
			http.StatusBadRequest,
			"inReplyTo status %s not owned by receiving account %s",
			inReplyToURIStr, fpir.receiving.URI,
		)
	}

	// Make sure reply is attributed to requester.
	attributedToURI, err := ap.ExtractAttributedToURI(statusable)
	if err != nil {
		err := gtserror.Newf("invalid status attributedTo value: %w", err)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}
	attributedToURIStr := attributedToURI.String()

	if attributedToURIStr != fpir.requesting.URI {
		return gtserror.NewfWithCode(
			http.StatusBadRequest,
			"inReplyTo status %s not attributed to requesting account %s",
			inReplyToURIStr, fpir.requesting.URI,
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
		InteractingAccountID:  fpir.requesting.ID,
		InteractingAccount:    fpir.requesting,
		InteractionRequestURI: fpir.intReqURI,
		InteractionURI:        ap.GetJSONLDId(statusable).String(),
		InteractionType:       gtsmodel.InteractionReply,
		Reply:                 nil, // Not settable yet.
	}

	switch err := f.state.DB.PutInteractionRequest(ctx, intReq); {
	case err == nil:
		// No problem.

	case errors.Is(err, db.ErrAlreadyExists):
		// Already processed this, race
		// condition? Just warn + return.
		log.Warnf(ctx,
			"avoided storing duplicate interaction request %s",
			fpir.intReqURI,
		)
		return nil

	default:
		// Proper DB error.
		return gtserror.Newf(
			"db error storing interaction request %s",
			fpir.intReqURI,
		)
	}

	// Further processing will be carried out
	// asynchronously, return 202 Accepted.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APActivityType: ap.ActivityCreate,
		APObjectType:   ap.ActivityReplyRequest,
		GTSModel:       intReq,
		APObject:       statusable,
		Receiving:      fpir.receiving,
		Requesting:     fpir.requesting,
	})

	return nil
}

func (f *DB) AnnounceRequest(ctx context.Context, announceReq vocab.GoToSocialAnnounceRequest) error {
	log.DebugKV(ctx, "AnnounceRequest", serialize{announceReq})

	// Parse out basic interaction request stuff.
	fpir, err := f.parseIntReq(ctx, announceReq)
	if err != nil {
		return err
	}

	// Parse instrument vocab.Type to Announce.
	if fpir.instrument.GetTypeName() != ap.ActivityAnnounce {
		return gtserror.NewWithCode(
			http.StatusBadRequest,
			"instrument of AnnounceRequest was not an Announce",
		)
	}

	announceable, ok := fpir.instrument.(vocab.ActivityStreamsAnnounce)
	if !ok {
		return gtserror.NewWithCode(
			http.StatusBadRequest,
			"could not parse instrument of AnnounceRequest to Announce",
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
		log.Debugf(ctx, "announce %s already stored", boost.URI)
		return nil
	}

	// We must have the boosted status stored.
	targetStatus, err := f.state.DB.GetStatusByURI(ctx, boost.BoostOfURI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf(
			"db error getting announce object %s: %w",
			boost.BoostOfURI, err,
		)
	}

	if targetStatus == nil {
		// Status doesn't seem to exist,
		// just drop this AnnounceRequest.
		log.Debugf(ctx,
			"got AnnounceRequest for non-existent status %s",
			boost.BoostOfURI,
		)
		return nil
	}

	// Status must belong to receiver.
	if targetStatus.AccountID != fpir.receiving.ID {
		return gtserror.NewfWithCode(
			http.StatusBadRequest,
			"announce object %s not owned by receiving account %s",
			boost.BoostOfURI, fpir.receiving.URI,
		)
	}

	// Ensure announce enacted by correct account.
	if boost.AccountID != fpir.requesting.ID {
		return gtserror.NewfWithCode(
			http.StatusForbidden,
			"requester %s is not expected actor %s",
			fpir.requesting.URI, boost.Account.URI,
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
		InteractionRequestURI: fpir.intReqURI,
		InteractionURI:        boost.URI,
		InteractionType:       gtsmodel.InteractionAnnounce,
		Announce:              boost,
	}

	switch err := f.state.DB.PutInteractionRequest(ctx, intReq); {
	case err == nil:
		// No problem.

	case errors.Is(err, db.ErrAlreadyExists):
		// Already processed this, race
		// condition? Just warn + return.
		log.Warnf(ctx,
			"avoided storing duplicate interaction request %s",
			fpir.intReqURI,
		)
		return nil

	default:
		// Proper DB error.
		return gtserror.Newf(
			"db error storing interaction request %s",
			fpir.intReqURI,
		)
	}

	// Further processing will be carried out
	// asynchronously, return 202 Accepted.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APActivityType: ap.ActivityCreate,
		APObjectType:   ap.ActivityAnnounceRequest,
		GTSModel:       intReq,
		Receiving:      fpir.receiving,
		Requesting:     fpir.requesting,
	})

	return nil
}
