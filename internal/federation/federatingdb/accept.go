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
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (f *federatingDB) GetAccept(
	ctx context.Context,
	acceptIRI *url.URL,
) (vocab.ActivityStreamsAccept, error) {
	approval, err := f.state.DB.GetInteractionRequestByURI(ctx, acceptIRI.String())
	if err != nil {
		return nil, err
	}
	return f.converter.InteractionReqToASAccept(ctx, approval)
}

func (f *federatingDB) Accept(ctx context.Context, accept vocab.ActivityStreamsAccept) error {
	log.DebugKV(ctx, "accept", serialize{accept})

	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	requestingAcct := activityContext.requestingAcct
	receivingAcct := activityContext.receivingAcct

	if requestingAcct.IsMoving() {
		// A Moving account
		// can't do this.
		return nil
	}

	acceptID := ap.GetJSONLDId(accept)
	if acceptID == nil {
		// We need an ID.
		const text = "Accept had no id property"
		return gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Ensure requester is the same as the
	// Actor of the Accept; you can't Accept
	// something on someone else's behalf.
	actorURI, err := ap.ExtractActorURI(accept)
	if err != nil {
		const text = "Accept had empty or invalid actor property"
		return gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	if requestingAcct.URI != actorURI.String() {
		const text = "Accept actor and requesting account were not the same"
		return gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Iterate all provided objects in the activity,
	// handling the ones we know how to handle.
	for _, object := range ap.ExtractObjects(accept) {
		if asType := object.GetType(); asType != nil {
			// Check and handle any vocab.Type objects.
			switch name := asType.GetTypeName(); {

			// ACCEPT FOLLOW
			case name == ap.ActivityFollow:
				if err := f.acceptFollowType(
					ctx,
					asType,
					receivingAcct,
					requestingAcct,
				); err != nil {
					return err
				}

			// ACCEPT TYPE-HINTED LIKE
			//
			// ie., a Like with just `id`
			// and `type` properties set.
			case name == ap.ActivityLike:
				objIRI := ap.GetJSONLDId(asType)
				if objIRI == nil {
					log.Debugf(ctx, "could not retrieve id of inlined Accept object %s", name)
					continue
				}

				if err := f.acceptLikeIRI(
					ctx,
					acceptID,
					accept,
					objIRI.String(),
					receivingAcct,
					requestingAcct,
				); err != nil {
					return err
				}

			// ACCEPT TYPE-HINTED REPLY OR ANNOUNCE.
			//
			// ie., a statusable or Announce with
			// just `id` and `type` properties set.
			case name == ap.ActivityAnnounce || ap.IsStatusable(name):
				objIRI := ap.GetJSONLDId(asType)
				if objIRI == nil {
					log.Debugf(ctx, "could not retrieve id of inlined Accept object %s", name)
					continue
				}

				if err := f.acceptOtherIRI(
					ctx,
					acceptID,
					accept,
					objIRI,
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

			// ACCEPT FOLLOW
			case uris.IsFollowPath(objIRI):
				if err := f.acceptFollowIRI(
					ctx,
					objIRI.String(),
					receivingAcct,
					requestingAcct,
				); err != nil {
					return err
				}

			// ACCEPT LIKE
			case uris.IsLikePath(objIRI):
				if err := f.acceptLikeIRI(
					ctx,
					acceptID,
					accept,
					objIRI.String(),
					receivingAcct,
					requestingAcct,
				); err != nil {
					return err
				}

			// ACCEPT OTHER (reply? announce?)
			//
			// Don't check on IsStatusesPath
			// as this may be a remote status.
			default:
				if err := f.acceptOtherIRI(
					ctx,
					acceptID,
					accept,
					objIRI,
					receivingAcct,
					requestingAcct,
				); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (f *federatingDB) acceptFollowType(
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

	// Accept and get the populated follow back.
	follow, err = f.state.DB.AcceptFollowRequest(
		ctx,
		follow.AccountID,
		follow.TargetAccountID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error accepting follow request: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	if follow == nil {
		// There was no follow request
		// to accept, just return 202.
		return nil
	}

	// Send the accepted follow through
	// the processor to do side effects.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActivityFollow,
		APActivityType: ap.ActivityAccept,
		GTSModel:       follow,
		Receiving:      receivingAcct,
		Requesting:     requestingAcct,
	})

	return nil
}

func (f *federatingDB) acceptFollowIRI(
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

	// Accept and get the populated follow back.
	follow, err := f.state.DB.AcceptFollowRequest(
		ctx,
		followReq.AccountID,
		followReq.TargetAccountID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error accepting follow request: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	if follow == nil {
		// There was no follow request
		// to accept, just return 202.
		return nil
	}

	// Send the accepted follow through
	// the processor to do side effects.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActivityFollow,
		APActivityType: ap.ActivityAccept,
		GTSModel:       follow,
		Receiving:      receivingAcct,
		Requesting:     requestingAcct,
	})

	return nil
}

func (f *federatingDB) acceptOtherIRI(
	ctx context.Context,
	acceptID *url.URL,
	accept vocab.ActivityStreamsAccept,
	objectIRI *url.URL,
	receivingAcct *gtsmodel.Account,
	requestingAcct *gtsmodel.Account,
) error {
	// See if we can get a status from the db.
	status, err := f.state.DB.GetStatusByURI(ctx, objectIRI.String())
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting status: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	if status != nil {
		// We had a status stored with this
		// objectIRI, proceed to accept it.
		return f.acceptStoredStatus(
			ctx,
			acceptID,
			accept,
			status,
			receivingAcct,
			requestingAcct,
		)
	}

	if objectIRI.Host == config.GetHost() ||
		objectIRI.Host == config.GetAccountDomain() {
		// Claims to be Accepting something of ours,
		// but we don't have a status stored for this
		// URI, so most likely it's been deleted in
		// the meantime, just bail.
		return nil
	}

	// This must be an Accept of a remote Activity
	// or Object. Ensure relevance of this message
	// by checking that receiver follows requester.
	following, err := f.state.DB.IsFollowing(
		ctx,
		receivingAcct.ID,
		requestingAcct.ID,
	)
	if err != nil {
		err := gtserror.Newf("db error checking following: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	if !following {
		// If we don't follow this person, and
		// they're not Accepting something we know
		// about, then we don't give a good goddamn.
		return nil
	}

	// This may be a reply, or it may be a boost,
	// we can't know yet without dereferencing it,
	// but let the processor worry about that.
	//
	// TODO: do something with type hinting here.
	apObjectType := ap.ObjectUnknown

	// Extract appropriate approvedByURI from the Accept.
	approvedByURI, err := approvedByURI(acceptID, accept)
	if err != nil {
		return gtserror.NewErrorForbidden(err, err.Error())
	}

	// Pass to the processor and let them handle side effects.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   apObjectType,
		APActivityType: ap.ActivityAccept,
		APIRI:          approvedByURI,
		APObject:       objectIRI,
		Receiving:      receivingAcct,
		Requesting:     requestingAcct,
	})

	return nil
}

func (f *federatingDB) acceptStoredStatus(
	ctx context.Context,
	acceptID *url.URL,
	accept vocab.ActivityStreamsAccept,
	status *gtsmodel.Status,
	receivingAcct *gtsmodel.Account,
	requestingAcct *gtsmodel.Account,
) error {
	// Lock on this status URI
	// as we may be updating it.
	unlock := f.state.FedLocks.Lock(status.URI)
	defer unlock()

	pendingApproval := util.PtrOrValue(status.PendingApproval, false)
	if !pendingApproval {
		// Status doesn't need approval or it's
		// already been approved by an Accept.
		// Just return.
		return nil
	}

	// Make sure the target of the interaction (reply/boost)
	// is the same as the account doing the Accept.
	if status.BoostOfAccountID != requestingAcct.ID &&
		status.InReplyToAccountID != requestingAcct.ID {
		const text = "status reply to or boost of account and requesting account were not the same"
		return gtserror.NewErrorForbidden(errors.New(text), text)
	}

	// Extract appropriate approvedByURI from the Accept.
	approvedByURI, err := approvedByURI(acceptID, accept)
	if err != nil {
		return gtserror.NewErrorForbidden(err, err.Error())
	}

	// Mark the status as approved by this URI.
	status.PendingApproval = util.Ptr(false)
	status.ApprovedByURI = approvedByURI.String()
	if err := f.state.DB.UpdateStatus(
		ctx,
		status,
		"pending_approval",
		"approved_by_uri",
	); err != nil {
		err := gtserror.Newf("db error accepting status: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	var apObjectType string
	if status.InReplyToID != "" {
		// Accepting a Reply.
		apObjectType = ap.ObjectNote
	} else {
		// Accepting an Announce.
		apObjectType = ap.ActivityAnnounce
	}

	// Send the now-approved status through to the
	// fedi worker again to process side effects.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   apObjectType,
		APActivityType: ap.ActivityAccept,
		GTSModel:       status,
		Receiving:      receivingAcct,
		Requesting:     requestingAcct,
	})

	return nil
}

func (f *federatingDB) acceptLikeIRI(
	ctx context.Context,
	acceptID *url.URL,
	accept vocab.ActivityStreamsAccept,
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
		// We don't process Accepts of Likes
		// that weren't created on our instance.
		// Just return.
		return nil
	}

	pendingApproval := util.PtrOrValue(fave.PendingApproval, false)
	if !pendingApproval {
		// Like doesn't need approval or it's
		// already been approved by an Accept.
		// Just return.
		return nil
	}

	// Make sure the creator of the original Like
	// is the same as the inbox processing the Accept;
	// this also ensures the Like is local.
	if fave.AccountID != receivingAcct.ID {
		const text = "fave creator account and inbox account were not the same"
		return gtserror.NewErrorUnprocessableEntity(errors.New(text), text)
	}

	// Make sure the target of the Like is the
	// same as the account doing the Accept.
	if fave.TargetAccountID != requestingAcct.ID {
		const text = "status fave target account and requesting account were not the same"
		return gtserror.NewErrorForbidden(errors.New(text), text)
	}

	// Extract appropriate approvedByURI from the Accept.
	approvedByURI, err := approvedByURI(acceptID, accept)
	if err != nil {
		return gtserror.NewErrorForbidden(err, err.Error())
	}

	// Mark the fave as approved by this URI.
	fave.PendingApproval = util.Ptr(false)
	fave.ApprovedByURI = approvedByURI.String()
	if err := f.state.DB.UpdateStatusFave(
		ctx,
		fave,
		"pending_approval",
		"approved_by_uri",
	); err != nil {
		err := gtserror.Newf("db error accepting status: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	// Send the now-approved fave through to the
	// fedi worker again to process side effects.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActivityLike,
		APActivityType: ap.ActivityAccept,
		GTSModel:       fave,
		Receiving:      receivingAcct,
		Requesting:     requestingAcct,
	})

	return nil
}

// approvedByURI extracts the appropriate *url.URL
// to use as an interaction's approvedBy value by
// checking to see if the Accept has a result URL set.
// If that result URL exists, is an IRI (not a type),
// and is on the same host as the Accept ID, then the
// result URI will be returned. In all other cases,
// the Accept ID is returned unchanged.
//
// Error is only returned if the result URI is set
// but the host differs from the Accept ID host.
//
// TODO: This function should be updated at some point
// to check for inlined result type, and see if type is
// a LikeApproval, ReplyApproval, or AnnounceApproval,
// and check the attributedTo, object, and target of
// the approval as well. But this'll do for now.
func approvedByURI(
	acceptID *url.URL,
	accept vocab.ActivityStreamsAccept,
) (*url.URL, error) {
	// Check if the Accept has a `result` property
	// set on it (which should be an approval).
	resultProp := accept.GetActivityStreamsResult()
	if resultProp == nil {
		// No result,
		// use AcceptID.
		return acceptID, nil
	}

	if resultProp.Len() != 1 {
		// Result was unexpected
		// length, can't use this.
		return acceptID, nil
	}

	result := resultProp.At(0)
	if result == nil {
		// Result entry
		// was nil, huh!
		return acceptID, nil
	}

	if !result.IsIRI() {
		// Can't handle
		// inlined yet.
		return acceptID, nil
	}

	resultIRI := result.GetIRI()
	if resultIRI == nil {
		// Result entry
		// was nil, huh!
		return acceptID, nil
	}

	if resultIRI.Host != acceptID.Host {
		// What the boobs is this?
		err := fmt.Errorf(
			"host of result %s differed from host of Accept %s",
			resultIRI, accept,
		)
		return nil, err
	}

	// Use the result IRI we've been
	// given instead of the acceptID.
	return resultIRI, nil
}
