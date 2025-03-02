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
	"time"

	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (f *federatingDB) Reject(ctx context.Context, reject vocab.ActivityStreamsReject) error {
	log.DebugKV(ctx, "reject", serialize{reject})

	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	requestingAcct := activityContext.requestingAcct
	receivingAcct := activityContext.receivingAcct

	activityID := ap.GetJSONLDId(reject)
	if activityID == nil {
		// We need an ID.
		const text = "Reject had no id property"
		return gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	for _, object := range ap.ExtractObjects(reject) {
		if asType := object.GetType(); asType != nil {
			// Check and handle any vocab.Type objects.
			switch name := asType.GetTypeName(); name {

			// REJECT FOLLOW
			case ap.ActivityFollow:
				if err := f.rejectFollowType(
					ctx,
					asType,
					receivingAcct,
					requestingAcct,
				); err != nil {
					return err
				}

			// UNHANDLED
			default:
				log.Debugf(ctx, "unhandled object type: %s", name)
			}

		} else if object.IsIRI() {
			// Check and handle any
			// IRI type objects.
			switch objIRI := object.GetIRI(); {

			// REJECT FOLLOW
			case uris.IsFollowPath(objIRI):
				if err := f.rejectFollowIRI(
					ctx,
					objIRI.String(),
					receivingAcct,
					requestingAcct,
				); err != nil {
					return err
				}

			// REJECT STATUS (reply/boost)
			case uris.IsStatusesPath(objIRI):
				if err := f.rejectStatusIRI(
					ctx,
					activityID.String(),
					objIRI.String(),
					receivingAcct,
					requestingAcct,
				); err != nil {
					return err
				}

			// REJECT LIKE
			case uris.IsLikePath(objIRI):
				if err := f.rejectLikeIRI(
					ctx,
					activityID.String(),
					objIRI.String(),
					receivingAcct,
					requestingAcct,
				); err != nil {
					return err
				}

			// UNHANDLED
			default:
				log.Debugf(ctx, "unhandled iri type: %s", objIRI)
			}
		}
	}

	return nil
}

func (f *federatingDB) rejectFollowType(
	ctx context.Context,
	asType vocab.Type,
	receivingAcct *gtsmodel.Account,
	requestingAcct *gtsmodel.Account,
) error {
	// Cast the vocab.Type object to known AS type.
	asFollow := asType.(vocab.ActivityStreamsFollow)

	// Reconstruct the follow.
	follow, err := f.converter.ASFollowToFollow(ctx, asFollow)
	if err != nil {
		err := gtserror.Newf("error converting Follow to *gtsmodel.Follow: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	// Lock on the Follow URI
	// as we may be updating it.
	unlock := f.state.FedLocks.Lock(follow.URI)
	defer unlock()

	// Make sure the creator of the original follow
	// is the same as whatever inbox this landed in.
	if follow.AccountID != receivingAcct.ID {
		const text = "Follow account and inbox account were not the same"
		return gtserror.NewErrorUnprocessableEntity(errors.New(text), text)
	}

	// Make sure the target of the original follow
	// is the same as the account making the request.
	if follow.TargetAccountID != requestingAcct.ID {
		const text = "Follow target account and requesting account were not the same"
		return gtserror.NewErrorForbidden(errors.New(text), text)
	}

	// Reject the follow.
	err = f.state.DB.RejectFollowRequest(
		ctx,
		follow.AccountID,
		follow.TargetAccountID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error rejecting follow request: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}

func (f *federatingDB) rejectFollowIRI(
	ctx context.Context,
	objectIRI string,
	receivingAcct *gtsmodel.Account,
	requestingAcct *gtsmodel.Account,
) error {
	// Lock on this potential Follow
	// URI as we may be updating it.
	unlock := f.state.FedLocks.Lock(objectIRI)
	defer unlock()

	// Get the follow req from the db.
	followReq, err := f.state.DB.GetFollowRequestByURI(ctx, objectIRI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting follow request: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	if followReq == nil {
		// We didn't have a follow request
		// with this URI, so nothing to do.
		// Just return.
		//
		// TODO: Handle Reject Follow to remove
		// an already-accepted follow relationship.
		return nil
	}

	// Make sure the creator of the original follow
	// is the same as whatever inbox this landed in.
	if followReq.AccountID != receivingAcct.ID {
		const text = "Follow account and inbox account were not the same"
		return gtserror.NewErrorUnprocessableEntity(errors.New(text), text)
	}

	// Make sure the target of the original follow
	// is the same as the account making the request.
	if followReq.TargetAccountID != requestingAcct.ID {
		const text = "Follow target account and requesting account were not the same"
		return gtserror.NewErrorForbidden(errors.New(text), text)
	}

	// Reject the follow.
	err = f.state.DB.RejectFollowRequest(
		ctx,
		followReq.AccountID,
		followReq.TargetAccountID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error rejecting follow request: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}

func (f *federatingDB) rejectStatusIRI(
	ctx context.Context,
	activityID string,
	objectIRI string,
	receivingAcct *gtsmodel.Account,
	requestingAcct *gtsmodel.Account,
) error {
	// Lock on this potential status URI.
	unlock := f.state.FedLocks.Lock(objectIRI)
	defer unlock()

	// Get the status from the db.
	status, err := f.state.DB.GetStatusByURI(ctx, objectIRI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting status: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	if status == nil {
		// We didn't have a status with
		// this URI, so nothing to do.
		// Just return.
		return nil
	}

	if !status.IsLocal() {
		// We don't process Rejects of statuses
		// that weren't created on our instance.
		// Just return.
		//
		// TODO: Handle Reject to remove *remote*
		// posts replying-to or boosting the
		// Rejecting account.
		return nil
	}

	// Make sure the creator of the original status
	// is the same as the inbox processing the Reject;
	// this also ensures the status is local.
	if status.AccountID != receivingAcct.ID {
		const text = "status author account and inbox account were not the same"
		return gtserror.NewErrorUnprocessableEntity(errors.New(text), text)
	}

	// Check if we're dealing with a reply
	// or an announce, and make sure the
	// requester is permitted to Reject.
	var apObjectType string
	if status.InReplyToID != "" {
		// Rejecting a Reply.
		apObjectType = ap.ObjectNote
		if status.InReplyToAccountID != requestingAcct.ID {
			const text = "status reply to account and requesting account were not the same"
			return gtserror.NewErrorForbidden(errors.New(text), text)
		}

		// You can't mention an account and then Reject replies from that
		// same account (harassment vector); don't process these Rejects.
		if status.InReplyTo != nil && status.InReplyTo.MentionsAccount(status.AccountID) {
			const text = "refusing to process Reject of a reply from a mentioned account"
			return gtserror.NewErrorForbidden(errors.New(text), text)
		}

	} else {
		// Rejecting an Announce.
		apObjectType = ap.ActivityAnnounce
		if status.BoostOfAccountID != requestingAcct.ID {
			const text = "status boost of account and requesting account were not the same"
			return gtserror.NewErrorForbidden(errors.New(text), text)
		}
	}

	// Check if there's an interaction request in the db for this status.
	req, err := f.state.DB.GetInteractionRequestByInteractionURI(ctx, status.URI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting interaction request: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	switch {
	case req == nil:
		// No interaction request existed yet for this
		// status, create a pre-rejected request now.
		req = &gtsmodel.InteractionRequest{
			ID:                   id.NewULID(),
			TargetAccountID:      requestingAcct.ID,
			TargetAccount:        requestingAcct,
			InteractingAccountID: receivingAcct.ID,
			InteractingAccount:   receivingAcct,
			InteractionURI:       status.URI,
			URI:                  activityID,
			RejectedAt:           time.Now(),
		}

		if apObjectType == ap.ObjectNote {
			// Reply.
			req.InteractionType = gtsmodel.InteractionReply
			req.StatusID = status.InReplyToID
			req.Status = status.InReplyTo
			req.Reply = status
		} else {
			// Announce.
			req.InteractionType = gtsmodel.InteractionAnnounce
			req.StatusID = status.BoostOfID
			req.Status = status.BoostOf
			req.Announce = status
		}

		if err := f.state.DB.PutInteractionRequest(ctx, req); err != nil {
			err := gtserror.Newf("db error inserting interaction request: %w", err)
			return gtserror.NewErrorInternalError(err)
		}

	case req.IsRejected():
		// Interaction has already been rejected. Just
		// update to this Reject URI and then return early.
		req.URI = activityID
		if err := f.state.DB.UpdateInteractionRequest(ctx, req, "uri"); err != nil {
			err := gtserror.Newf("db error updating interaction request: %w", err)
			return gtserror.NewErrorInternalError(err)
		}
		return nil

	default:
		// Mark existing interaction request as
		// Rejected, even if previously Accepted.
		req.AcceptedAt = time.Time{}
		req.RejectedAt = time.Now()
		req.URI = activityID
		if err := f.state.DB.UpdateInteractionRequest(ctx, req,
			"accepted_at",
			"rejected_at",
			"uri",
		); err != nil {
			err := gtserror.Newf("db error updating interaction request: %w", err)
			return gtserror.NewErrorInternalError(err)
		}
	}

	// Send the rejected request through to
	// the fedi worker to process side effects.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   apObjectType,
		APActivityType: ap.ActivityReject,
		GTSModel:       req,
		Receiving:      receivingAcct,
		Requesting:     requestingAcct,
	})

	return nil
}

func (f *federatingDB) rejectLikeIRI(
	ctx context.Context,
	activityID string,
	objectIRI string,
	receivingAcct *gtsmodel.Account,
	requestingAcct *gtsmodel.Account,
) error {
	// Lock on this potential Like
	// URI as we may be updating it.
	unlock := f.state.FedLocks.Lock(objectIRI)
	defer unlock()

	// Get the fave from the db.
	fave, err := f.state.DB.GetStatusFaveByURI(ctx, objectIRI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting fave: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	if fave == nil {
		// We didn't have a fave with
		// this URI, so nothing to do.
		// Just return.
		return nil
	}

	if !fave.Account.IsLocal() {
		// We don't process Rejects of Likes
		// that weren't created on our instance.
		// Just return.
		//
		// TODO: Handle Reject to remove *remote*
		// likes targeting the Rejecting account.
		return nil
	}

	// Make sure the creator of the original Like
	// is the same as the inbox processing the Reject;
	// this also ensures the Like is local.
	if fave.AccountID != receivingAcct.ID {
		const text = "fave creator account and inbox account were not the same"
		return gtserror.NewErrorUnprocessableEntity(errors.New(text), text)
	}

	// Make sure the target of the Like is the
	// same as the account doing the Reject.
	if fave.TargetAccountID != requestingAcct.ID {
		const text = "status fave target account and requesting account were not the same"
		return gtserror.NewErrorForbidden(errors.New(text), text)
	}

	// Check if there's an interaction request in the db for this like.
	req, err := f.state.DB.GetInteractionRequestByInteractionURI(ctx, fave.URI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting interaction request: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	switch {
	case req == nil:
		// No interaction request existed yet for this
		// fave, create a pre-rejected request now.
		req = &gtsmodel.InteractionRequest{
			ID:                   id.NewULID(),
			TargetAccountID:      requestingAcct.ID,
			TargetAccount:        requestingAcct,
			InteractingAccountID: receivingAcct.ID,
			InteractingAccount:   receivingAcct,
			InteractionURI:       fave.URI,
			InteractionType:      gtsmodel.InteractionLike,
			Like:                 fave,
			URI:                  activityID,
			RejectedAt:           time.Now(),
		}

		if err := f.state.DB.PutInteractionRequest(ctx, req); err != nil {
			err := gtserror.Newf("db error inserting interaction request: %w", err)
			return gtserror.NewErrorInternalError(err)
		}

	case req.IsRejected():
		// Interaction has already been rejected. Just
		// update to this Reject URI and then return early.
		req.URI = activityID
		if err := f.state.DB.UpdateInteractionRequest(ctx, req, "uri"); err != nil {
			err := gtserror.Newf("db error updating interaction request: %w", err)
			return gtserror.NewErrorInternalError(err)
		}
		return nil

	default:
		// Mark existing interaction request as
		// Rejected, even if previously Accepted.
		req.AcceptedAt = time.Time{}
		req.RejectedAt = time.Now()
		req.URI = activityID
		if err := f.state.DB.UpdateInteractionRequest(ctx, req,
			"accepted_at",
			"rejected_at",
			"uri",
		); err != nil {
			err := gtserror.Newf("db error updating interaction request: %w", err)
			return gtserror.NewErrorInternalError(err)
		}
	}

	// Send the rejected request through to
	// the fedi worker to process side effects.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActivityLike,
		APActivityType: ap.ActivityReject,
		GTSModel:       req,
		Receiving:      receivingAcct,
		Requesting:     requestingAcct,
	})

	return nil
}
