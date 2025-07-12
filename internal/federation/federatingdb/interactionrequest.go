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
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

// firstPassIntReq represents a partially-parsed
// interaction request returned from the util
// function parseInteractionReq.
type firstPassIntReq struct {
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
	// Mark activity as handled.
	f.storeActivityID(intReq)

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
		err := gtserror.Newf("invalid object len %d, wanted 1", l)
		return nil, gtserror.WrapWithCode(http.StatusBadRequest, err)
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
		err := gtserror.Newf("invalid instrument len %d, wanted 1", l)
		return nil, gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	// Instrument should be a
	// type and not just an IRI.
	instrument := instruments[0].GetType()
	if instrument == nil {
		err := gtserror.New("instrument was not a type")
		return nil, gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	return &firstPassIntReq{
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
		err := gtserror.New("instrument of LikeRequest was not a Like")
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	likeable, ok := fpir.instrument.(vocab.ActivityStreamsLike)
	if !ok {
		err := gtserror.New("could not parse instrument of LikeRequest to Like")
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	// Convert received AS like type to internal fave model.
	fave, err := f.converter.ASLikeToFave(ctx, likeable)
	if err != nil {
		err := gtserror.Newf("error converting from AS type: %w", err)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	// Ensure fave enacted by correct account.
	if fave.AccountID != fpir.requesting.ID {
		return gtserror.NewfWithCode(http.StatusForbidden, "requester %s is not expected actor %s",
			fpir.requesting.URI, fave.Account.URI)
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

	

	return nil
}

func (f *DB) ReplyRequest(ctx context.Context, replyReq vocab.GoToSocialReplyRequest) error {
	return nil
}

func (f *DB) AnnounceRequest(ctx context.Context, announceReq vocab.GoToSocialAnnounceRequest) error {
	return nil
}
