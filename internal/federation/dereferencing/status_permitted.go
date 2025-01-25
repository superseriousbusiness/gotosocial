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

package dereferencing

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// isPermittedStatus returns whether the given status
// is permitted to be stored on this instance, checking:
//
//   - author is not suspended
//   - status passes visibility checks
//   - status passes interaction policy checks
//
// If status is not permitted to be stored, the function
// will clean up after itself by removing the status.
//
// If status is a reply or a boost, and the author of
// the given status is only permitted to reply or boost
// pending approval, then "PendingApproval" will be set
// to "true" on status. Callers should check this
// and handle it as appropriate.
//
// If status is a reply that is not permitted based on
// interaction policies, or status replies to a status
// that's been Rejected before (ie., it has a rejected
// InteractionRequest stored in the db) then the reply
// will also be rejected, and a pre-rejected interaction
// request will be stored for it before doing cleanup,
// if one didn't already exist.
func (d *Dereferencer) isPermittedStatus(
	ctx context.Context,
	requestUser string,
	existing *gtsmodel.Status,
	status *gtsmodel.Status,
	isNew bool,
) (
	permitted bool, // is permitted?
	err error,
) {
	switch {
	case status.Account.IsSuspended():
		// we shouldn't reach this point, log to poke devs to investigate.
		log.Warnf(ctx, "status author suspended: %s", status.AccountURI)
		permitted = false

	case status.InReplyToURI != "":
		// Status is a reply, check permissivity.
		permitted, err = d.isPermittedReply(ctx,
			requestUser,
			status,
		)
		if err != nil {
			return false, gtserror.Newf("error checking reply permissivity: %w", err)
		}

	case status.BoostOf != nil:
		// Status is a boost, check permissivity.
		permitted, err = d.isPermittedBoost(ctx,
			requestUser,
			status,
		)
		if err != nil {
			return false, gtserror.Newf("error checking boost permissivity: %w", err)
		}

	default:
		// In all other cases
		// permit this status.
		permitted = true
	}

	if !permitted && !isNew {
		log.Infof(ctx, "deleting unpermitted: %s", existing.URI)

		// Delete existing status from database as it's no longer permitted.
		if err := d.state.DB.DeleteStatusByID(ctx, existing.ID); err != nil {
			log.Errorf(ctx, "error deleting %s after permissivity fail: %v", existing.URI, err)
		}
	}

	return
}

// isPermittedReply ...
func (d *Dereferencer) isPermittedReply(
	ctx context.Context,
	requestUser string,
	reply *gtsmodel.Status,
) (bool, error) {

	var (
		replyURI     = reply.URI           // Definitely set.
		inReplyToURI = reply.InReplyToURI  // Definitely set.
		inReplyTo    = reply.InReplyTo     // Might not be set.
		acceptIRI    = reply.ApprovedByURI // Might not be set.
	)

	// Check if we have a stored interaction request for parent status.
	parentReq, err := d.state.DB.GetInteractionRequestByInteractionURI(
		gtscontext.SetBarebones(ctx),
		inReplyToURI,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting interaction request: %w", err)
		return false, err
	}

	// Check if we have a stored interaction request for this reply.
	thisReq, err := d.state.DB.GetInteractionRequestByInteractionURI(
		gtscontext.SetBarebones(ctx),
		replyURI,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting interaction request: %w", err)
		return false, err
	}

	parentRejected := (parentReq != nil && parentReq.IsRejected())
	thisRejected := (thisReq != nil && thisReq.IsRejected())

	if parentRejected {
		// If this status's parent was rejected,
		// implicitly this reply should be too;
		// there's nothing more to check here.
		return false, d.unpermittedByParent(ctx,
			reply,
			thisReq,
			parentReq,
		)
	}

	// Parent wasn't rejected. Check if this
	// reply itself was rejected previously.
	//
	// If it was, and it doesn't now claim to
	// be approved, then we should just reject it
	// again, as nothing's changed since last time.
	if thisRejected && acceptIRI == "" {

		// Nothing changed,
		// still rejected.
		return false, nil
	}

	// This reply wasn't rejected previously, or
	// it was rejected previously and now claims
	// to be approved. Continue permission checks.

	if inReplyTo == nil {

		// If we didn't have the replied-to status
		// in our database (yet), we can't check
		// right now if this reply is permitted.
		//
		// For now, just return permitted if reply
		// was not explicitly rejected before; worst-
		// case, the reply stays on the instance for
		// a couple hours until we try to deref it
		// again and realize it should be forbidden.
		return !thisRejected, nil
	}

	// We have the replied-to status; ensure it's fully populated.
	if err := d.state.DB.PopulateStatus(ctx, inReplyTo); err != nil {
		return false, gtserror.Newf("error populating status %s: %w", reply.ID, err)
	}

	// Make sure replied-to status is not
	// a boost wrapper, and make sure it's
	// actually visible to the requester.
	if inReplyTo.BoostOfID != "" {
		// We do not permit replies
		// to boost wrapper statuses.
		log.Info(ctx, "rejecting reply to boost wrapper status")
		return false, nil
	}

	if inReplyTo.IsLocal() {
		visible, err := d.visFilter.StatusVisible(ctx,
			reply.Account,
			inReplyTo,
		)
		if err != nil {
			err := gtserror.Newf("error checking inReplyTo visibility: %w", err)
			return false, err
		}

		// Our status is not visible to the
		// account trying to do the reply.
		if !visible {
			return false, nil
		}
	}

	// If this reply claims to be approved,
	// validate this by dereferencing the
	// Accept and checking the return value.
	// No further checks are required.
	if acceptIRI != "" {
		return d.isPermittedByAcceptIRI(
			ctx,
			requestUser,
			reply,
			inReplyTo,
			thisReq,
			acceptIRI,
		)
	}

	// Status doesn't claim to be approved.
	// Check interaction policy of inReplyTo
	// to see what we need to do with it.
	replyable, err := d.intFilter.StatusReplyable(ctx,
		reply.Account,
		inReplyTo,
	)
	if err != nil {
		err := gtserror.Newf("error checking status replyability: %w", err)
		return false, err
	}

	if replyable.Forbidden() {
		// Reply is not permitted according to policy.
		//
		// Either insert a pre-rejected interaction
		// req into the db, or update the existing
		// one, and return. This ensures that replies
		// to this rejected reply also aren't permitted.
		return false, d.rejectedByPolicy(
			ctx,
			reply,
			inReplyTo,
			thisReq,
		)
	}

	if replyable.Permitted() &&
		!replyable.MatchedOnCollection() {
		// Reply is permitted and match was *not* made
		// based on inclusion in a followers/following
		// collection. Just permit the reply full stop
		// as no approval / accept URI is necessary.
		return true, nil
	}

	// Reply is either permitted based on inclusion in a
	// followers/following collection, *or* is permitted
	// pending approval, though we know at this point
	// that the status did not include an approvedBy URI.

	if !inReplyTo.IsLocal() {
		// If the replied-to status is remote, we should just
		// drop this reply at this point, as we can't verify
		// that the remote replied-to account approves it, and
		// we can't verify the presence of a remote account
		// in one of another remote account's collections.
		//
		// It's possible we'll get an Accept from the replied-
		// to account later, and we can store this reply then.
		return false, nil
	}

	// Replied-to status is ours, so the
	// replied-to account is ours as well.

	if replyable.MatchedOnCollection() {
		// If permission was granted based on inclusion in
		// a followers/following collection, pre-approve the
		// reply, as we ourselves can validate presence of the
		// replier in the appropriate collection. Pre-approval
		// lets the processor know it should send out an Accept
		// straight away on behalf of the replied-to account.
		reply.PendingApproval = util.Ptr(true)
		reply.PreApproved = true
		return true, nil
	}

	// Reply just requires approval from the local account
	// it replies to. Set PendingApproval so the processor
	// knows to create a pending interaction request.
	reply.PendingApproval = util.Ptr(true)
	return true, nil
}

// unpermittedByParent marks the given reply as rejected
// based on the fact that its parent was rejected.
//
// This will create a rejected interaction request for
// the status in the db, if one didn't exist already,
// or update an existing interaction request instead.
func (d *Dereferencer) unpermittedByParent(
	ctx context.Context,
	reply *gtsmodel.Status,
	thisReq *gtsmodel.InteractionRequest,
	parentReq *gtsmodel.InteractionRequest,
) error {
	if thisReq != nil && thisReq.IsRejected() {
		// This interaction request is
		// already marked as rejected,
		// there's nothing more to do.
		return nil
	}

	if thisReq != nil {
		// Before we return, ensure interaction
		// request is marked as rejected.
		thisReq.RejectedAt = time.Now()
		thisReq.AcceptedAt = time.Time{}
		err := d.state.DB.UpdateInteractionRequest(
			ctx,
			thisReq,
			"rejected_at",
			"accepted_at",
		)
		if err != nil {
			return gtserror.Newf("db error updating interaction request: %w", err)
		}

		return nil
	}

	// We haven't stored a rejected interaction
	// request for this status yet, do it now.
	rejectID := id.NewULID()

	// To ensure the Reject chain stays coherent,
	// borrow fields from the up-thread rejection.
	// This collapses the chain beyond the first
	// rejected reply and allows us to avoid derefing
	// further replies we already know we don't want.
	inReplyToID := parentReq.StatusID
	targetAccountID := parentReq.TargetAccountID

	// As nobody is actually Rejecting the reply
	// directly, but it's an implicit Reject coming
	// from our internal logic, don't bother setting
	// a URI (it's not a required field anyway).
	uri := ""

	rejection := &gtsmodel.InteractionRequest{
		ID:                   rejectID,
		StatusID:             inReplyToID,
		TargetAccountID:      targetAccountID,
		InteractingAccountID: reply.AccountID,
		InteractionURI:       reply.URI,
		InteractionType:      gtsmodel.InteractionReply,
		URI:                  uri,
		RejectedAt:           time.Now(),
	}
	err := d.state.DB.PutInteractionRequest(ctx, rejection)
	if err != nil && !errors.Is(err, db.ErrAlreadyExists) {
		return gtserror.Newf("db error putting pre-rejected interaction request: %w", err)
	}

	return nil
}

// isPermittedByAcceptIRI checks whether the given acceptIRI
// permits the given reply to the given inReplyTo status.
// If yes, then thisReq will be updated to reflect the
// acceptance, if it's not nil.
func (d *Dereferencer) isPermittedByAcceptIRI(
	ctx context.Context,
	requestUser string,
	reply *gtsmodel.Status,
	inReplyTo *gtsmodel.Status,
	thisReq *gtsmodel.InteractionRequest,
	acceptIRI string,
) (bool, error) {
	permitted, err := d.isValidAccept(
		ctx,
		requestUser,
		acceptIRI,
		reply.URI,
		inReplyTo.AccountURI,
	)
	if err != nil {
		// Error dereferencing means we couldn't
		// get the Accept right now or it wasn't
		// valid, so we shouldn't store this status.
		err := gtserror.Newf("undereferencable ApprovedByURI: %w", err)
		return false, err
	}

	if !permitted {
		// It's a no from
		// us, squirt.
		return false, nil
	}

	// Reply is permitted by this Accept.
	// If it was previously rejected or
	// pending approval, clear that now.
	reply.PendingApproval = util.Ptr(false)
	if thisReq != nil {
		thisReq.URI = acceptIRI
		thisReq.AcceptedAt = time.Now()
		thisReq.RejectedAt = time.Time{}
		err := d.state.DB.UpdateInteractionRequest(
			ctx,
			thisReq,
			"uri",
			"accepted_at",
			"rejected_at",
		)
		if err != nil {
			return false, gtserror.Newf("db error updating interaction request: %w", err)
		}
	}

	// All good!
	return true, nil
}

func (d *Dereferencer) rejectedByPolicy(
	ctx context.Context,
	reply *gtsmodel.Status,
	inReplyTo *gtsmodel.Status,
	thisReq *gtsmodel.InteractionRequest,
) error {
	var (
		rejectID  string
		rejectURI string
	)

	if thisReq != nil {
		// Reuse existing ID.
		rejectID = thisReq.ID
	} else {
		// Generate new ID.
		rejectID = id.NewULID()
	}

	if inReplyTo.IsLocal() {
		// If this a reply to one of our statuses
		// we should generate a URI for the Reject,
		// else just use an implicit (empty) URI.
		rejectURI = uris.GenerateURIForReject(
			inReplyTo.Account.Username,
			rejectID,
		)
	}

	if thisReq != nil {
		// Before we return, ensure interaction
		// request is marked as rejected.
		thisReq.RejectedAt = time.Now()
		thisReq.AcceptedAt = time.Time{}
		thisReq.URI = rejectURI
		err := d.state.DB.UpdateInteractionRequest(
			ctx,
			thisReq,
			"rejected_at",
			"accepted_at",
			"uri",
		)
		if err != nil {
			return gtserror.Newf("db error updating interaction request: %w", err)
		}

		return nil
	}

	// We haven't stored a rejected interaction
	// request for this status yet, do it now.
	rejection := &gtsmodel.InteractionRequest{
		ID:                   rejectID,
		StatusID:             inReplyTo.ID,
		TargetAccountID:      inReplyTo.AccountID,
		InteractingAccountID: reply.AccountID,
		InteractionURI:       reply.URI,
		InteractionType:      gtsmodel.InteractionReply,
		URI:                  rejectURI,
		RejectedAt:           time.Now(),
	}
	err := d.state.DB.PutInteractionRequest(ctx, rejection)
	if err != nil && !errors.Is(err, db.ErrAlreadyExists) {
		return gtserror.Newf("db error putting pre-rejected interaction request: %w", err)
	}

	return nil
}

func (d *Dereferencer) isPermittedBoost(
	ctx context.Context,
	requestUser string,
	status *gtsmodel.Status,
) (bool, error) {

	// Extract boost from status.
	boostOf := status.BoostOf
	if boostOf.BoostOfID != "" {

		// We do not permit boosts of
		// boost wrapper statuses. (this
		// shouldn't be able to happen).
		return false, nil
	}

	// Check visibility of local
	// boostOf to boosting account.
	if boostOf.IsLocal() {
		visible, err := d.visFilter.StatusVisible(ctx,
			status.Account,
			boostOf,
		)
		if err != nil {
			err := gtserror.Newf("error checking boostOf visibility: %w", err)
			return false, err
		}

		// Our status is not visible to the
		// account trying to do the boost.
		if !visible {
			return false, nil
		}
	}

	// Check interaction policy of boostOf.
	boostable, err := d.intFilter.StatusBoostable(ctx,
		status.Account,
		boostOf,
	)
	if err != nil {
		err := gtserror.Newf("error checking status boostability: %w", err)
		return false, err
	}

	if boostable.Forbidden() {
		// Booster is not permitted
		// to do this interaction.
		return false, nil
	}

	if boostable.Permitted() &&
		!boostable.MatchedOnCollection() {
		// Booster is permitted to do this
		// interaction, and didn't match on
		// a collection so we don't need to
		// do further checking.
		return true, nil
	}

	// Booster is permitted to do this
	// interaction pending approval, or
	// permitted but matched on a collection.
	//
	// Check if we can dereference
	// an Accept that grants approval.

	if status.ApprovedByURI == "" {
		// Status doesn't claim to be approved.
		//
		// For boosts of local statuses that's
		// fine, we can put it in the DB pending
		// approval, and continue processing it.
		//
		// If permission was granted based on a match
		// with a followers or following collection,
		// we can mark it as PreApproved so the processor
		// sends an accept out for it immediately.
		//
		// For boosts of remote statuses, though
		// we should be polite and just drop it.
		if boostOf.IsLocal() {
			status.PendingApproval = util.Ptr(true)
			status.PreApproved = boostable.MatchedOnCollection()
			return true, nil
		}

		return false, nil
	}

	// Boost claims to be approved, check
	// this by dereferencing the Accept and
	// inspecting the return value.
	permitted, err := d.isValidAccept(
		ctx,
		requestUser,
		status.ApprovedByURI,
		status.URI,
		boostOf.AccountURI,
	)
	if err != nil {
		// Error dereferencing means we couldn't
		// get the Accept right now or it wasn't
		// valid, so we shouldn't store this status.
		err := gtserror.Newf("undereferencable ApprovedByURI: %w", err)
		return false, err
	}

	if !permitted {
		return false, nil
	}

	// Status has been approved.
	status.PendingApproval = util.Ptr(false)
	return true, nil
}

// isValidAccept dereferences the activitystreams Accept at the
// specified IRI, and checks the Accept for validity against the
// provided expectedObject and expectedActor.
//
// Will return either (true, nil) if everything looked OK, an error
// if something went wrong internally during deref, or (false, nil)
// if the dereferenced Accept did not meet expectations.
func (d *Dereferencer) isValidAccept(
	ctx context.Context,
	requestUser string,
	acceptIRIStr string, // Eg., "https://example.org/users/someone/accepts/01J2736AWWJ3411CPR833F6D03"
	expectObjectURIStr string, // Eg., "https://some.instance.example.org/users/someone_else/statuses/01J27414TWV9F7DC39FN8ABB5R"
	expectActorURIStr string, // Eg., "https://example.org/users/someone"
) (bool, error) {
	l := log.
		WithContext(ctx).
		WithField("acceptIRI", acceptIRIStr)

	acceptIRI, err := url.Parse(acceptIRIStr)
	if err != nil {
		// Real returnable error.
		err := gtserror.Newf("error parsing acceptIRI: %w", err)
		return false, err
	}

	// Don't make calls to the Accept IRI
	// if it's blocked, just return false.
	blocked, err := d.state.DB.IsDomainBlocked(ctx, acceptIRI.Host)
	if err != nil {
		// Real returnable error.
		err := gtserror.Newf("error checking domain block: %w", err)
		return false, err
	}

	if blocked {
		l.Info("Accept host is blocked")
		return false, nil
	}

	tsport, err := d.transportController.NewTransportForUsername(ctx, requestUser)
	if err != nil {
		// Real returnable error.
		err := gtserror.Newf("error creating transport: %w", err)
		return false, err
	}

	// Make the call to resolve into an Acceptable.
	// Log any error encountered here but don't
	// return it as it's not *our* error.
	rsp, err := tsport.Dereference(ctx, acceptIRI)
	if err != nil {
		l.Errorf("error dereferencing Accept: %v", err)
		return false, nil
	}

	acceptable, err := ap.ResolveAcceptable(ctx, rsp.Body)

	// Tidy up rsp body.
	_ = rsp.Body.Close()

	if err != nil {
		l.Errorf("error resolving to Accept: %v", err)
		return false, err
	}

	// Extract the URI/ID of the Accept.
	acceptID := ap.GetJSONLDId(acceptable)
	acceptIDStr := acceptID.String()

	// Check whether input URI and final returned URI
	// have changed (i.e. we followed some redirects).
	rspURL := rsp.Request.URL
	rspURLStr := rspURL.String()
	if rspURLStr != acceptIRIStr {
		// If rspURLStr != acceptIRIStr, make sure final
		// response URL is at least on the same host as
		// what we expected (ie., we weren't redirected
		// across domains), and make sure it's the same
		// as the ID of the Accept we were returned.
		switch {
		case rspURL.Host != acceptIRI.Host:
			l.Errorf(
				"final deref host %s did not match acceptIRI host",
				rspURL.Host,
			)
			return false, nil

		case acceptIDStr != rspURLStr:
			l.Errorf(
				"final deref uri %s did not match returned Accept ID %s",
				rspURLStr, acceptIDStr,
			)
			return false, nil
		}
	}

	// Response is superficially OK,
	// check in more detail now.

	// Extract the actor IRI and string from Accept.
	actorIRIs := ap.GetActorIRIs(acceptable)
	actorIRI, actorIRIStr := extractIRI(actorIRIs)
	switch {
	case actorIRIStr == "":
		l.Error("Accept missing actor IRI")
		return false, nil

	// Ensure the Accept Actor is on
	// the instance hosting the Accept.
	case actorIRI.Host != acceptID.Host:
		l.Errorf(
			"actor %s not on the same host as Accept",
			actorIRIStr,
		)
		return false, nil

	// Ensure the Accept Actor is who we expect
	// it to be, and not someone else trying to
	// do an Accept for an interaction with a
	// statusable they don't own.
	case actorIRIStr != expectActorURIStr:
		l.Errorf(
			"actor %s was not the same as expected actor %s",
			actorIRIStr, expectActorURIStr,
		)
		return false, nil
	}

	// Extract the object IRI string from Accept.
	objectIRIs := ap.GetObjectIRIs(acceptable)
	_, objectIRIStr := extractIRI(objectIRIs)
	switch {
	case objectIRIStr == "":
		l.Error("missing Accept object IRI")
		return false, nil

	// Ensure the Accept Object is what we expect
	// it to be, ie., it's Accepting the interaction
	// we need it to Accept, and not something else.
	case objectIRIStr != expectObjectURIStr:
		l.Errorf(
			"resolved Accept object IRI %s was not the same as expected object %s",
			objectIRIStr, expectObjectURIStr,
		)
		return false, nil
	}

	// Everything looks OK.
	return true, nil
}

// extractIRI is shorthand to extract the first IRI
// url.URL{} object and serialized form from slice.
func extractIRI(iris []*url.URL) (*url.URL, string) {
	if len(iris) == 0 {
		return nil, ""
	}
	u := iris[0]
	return u, u.String()
}
