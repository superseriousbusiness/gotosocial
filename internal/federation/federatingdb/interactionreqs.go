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
	"net/http"
	"time"

	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
)

func (f *DB) LikeRequest(
	ctx context.Context,
	likeReq vocab.GoToSocialLikeRequest,
) error {
	log.DebugKV(ctx, "LikeRequest", serialize{likeReq})

	// Mark activity as handled.
	f.storeActivityID(likeReq)

	// Extract relevant values from passed ctx.
	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	requesting := activityContext.requestingAcct
	receiving := activityContext.receivingAcct

	if receiving.IsMoving() {
		// A Moving account
		// can't accept a Like.
		return nil
	}

	// Convert received LikeRequest type to dummy
	// fave, so that we can check against policies.
	// This dummy won't be stored in the database,
	// it's used purely for doing permission checks.
	dummyFave, err := f.converter.ASLikeToFave(ctx, likeReq)
	if err != nil {
		err := gtserror.Newf("error converting from AS type: %w", err)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	if !*dummyFave.Status.Local {
		// Only process like requests for local statuses.
		//
		// If the remote has sent us a like request for a
		// status that's not ours, we should ignore it.
		return nil
	}

	// Ensure fave would be enacted by correct account.
	if dummyFave.AccountID != requesting.ID {
		return gtserror.NewfWithCode(http.StatusForbidden, "requester %s is not expected actor %s",
			requesting.URI, dummyFave.Account.URI)
	}

	// Ensure fave would be received by correct account.
	if dummyFave.TargetAccountID != receiving.ID {
		return gtserror.NewfWithCode(http.StatusForbidden, "receiver %s is not expected object %s",
			receiving.URI, dummyFave.TargetAccount.URI)
	}

	// Check how we should handle this request.
	policyResult, err := f.intFilter.StatusLikeable(ctx,
		requesting,
		dummyFave.Status,
	)
	if err != nil {
		return gtserror.Newf("error seeing if status %s is likeable: %w", dummyFave.Status.URI, err)
	}

	// Determine whether to automatically accept,
	// automatically reject, or pend approval.
	var (
		acceptedAt time.Time
		rejectedAt time.Time
	)
	if policyResult.AutomaticApproval() {
		acceptedAt = time.Now()
	} else if policyResult.Forbidden() {
		rejectedAt = time.Now()
	}

	interactionReq := &gtsmodel.InteractionRequest{
		ID:                   id.NewULID(),
		StatusID:             dummyFave.Status.ID,
		Status:               dummyFave.Status,
		TargetAccountID:      receiving.ID,
		TargetAccount:        receiving,
		InteractingAccountID: requesting.ID,
		InteractingAccount:   requesting,
		InteractionURI:       dummyFave.URI,
		InteractionType:      gtsmodel.InteractionLikeRequest,
		AcceptedAt:           acceptedAt,
		RejectedAt:           rejectedAt,

		// Empty as reject/accept
		// response not yet sent.
		URI: "",
	}

	// Send the interactionReq through to
	// the processor to handle side effects.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActivityLikeRequest,
		APActivityType: ap.ActivityCreate,
		GTSModel:       interactionReq,
		Receiving:      receiving,
		Requesting:     requesting,
	})

	return nil
}

func (f *DB) ReplyRequest(
	ctx context.Context,
	replyReq vocab.GoToSocialReplyRequest,
) error {
	return nil
}

func (f *DB) AnnounceRequest(
	ctx context.Context,
	announceReq vocab.GoToSocialAnnounceRequest,
) error {
	return nil
}
