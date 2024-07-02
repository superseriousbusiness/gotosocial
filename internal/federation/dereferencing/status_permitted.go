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

	if replyable == gtsmodel.PolicyResultPermitted {
		// Replier is permitted
		// to do this interaction.
		return true, nil
	}

	if replyable == gtsmodel.PolicyResultForbidden {
		// Replier is not permitted
		// to do this interaction.
		return onFalse()
	}

	// Replier is permitted to do this
	// interaction pending approval.
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
		// For replies to remote statuses, though
		// we should be polite and just drop it.
		if inReplyTo.IsLocal() {
			status.PendingApproval = util.Ptr(true)
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

	if boostable == gtsmodel.PolicyResultPermitted {
		// Booster is permitted
		// to do this interaction.
		return true, nil
	}

	if boostable == gtsmodel.PolicyResultForbidden {
		// Booster is not permitted
		// to do this interaction.
		return onFalse()
	}

	// Booster is permitted to do this
	// interaction pending approval.
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
		// For boosts of remote statuses, though
		// we should be polite and just drop it.
		if boostOf.IsLocal() {
			status.PendingApproval = util.Ptr(true)
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
	approvedByURIStr string,
	expectedObject string,
	expectedActor string,
) error {
	approvedByURI, err := url.Parse(approvedByURIStr)
	if err != nil {
		err := gtserror.Newf("error parsing approvedByURI: %w", err)
		return err
	}

	if blocked, err := d.state.DB.IsDomainBlocked(ctx, approvedByURI.Host); blocked || err != nil {
		err := gtserror.Newf("domain %s is blocked", approvedByURI.Host)
		return err
	}

	transport, err := d.transportController.NewTransportForUsername(ctx, requestUser)
	if err != nil {
		err := gtserror.Newf("error creating transport: %w", err)
		return err
	}

	rsp, err := transport.Dereference(ctx, approvedByURI)
	if err != nil {
		err := gtserror.Newf("error dereferencing %s: %w", approvedByURIStr, err)
		return err
	}

	accept, err := ap.ResolveAccept(ctx, rsp.Body)

	// Tidy up rsp body.
	_ = rsp.Body.Close()

	if err != nil {
		err := gtserror.Newf("error resolving Accept %s: %w", approvedByURIStr, err)
		return err
	}

	// Check whether input URI and final returned URI
	// have changed (i.e. we followed some redirects).
	var (
		finalURI    *url.URL
		finalURIStr string
	)
	if rspURLStr := rsp.Request.URL.String(); rspURLStr != approvedByURIStr {
		// It changed!
		finalURI = rsp.Request.URL
		finalURIStr = rspURLStr
	} else {
		// Didn't change.
		finalURI = approvedByURI
		finalURIStr = approvedByURIStr
	}

	// Ensure the final URI has the expected ID.
	acceptURI := ap.GetJSONLDId(accept)
	acceptURIStr := acceptURI.String()
	if acceptURIStr != finalURIStr {
		err := gtserror.Newf(
			"resolved Accept ID/URI %s did not match final Accept uri %s",
			acceptURIStr, finalURIStr,
		)
		return err
	}

	// Ensure the final URI has the same host as the
	// Actor of the Accept, so we know we're getting
	// the Accept straight from the horse's mouth.
	actorIRIs := ap.GetActorIRIs(accept)
	if len(actorIRIs) != 1 {
		err := gtserror.New("resolved Accept actor(s) length was not 1")
		return gtserror.SetMalformed(err)
	}
	actorIRI := actorIRIs[0]
	actorStr := actorIRI.String()

	if actorIRI.Host != finalURI.Host {
		err := gtserror.Newf(
			"resolved Accept Actor uri %s was not the same host as final Accept uri %s",
			actorStr, finalURIStr,
		)
		return err
	}

	// Ensure the Accept Actor
	// is who we expect it to be.
	if actorStr != expectedActor {
		err := gtserror.Newf(
			"resolved Accept Actor uri %s was not the same as expected actor %s",
			actorStr, expectedActor,
		)
		return err
	}

	// Ensure the Accept Object is
	// what we expect it to be.
	objectIRIs := ap.GetObjectIRIs(accept)
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
