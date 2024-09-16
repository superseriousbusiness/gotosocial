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

	if !permitted && existing != nil {
		log.Infof(ctx, "deleting unpermitted: %s", existing.URI)

		// Delete existing status from database as it's no longer permitted.
		if err := d.state.DB.DeleteStatusByID(ctx, existing.ID); err != nil {
			log.Errorf(ctx, "error deleting %s after permissivity fail: %v", existing.URI, err)
		}
	}

	return
}

func (d *Dereferencer) isPermittedReply(
	ctx context.Context,
	requestUser string,
	status *gtsmodel.Status,
) (bool, error) {
	var (
		statusURI    = status.URI          // Definitely set.
		inReplyToURI = status.InReplyToURI // Definitely set.
		inReplyTo    = status.InReplyTo    // Might not yet be set.
	)

	// Check if status with this URI has previously been rejected.
	req, err := d.state.DB.GetInteractionRequestByInteractionURI(
		gtscontext.SetBarebones(ctx),
		statusURI,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting interaction request: %w", err)
		return false, err
	}

	if req != nil && req.IsRejected() {
		// This status has been
		// rejected reviously, so
		// it's not permitted now.
		return false, nil
	}

	// Check if replied-to status has previously been rejected.
	req, err = d.state.DB.GetInteractionRequestByInteractionURI(
		gtscontext.SetBarebones(ctx),
		inReplyToURI,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting interaction request: %w", err)
		return false, err
	}

	if req != nil && req.IsRejected() {
		// This status's parent was rejected, so
		// implicitly this reply should be rejected too.
		//
		// We know already that we haven't inserted
		// a rejected interaction request for this
		// status yet so do it before returning.
		id := id.NewULID()

		// To ensure the Reject chain stays coherent,
		// borrow fields from the up-thread rejection.
		// This collapses the chain beyond the first
		// rejected reply and allows us to avoid derefing
		// further replies we already know we don't want.
		statusID := req.StatusID
		targetAccountID := req.TargetAccountID

		// As nobody is actually Rejecting the reply
		// directly, but it's an implicit Reject coming
		// from our internal logic, don't bother setting
		// a URI (it's not a required field anyway).
		uri := ""

		rejection := &gtsmodel.InteractionRequest{
			ID:                   id,
			StatusID:             statusID,
			TargetAccountID:      targetAccountID,
			InteractingAccountID: status.AccountID,
			InteractionURI:       statusURI,
			InteractionType:      gtsmodel.InteractionReply,
			URI:                  uri,
			RejectedAt:           time.Now(),
		}
		err := d.state.DB.PutInteractionRequest(ctx, rejection)
		if err != nil && !errors.Is(err, db.ErrAlreadyExists) {
			return false, gtserror.Newf("db error putting pre-rejected interaction request: %w", err)
		}

		return false, nil
	}

	if inReplyTo == nil {
		// We didn't have the replied-to status in
		// our database (yet) so we can't know if
		// this reply is permitted or not. For now
		// just return true; worst-case, the status
		// sticks around on the instance for a couple
		// hours until we try to dereference it again
		// and realize it should be forbidden.
		return true, nil
	}

	if inReplyTo.BoostOfID != "" {
		// We do not permit replies to
		// boost wrapper statuses. (this
		// shouldn't be able to happen).
		log.Info(ctx, "rejecting reply to boost wrapper status")
		return false, nil
	}

	// Check visibility of local
	// inReplyTo to replying account.
	if inReplyTo.IsLocal() {
		visible, err := d.visFilter.StatusVisible(ctx,
			status.Account,
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

	// Check interaction policy of inReplyTo.
	replyable, err := d.intFilter.StatusReplyable(ctx,
		status.Account,
		inReplyTo,
	)
	if err != nil {
		err := gtserror.Newf("error checking status replyability: %w", err)
		return false, err
	}

	if replyable.Forbidden() {
		// Reply is not permitted.
		//
		// Insert a pre-rejected interaction request
		// into the db and return. This ensures that
		// replies to this now-rejected status aren't
		// inadvertently permitted.
		id := id.NewULID()
		rejection := &gtsmodel.InteractionRequest{
			ID:                   id,
			StatusID:             inReplyTo.ID,
			TargetAccountID:      inReplyTo.AccountID,
			InteractingAccountID: status.AccountID,
			InteractionURI:       statusURI,
			InteractionType:      gtsmodel.InteractionReply,
			URI:                  uris.GenerateURIForReject(inReplyTo.Account.Username, id),
			RejectedAt:           time.Now(),
		}
		err := d.state.DB.PutInteractionRequest(ctx, rejection)
		if err != nil && !errors.Is(err, db.ErrAlreadyExists) {
			return false, gtserror.Newf("db error putting pre-rejected interaction request: %w", err)
		}

		return false, nil
	}

	if replyable.Permitted() &&
		!replyable.MatchedOnCollection() {
		// Replier is permitted to do this
		// interaction, and didn't match on
		// a collection so we don't need to
		// do further checking.
		return true, nil
	}

	// Replier is permitted to do this
	// interaction pending approval, or
	// permitted but matched on a collection.
	//
	// Check if we can dereference
	// an Accept that grants approval.

	if status.ApprovedByURI == "" {
		// Status doesn't claim to be approved.
		//
		// For replies to local statuses that's
		// fine, we can put it in the DB pending
		// approval, and continue processing it.
		//
		// If permission was granted based on a match
		// with a followers or following collection,
		// we can mark it as PreApproved so the processor
		// sends an accept out for it immediately.
		//
		// For replies to remote statuses, though
		// we should be polite and just drop it.
		if inReplyTo.IsLocal() {
			status.PendingApproval = util.Ptr(true)
			status.PreApproved = replyable.MatchedOnCollection()
			return true, nil
		}

		return false, nil
	}

	// Status claims to be approved, check
	// this by dereferencing the Accept and
	// inspecting the return value.
	if err := d.validateApprovedBy(
		ctx,
		requestUser,
		status.ApprovedByURI,
		statusURI,
		inReplyTo.AccountURI,
	); err != nil {

		// Error dereferencing means we couldn't
		// get the Accept right now or it wasn't
		// valid, so we shouldn't store this status.
		log.Errorf(ctx, "undereferencable ApprovedByURI: %v", err)
		return false, nil
	}

	// Status has been approved.
	status.PendingApproval = util.Ptr(false)
	return true, nil
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
	if err := d.validateApprovedBy(
		ctx,
		requestUser,
		status.ApprovedByURI,
		status.URI,
		boostOf.AccountURI,
	); err != nil {

		// Error dereferencing means we couldn't
		// get the Accept right now or it wasn't
		// valid, so we shouldn't store this status.
		log.Errorf(ctx, "undereferencable ApprovedByURI: %v", err)
		return false, nil
	}

	// Status has been approved.
	status.PendingApproval = util.Ptr(false)
	return true, nil
}

// validateApprovedBy dereferences the activitystreams Accept at
// the specified IRI, and checks the Accept for validity against
// the provided expectedObject and expectedActor.
//
// Will return either nil if everything looked OK, or an error if
// something went wrong during deref, or if the dereffed Accept
// did not meet expectations.
func (d *Dereferencer) validateApprovedBy(
	ctx context.Context,
	requestUser string,
	approvedByURIStr string, // Eg., "https://example.org/users/someone/accepts/01J2736AWWJ3411CPR833F6D03"
	expectObjectURIStr string, // Eg., "https://some.instance.example.org/users/someone_else/statuses/01J27414TWV9F7DC39FN8ABB5R"
	expectActorURIStr string, // Eg., "https://example.org/users/someone"
) error {
	approvedByURI, err := url.Parse(approvedByURIStr)
	if err != nil {
		err := gtserror.Newf("error parsing approvedByURI: %w", err)
		return err
	}

	// Don't make calls to the remote if it's blocked.
	if blocked, err := d.state.DB.IsDomainBlocked(ctx, approvedByURI.Host); blocked || err != nil {
		err := gtserror.Newf("domain %s is blocked", approvedByURI.Host)
		return err
	}

	tsport, err := d.transportController.NewTransportForUsername(ctx, requestUser)
	if err != nil {
		err := gtserror.Newf("error creating transport: %w", err)
		return err
	}

	// Make the call to resolve into an Acceptable.
	rsp, err := tsport.Dereference(ctx, approvedByURI)
	if err != nil {
		err := gtserror.Newf("error dereferencing %s: %w", approvedByURIStr, err)
		return err
	}

	acceptable, err := ap.ResolveAcceptable(ctx, rsp.Body)

	// Tidy up rsp body.
	_ = rsp.Body.Close()

	if err != nil {
		err := gtserror.Newf("error resolving Accept %s: %w", approvedByURIStr, err)
		return err
	}

	// Extract the URI/ID of the Accept.
	acceptURI := ap.GetJSONLDId(acceptable)
	acceptURIStr := acceptURI.String()

	// Check whether input URI and final returned URI
	// have changed (i.e. we followed some redirects).
	rspURL := rsp.Request.URL
	rspURLStr := rspURL.String()
	switch {
	case rspURLStr == approvedByURIStr:

	// i.e. from here, rspURLStr != approvedByURIStr.
	//
	// Make sure it's at least on the same host as
	// what we expected (ie., we weren't redirected
	// across domains), and make sure it's the same
	// as the ID of the Accept we were returned.
	case rspURL.Host != approvedByURI.Host:
		return gtserror.Newf(
			"final dereference host %s did not match approvedByURI host %s",
			rspURL.Host, approvedByURI.Host,
		)
	case acceptURIStr != rspURLStr:
		return gtserror.Newf(
			"final dereference uri %s did not match returned Accept ID/URI %s",
			rspURLStr, acceptURIStr,
		)
	}

	// Extract the actor IRI and string from Accept.
	actorIRIs := ap.GetActorIRIs(acceptable)
	actorIRI, actorIRIStr := extractIRI(actorIRIs)
	switch {
	case actorIRIStr == "":
		err := gtserror.New("missing Accept actor IRI")
		return gtserror.SetMalformed(err)

	// Ensure the Accept Actor is who we expect
	// it to be, and not someone else trying to
	// do an Accept for an interaction with a
	// statusable they don't own.
	case actorIRI.Host != acceptURI.Host:
		return gtserror.Newf(
			"Accept Actor %s was not the same host as Accept %s",
			actorIRIStr, acceptURIStr,
		)

	// Ensure the Accept Actor is who we expect
	// it to be, and not someone else trying to
	// do an Accept for an interaction with a
	// statusable they don't own.
	case actorIRIStr != expectActorURIStr:
		return gtserror.Newf(
			"Accept Actor %s was not the same as expected actor %s",
			actorIRIStr, expectActorURIStr,
		)
	}

	// Extract the object IRI string from Accept.
	objectIRIs := ap.GetObjectIRIs(acceptable)
	_, objectIRIStr := extractIRI(objectIRIs)
	switch {
	case objectIRIStr == "":
		err := gtserror.New("missing Accept object IRI")
		return gtserror.SetMalformed(err)

	// Ensure the Accept Object is what we expect
	// it to be, ie., it's Accepting the interaction
	// we need it to Accept, and not something else.
	case objectIRIStr != expectObjectURIStr:
		return gtserror.Newf(
			"resolved Accept Object uri %s was not the same as expected object %s",
			objectIRIStr, expectObjectURIStr,
		)
	}

	return nil
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
