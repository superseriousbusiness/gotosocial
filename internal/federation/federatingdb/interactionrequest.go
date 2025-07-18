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

	return &firstPassIntReq{
		intReqURI:  ap.GetJSONLDId(intReq).String(),
		requesting: requesting,
		receiving:  receiving,
		object:     status,
		instrument: instrument,
	}, nil
}

func (f *DB) LikeRequest(ctx context.Context, likeReq vocab.GoToSocialLikeRequest) error {
	log.DebugKV(ctx, "like", serialize{likeReq})

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
	// Set some fields on the fave.
	fave.ID = id.NewULID()

	// Politely-requested interactions always
	// necessitate sending a response.
	fave.PendingApproval = util.Ptr(true)

	// If it's pre-approved the processor can
	// immediately send out an Accept for it.
	fave.PreApproved = policyCheckResult.AutomaticApproval()

	// Further processing will be carried out
	// asynchronously, return 202 Accepted.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActivityLike,
		APActivityType: ap.ActivityLikeRequest,
		GTSModel:       intReq,
		Receiving:      fpir.receiving,
		Requesting:     fpir.requesting,
	})

	return nil
}

func (f *DB) ReplyRequest(ctx context.Context, replyReq vocab.GoToSocialReplyRequest) error {
	return nil
}

func (f *DB) AnnounceRequest(ctx context.Context, announceReq vocab.GoToSocialAnnounceRequest) error {
	return nil
}
