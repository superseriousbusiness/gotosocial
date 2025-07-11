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
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

func (f *DB) LikeRequest(ctx context.Context, likeReq vocab.GoToSocialLikeRequest) error {
	log.DebugKV(ctx, "like", serialize{likeReq})

	// Mark activity as handled.
	f.storeActivityID(likeReq)

	// Extract relevant values from passed ctx.
	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	requesting := activityContext.requestingAcct
	receiving := activityContext.receivingAcct

	if requesting.IsMoving() {
		// A Moving account
		// can't do this.
		return nil
	}

	if receiving.IsMoving() {
		// Moving accounts can't
		// do anything with interaction
		// requests, so ignore it.
		return nil
	}

	// Make sure we have a single
	// object of the interaction request.
	objectIRIs := ap.GetObjectIRIs(likeReq)
	if l := len(objectIRIs); l != 1 {
		err := gtserror.Newf("invalid object len %d, wanted 1", l)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}
	statusIRI := objectIRIs[0]
	statusIRIStr := statusIRI.String()

	// Object should be a status.
	status, err := f.state.DB.GetStatusByURI(ctx, statusIRIStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting like object status %s: %w", statusIRIStr, err)
		return err
	}

	if status == nil {
		// Status doesn't exist
		// (anymore); do nothing.
		return nil
	}

	// Ensure like req received by correct account.
	if status.AccountID != receiving.ID {
		err := gtserror.NewfWithCode(
			http.StatusForbidden,
			"receiver %s is not owner of like-requested status",
			receiving.URI,
		)
		return err
	}

	// We should have one instrument,
	// and it should be a fave.
	instrs := ap.ExtractInstruments(likeReq)
	if l := len(instrs); l != 1 {
		err := gtserror.Newf("invalid instrument len %d, wanted 1", l)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	// Ensure type and not just IRI.
	instrType := instrs[0].GetType()
	if instrType == nil {
		err := gtserror.New("instrument was not a type")
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	// Make sure it's a Like.
	if instrType.GetTypeName() != ap.ActivityLike {
		err := gtserror.New("instrument type was not Like")
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	likeable, ok := instrType.(vocab.ActivityStreamsLike)
	if !ok {
		err := gtserror.New("instrument was not a Like")
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	// Convert received AS like type to internal fave model.
	fave, err := f.converter.ASLikeToFave(ctx, likeable)
	if err != nil {
		err := gtserror.Newf("error converting from AS type: %w", err)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	// Ensure fave enacted by correct account.
	if fave.AccountID != requesting.ID {
		return gtserror.NewfWithCode(http.StatusForbidden, "requester %s is not expected actor %s",
			requesting.URI, fave.Account.URI)
	}

	// Ensure fave received by correct account.
	if fave.TargetAccountID != receiving.ID {
		return gtserror.NewfWithCode(http.StatusForbidden, "receiver %s is not expected object %s",
			receiving.URI, fave.TargetAccount.URI)
	}

	// Make sure Like target is the same as the 

	return nil
}

func (f *DB) ReplyRequest(ctx context.Context, replyReq vocab.GoToSocialReplyRequest) error {
	return nil
}

func (f *DB) AnnounceRequest(ctx context.Context, announceReq vocab.GoToSocialAnnounceRequest) error {
	return nil
}
