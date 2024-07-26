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
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
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
func (d *Dereferencer) isPermittedStatus(
	ctx context.Context,
	requestUser string,
	existing *gtsmodel.Status,
	status *gtsmodel.Status,
) (
	bool, // is permitted?
	error,
) {
	// our failure condition handling
	// at the end of this function for
	// the case of permission = false.
	onFalse := func() (bool, error) {
		if existing != nil {
			log.Infof(ctx, "deleting unpermitted: %s", existing.URI)

			// Delete existing status from database as it's no longer permitted.
			if err := d.state.DB.DeleteStatusByID(ctx, existing.ID); err != nil {
				log.Errorf(ctx, "error deleting %s after permissivity fail: %v", existing.URI, err)
			}
		}
		return false, nil
	}

	if status.Account.IsSuspended() {
		// The status author is suspended,
		// this shouldn't have reached here
		// but it's a fast check anyways.
		log.Debugf(ctx,
			"status author %s is suspended",
			status.AccountURI,
		)
		return onFalse()
	}

	if inReplyTo := status.InReplyTo; inReplyTo != nil {
		return d.isPermittedReply(
			ctx,
			requestUser,
			status,
			inReplyTo,
			onFalse,
		)
	} else if boostOf := status.BoostOf; boostOf != nil {
		return d.isPermittedBoost(
			ctx,
			requestUser,
			status,
			boostOf,
			onFalse,
		)
	}

	// Nothing else stopping this.
	return true, nil
}

func (d *Dereferencer) isPermittedReply(
	ctx context.Context,
	requestUser string,
	status *gtsmodel.Status,
	inReplyTo *gtsmodel.Status,
	onFalse func() (bool, error),
) (bool, error) {
	if inReplyTo.BoostOfID != "" {
		// We do not permit replies to
		// boost wrapper statuses. (this
		// shouldn't be able to happen).
		log.Info(ctx, "rejecting reply to boost wrapper status")
		return onFalse()
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
			return onFalse()
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
		// Replier is not permitted
		// to do this interaction.
		return onFalse()
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

		return onFalse()
	}

	// Status claims to be approved, check
	// this by dereferencing the Accept and
	// inspecting the return value.
	if err := d.validateApprovedBy(
		ctx,
		requestUser,
		status.ApprovedByURI,
		status.URI,
		inReplyTo.AccountURI,
	); err != nil {
		// Error dereferencing means we couldn't
		// get the Accept right now or it wasn't
		// valid, so we shouldn't store this status.
		//
		// Do log the error though as it may be
		// interesting for admins to see.
		log.Info(ctx, "rejecting reply with undereferenceable ApprovedByURI: %v", err)
		return onFalse()
	}

	// Status has been approved.
	status.PendingApproval = util.Ptr(false)
	return true, nil
}

func (d *Dereferencer) isPermittedBoost(
	ctx context.Context,
	requestUser string,
	status *gtsmodel.Status,
	boostOf *gtsmodel.Status,
	onFalse func() (bool, error),
) (bool, error) {
	if boostOf.BoostOfID != "" {
		// We do not permit boosts of
		// boost wrapper statuses. (this
		// shouldn't be able to happen).
		log.Info(ctx, "rejecting boost of boost wrapper status")
		return onFalse()
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
			return onFalse()
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
		return onFalse()
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

		return onFalse()
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
		//
		// Do log the error though as it may be
		// interesting for admins to see.
		log.Info(ctx, "rejecting boost with undereferenceable ApprovedByURI: %v", err)
		return onFalse()
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
	expectedObject string, // Eg., "https://some.instance.example.org/users/someone_else/statuses/01J27414TWV9F7DC39FN8ABB5R"
	expectedActor string, // Eg., "https://example.org/users/someone"
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

	transport, err := d.transportController.NewTransportForUsername(ctx, requestUser)
	if err != nil {
		err := gtserror.Newf("error creating transport: %w", err)
		return err
	}

	// Make the call to resolve into an Acceptable.
	rsp, err := transport.Dereference(ctx, approvedByURI)
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
	if rspURLStr != approvedByURIStr {
		// Final URI was different from approvedByURIStr.
		//
		// Make sure it's at least on the same host as
		// what we expected (ie., we weren't redirected
		// across domains), and make sure it's the same
		// as the ID of the Accept we were returned.
		if rspURL.Host != approvedByURI.Host {
			err := gtserror.Newf(
				"final dereference host %s did not match approvedByURI host %s",
				rspURL.Host, approvedByURI.Host,
			)
			return err
		}

		if acceptURIStr != rspURLStr {
			err := gtserror.Newf(
				"final dereference uri %s did not match returned Accept ID/URI %s",
				rspURLStr, acceptURIStr,
			)
			return err
		}
	}

	// Ensure the Accept URI has the same host
	// as the Accept Actor, so we know we're
	// not dealing with someone on a different
	// domain just pretending to be the Actor.
	actorIRIs := ap.GetActorIRIs(acceptable)
	if len(actorIRIs) != 1 {
		err := gtserror.New("resolved Accept actor(s) length was not 1")
		return gtserror.SetMalformed(err)
	}

	actorIRI := actorIRIs[0]
	actorStr := actorIRI.String()

	if actorIRI.Host != acceptURI.Host {
		err := gtserror.Newf(
			"Accept Actor %s was not the same host as Accept %s",
			actorStr, acceptURIStr,
		)
		return err
	}

	// Ensure the Accept Actor is who we expect
	// it to be, and not someone else trying to
	// do an Accept for an interaction with a
	// statusable they don't own.
	if actorStr != expectedActor {
		err := gtserror.Newf(
			"Accept Actor %s was not the same as expected actor %s",
			actorStr, expectedActor,
		)
		return err
	}

	// Ensure the Accept Object is what we expect
	// it to be, ie., it's Accepting the interaction
	// we need it to Accept, and not something else.
	objectIRIs := ap.GetObjectIRIs(acceptable)
	if len(objectIRIs) != 1 {
		err := gtserror.New("resolved Accept object(s) length was not 1")
		return err
	}

	objectIRI := objectIRIs[0]
	objectStr := objectIRI.String()

	if objectStr != expectedObject {
		err := gtserror.Newf(
			"resolved Accept Object uri %s was not the same as expected object %s",
			objectStr, expectedObject,
		)
		return err
	}

	return nil
}
