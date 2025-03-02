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

package workers

import (
	"context"
	"net/url"

	"codeberg.org/superseriousbusiness/activity/streams"
	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// federate wraps functions for federating
// something out via ActivityPub in response
// to message processing.
type federate struct {
	// Embed federator to give access
	// to send and retrieve functions.
	*federation.Federator
	state     *state.State
	converter *typeutils.Converter
}

// parseURI is a cheeky little
// shortcut to wrap parsing errors.
//
// The returned err will be prepended
// with the name of the function that
// called this function, so it can be
// returned without further wrapping.
func parseURI(s string) (*url.URL, error) {
	const (
		// Provides enough calldepth to
		// prepend the name of whatever
		// function called *this* one,
		// so that they don't have to
		// wrap the error themselves.
		calldepth = 3
		errFmt    = "error parsing uri %s: %w"
	)

	uri, err := url.Parse(s)
	if err != nil {
		return nil, gtserror.NewfAt(calldepth, errFmt, s, err)
	}

	return uri, err
}

func (f *federate) DeleteAccount(ctx context.Context, account *gtsmodel.Account) error {
	// Do nothing if it's not our
	// account that's been deleted.
	if !account.IsLocal() {
		return nil
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(account.OutboxURI)
	if err != nil {
		return err
	}

	actorIRI, err := parseURI(account.URI)
	if err != nil {
		return err
	}

	followersIRI, err := parseURI(account.FollowersURI)
	if err != nil {
		return err
	}

	// Create a new delete.
	// todo: tc.AccountToASDelete
	delete := streams.NewActivityStreamsDelete()

	// Set the Actor for the delete; no matter
	// who actually did the delete, we should
	// use the account owner for this.
	deleteActor := streams.NewActivityStreamsActorProperty()
	deleteActor.AppendIRI(actorIRI)
	delete.SetActivityStreamsActor(deleteActor)

	// Set the account's IRI as the 'object' property.
	deleteObject := streams.NewActivityStreamsObjectProperty()
	deleteObject.AppendIRI(actorIRI)
	delete.SetActivityStreamsObject(deleteObject)

	// Address the delete To followers.
	deleteTo := streams.NewActivityStreamsToProperty()
	deleteTo.AppendIRI(followersIRI)
	delete.SetActivityStreamsTo(deleteTo)

	// Address the delete CC public.
	deleteCC := streams.NewActivityStreamsCcProperty()
	deleteCC.AppendIRI(ap.PublicURI())
	delete.SetActivityStreamsCc(deleteCC)

	// Send the Delete via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, delete,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			delete, outboxIRI, err,
		)
	}

	return nil
}

// CreateStatus sends the given status out to relevant
// recipients with the Outbox of the status creator.
//
// If the status is pending approval, then it will be
// sent **ONLY** to the inbox of the account it replies to,
// ignoring shared inboxes.
func (f *federate) CreateStatus(ctx context.Context, status *gtsmodel.Status) error {
	// Do nothing if the status
	// shouldn't be federated.
	if status.IsLocalOnly() {
		return nil
	}

	// Do nothing if this
	// isn't our status.
	if !*status.Local {
		return nil
	}

	// Ensure the status model is fully populated.
	if err := f.state.DB.PopulateStatus(ctx, status); err != nil {
		return gtserror.Newf("error populating status: %w", err)
	}

	// Convert status to AS Statusable implementing type.
	statusable, err := f.converter.StatusToAS(ctx, status)
	if err != nil {
		return gtserror.Newf("error converting status to Statusable: %w", err)
	}

	// If status is pending approval,
	// it must be a reply. Deliver it
	// **ONLY** to the account it replies
	// to, on behalf of the replier.
	if util.PtrOrValue(status.PendingApproval, false) {
		return f.deliverToInboxOnly(
			ctx,
			status.Account,
			status.InReplyToAccount,
			// Status has to be wrapped in Create activity.
			typeutils.WrapStatusableInCreate(statusable, false),
		)
	}

	// Parse the outbox URI of the status author.
	outboxIRI, err := parseURI(status.Account.OutboxURI)
	if err != nil {
		return err
	}

	// Send a Create activity with Statusable via the Actor's outbox.
	create := typeutils.WrapStatusableInCreate(statusable, false)
	if _, err := f.FederatingActor().Send(ctx, outboxIRI, create); err != nil {
		return gtserror.Newf("error sending Create activity via outbox %s: %w", outboxIRI, err)
	}
	return nil
}

func (f *federate) CreatePollVote(ctx context.Context, poll *gtsmodel.Poll, vote *gtsmodel.PollVote) error {
	// Extract status from poll.
	status := poll.Status

	// Do nothing if the status
	// shouldn't be federated.
	if status.IsLocalOnly() {
		return nil
	}

	// Do nothing if this is
	// a vote in our status.
	if *status.Local {
		return nil
	}

	// Parse the outbox URI of the poll vote author.
	outboxIRI, err := parseURI(vote.Account.OutboxURI)
	if err != nil {
		return err
	}

	// Convert vote to AS Creates with vote choices as Objects.
	creates, err := f.converter.PollVoteToASCreates(ctx, vote)
	if err != nil {
		return gtserror.Newf("error converting to notes: %w", err)
	}

	var errs gtserror.MultiError

	// Send each create activity.
	actor := f.FederatingActor()
	for _, create := range creates {
		if _, err := actor.Send(ctx, outboxIRI, create); err != nil {
			errs.Appendf("error sending Create activity via outbox %s: %w", outboxIRI, err)
		}
	}

	return errs.Combine()
}

func (f *federate) DeleteStatus(ctx context.Context, status *gtsmodel.Status) error {
	// Do nothing if the status
	// shouldn't be federated.
	if status.IsLocalOnly() {
		return nil
	}

	// Do nothing if this
	// isn't our status.
	if !*status.Local {
		return nil
	}

	// Parse the outbox URI of the status author.
	outboxIRI, err := parseURI(status.Account.OutboxURI)
	if err != nil {
		return err
	}

	// Wrap the status URI in a Delete activity.
	delete, err := f.converter.StatusToASDelete(ctx, status)
	if err != nil {
		return gtserror.Newf("error creating Delete: %w", err)
	}

	// Send the Delete via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, delete,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			delete, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) UpdateStatus(ctx context.Context, status *gtsmodel.Status) error {
	// Do nothing if the status
	// shouldn't be federated.
	if status.IsLocalOnly() {
		return nil
	}

	// Do nothing if this
	// isn't our status.
	if !*status.Local {
		return nil
	}

	// Ensure the status model is fully populated.
	if err := f.state.DB.PopulateStatus(ctx, status); err != nil {
		return gtserror.Newf("error populating status: %w", err)
	}

	// Parse the outbox URI of the status author.
	outboxIRI, err := parseURI(status.Account.OutboxURI)
	if err != nil {
		return err
	}

	// Convert status to ActivityStreams Statusable implementing type.
	statusable, err := f.converter.StatusToAS(ctx, status)
	if err != nil {
		return gtserror.Newf("error converting status to Statusable: %w", err)
	}

	// Send an Update activity with Statusable via the Actor's outbox.
	update := typeutils.WrapStatusableInUpdate(statusable, false)
	if _, err := f.FederatingActor().Send(ctx, outboxIRI, update); err != nil {
		return gtserror.Newf("error sending Update activity via outbox %s: %w", outboxIRI, err)
	}

	return nil
}

func (f *federate) Follow(ctx context.Context, follow *gtsmodel.Follow) error {
	// Populate model.
	if err := f.state.DB.PopulateFollow(ctx, follow); err != nil {
		return gtserror.Newf("error populating follow: %w", err)
	}

	// Do nothing if both accounts are local.
	if follow.Account.IsLocal() &&
		follow.TargetAccount.IsLocal() {
		return nil
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(follow.Account.OutboxURI)
	if err != nil {
		return err
	}

	// Convert follow to ActivityStreams Follow.
	asFollow, err := f.converter.FollowToAS(ctx, follow)
	if err != nil {
		return gtserror.Newf("error converting follow to AS: %s", err)
	}

	// Send the Follow via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, asFollow,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			asFollow, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) UndoFollow(ctx context.Context, follow *gtsmodel.Follow) error {
	// Populate model.
	if err := f.state.DB.PopulateFollow(ctx, follow); err != nil {
		return gtserror.Newf("error populating follow: %w", err)
	}

	// Do nothing if both accounts are local.
	if follow.Account.IsLocal() &&
		follow.TargetAccount.IsLocal() {
		return nil
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(follow.Account.OutboxURI)
	if err != nil {
		return err
	}

	targetAccountIRI, err := parseURI(follow.TargetAccount.URI)
	if err != nil {
		return err
	}

	// Recreate the ActivityStreams Follow.
	asFollow, err := f.converter.FollowToAS(ctx, follow)
	if err != nil {
		return gtserror.Newf("error converting follow to AS: %w", err)
	}

	// Create a new Undo.
	// todo: tc.FollowToASUndo
	undo := streams.NewActivityStreamsUndo()

	// Set the Actor for the Undo:
	// same as the actor for the Follow.
	undo.SetActivityStreamsActor(asFollow.GetActivityStreamsActor())

	// Set recreated Follow as the 'object' property.
	//
	// For most AP implementations, it's not enough
	// to just send the URI of the original Follow,
	// we have to send the whole object again.
	undoObject := streams.NewActivityStreamsObjectProperty()
	undoObject.AppendActivityStreamsFollow(asFollow)
	undo.SetActivityStreamsObject(undoObject)

	// Address the Undo To the target account.
	undoTo := streams.NewActivityStreamsToProperty()
	undoTo.AppendIRI(targetAccountIRI)
	undo.SetActivityStreamsTo(undoTo)

	// Send the Undo via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, undo,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			undo, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) UndoLike(ctx context.Context, fave *gtsmodel.StatusFave) error {
	// Populate model.
	if err := f.state.DB.PopulateStatusFave(ctx, fave); err != nil {
		return gtserror.Newf("error populating fave: %w", err)
	}

	// Do nothing if both accounts are local.
	if fave.Account.IsLocal() &&
		fave.TargetAccount.IsLocal() {
		return nil
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(fave.Account.OutboxURI)
	if err != nil {
		return err
	}

	targetAccountIRI, err := parseURI(fave.TargetAccount.URI)
	if err != nil {
		return err
	}

	// Recreate the ActivityStreams Like.
	like, err := f.converter.FaveToAS(ctx, fave)
	if err != nil {
		return gtserror.Newf("error converting fave to AS: %w", err)
	}

	// Create a new Undo.
	// todo: tc.FaveToASUndo
	undo := streams.NewActivityStreamsUndo()

	// Set the Actor for the Undo:
	// same as the actor for the Like.
	undo.SetActivityStreamsActor(like.GetActivityStreamsActor())

	// Set recreated Like as the 'object' property.
	//
	// For most AP implementations, it's not enough
	// to just send the URI of the original Like,
	// we have to send the whole object again.
	undoObject := streams.NewActivityStreamsObjectProperty()
	undoObject.AppendActivityStreamsLike(like)
	undo.SetActivityStreamsObject(undoObject)

	// Address the Undo To the target account.
	undoTo := streams.NewActivityStreamsToProperty()
	undoTo.AppendIRI(targetAccountIRI)
	undo.SetActivityStreamsTo(undoTo)

	// Send the Undo via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, undo,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			undo, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) UndoAnnounce(ctx context.Context, boost *gtsmodel.Status) error {
	// Populate model.
	if err := f.state.DB.PopulateStatus(ctx, boost); err != nil {
		return gtserror.Newf("error populating status: %w", err)
	}

	// Do nothing if boosting
	// account isn't ours.
	if !boost.Account.IsLocal() {
		return nil
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(boost.Account.OutboxURI)
	if err != nil {
		return err
	}

	// Recreate the ActivityStreams Announce.
	asAnnounce, err := f.converter.BoostToAS(
		ctx,
		boost,
		boost.Account,
		boost.BoostOfAccount,
	)
	if err != nil {
		return gtserror.Newf("error converting boost to AS: %w", err)
	}

	// Create a new Undo.
	// todo: tc.AnnounceToASUndo
	undo := streams.NewActivityStreamsUndo()

	// Set the Actor for the Undo:
	// same as the actor for the Announce.
	undo.SetActivityStreamsActor(asAnnounce.GetActivityStreamsActor())

	// Set recreated Announce as the 'object' property.
	//
	// For most AP implementations, it's not enough
	// to just send the URI of the original Announce,
	// we have to send the whole object again.
	undoObject := streams.NewActivityStreamsObjectProperty()
	undoObject.AppendActivityStreamsAnnounce(asAnnounce)
	undo.SetActivityStreamsObject(undoObject)

	// Address the Undo To the Announce To.
	undo.SetActivityStreamsTo(asAnnounce.GetActivityStreamsTo())

	// Address the Undo CC the Announce CC.
	undo.SetActivityStreamsCc(asAnnounce.GetActivityStreamsCc())

	// Send the Undo via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, undo,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			undo, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) AcceptFollow(ctx context.Context, follow *gtsmodel.Follow) error {
	// Populate model.
	if err := f.state.DB.PopulateFollow(ctx, follow); err != nil {
		return gtserror.Newf("error populating follow: %w", err)
	}

	// Bail if requesting account is ours:
	// we've already accepted internally and
	// shouldn't send an Accept to ourselves.
	if follow.Account.IsLocal() {
		return nil
	}

	// Bail if target account isn't ours:
	// we can't Accept a follow on
	// another instance's behalf.
	if follow.TargetAccount.IsRemote() {
		return nil
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(follow.TargetAccount.OutboxURI)
	if err != nil {
		return err
	}

	acceptingAccountIRI, err := parseURI(follow.TargetAccount.URI)
	if err != nil {
		return err
	}

	requestingAccountIRI, err := parseURI(follow.Account.URI)
	if err != nil {
		return err
	}

	// Recreate the ActivityStreams Follow.
	asFollow, err := f.converter.FollowToAS(ctx, follow)
	if err != nil {
		return gtserror.Newf("error converting follow to AS: %w", err)
	}

	// Create a new Accept.
	// todo: tc.FollowToASAccept
	accept := streams.NewActivityStreamsAccept()

	// Set the requestee as Actor of the Accept.
	acceptActorProp := streams.NewActivityStreamsActorProperty()
	acceptActorProp.AppendIRI(acceptingAccountIRI)
	accept.SetActivityStreamsActor(acceptActorProp)

	// Set recreated Follow as the 'object' property.
	//
	// For most AP implementations, it's not enough
	// to just send the URI of the original Follow,
	// we have to send the whole object again.
	acceptObject := streams.NewActivityStreamsObjectProperty()
	acceptObject.AppendActivityStreamsFollow(asFollow)
	accept.SetActivityStreamsObject(acceptObject)

	// Address the Accept To the Follow requester.
	acceptTo := streams.NewActivityStreamsToProperty()
	acceptTo.AppendIRI(requestingAccountIRI)
	accept.SetActivityStreamsTo(acceptTo)

	// Send the Accept via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, accept,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			accept, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) RejectFollow(ctx context.Context, follow *gtsmodel.Follow) error {
	// Ensure follow populated before proceeding.
	if err := f.state.DB.PopulateFollow(ctx, follow); err != nil {
		return gtserror.Newf("error populating follow: %w", err)
	}

	// Bail if requesting account is ours:
	// we've already rejected internally and
	// shouldn't send an Reject to ourselves.
	if follow.Account.IsLocal() {
		return nil
	}

	// Bail if target account isn't ours:
	// we can't Reject a follow on
	// another instance's behalf.
	if follow.TargetAccount.IsRemote() {
		return nil
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(follow.TargetAccount.OutboxURI)
	if err != nil {
		return err
	}

	rejectingAccountIRI, err := parseURI(follow.TargetAccount.URI)
	if err != nil {
		return err
	}

	requestingAccountIRI, err := parseURI(follow.Account.URI)
	if err != nil {
		return err
	}

	// Recreate the ActivityStreams Follow.
	asFollow, err := f.converter.FollowToAS(ctx, follow)
	if err != nil {
		return gtserror.Newf("error converting follow to AS: %w", err)
	}

	// Create a new Reject.
	// todo: tc.FollowRequestToASReject
	reject := streams.NewActivityStreamsReject()

	// Set the requestee as Actor of the Reject.
	rejectActorProp := streams.NewActivityStreamsActorProperty()
	rejectActorProp.AppendIRI(rejectingAccountIRI)
	reject.SetActivityStreamsActor(rejectActorProp)

	// Set recreated Follow as the 'object' property.
	//
	// For most AP implementations, it's not enough
	// to just send the URI of the original Follow,
	// we have to send the whole object again.
	rejectObject := streams.NewActivityStreamsObjectProperty()
	rejectObject.AppendActivityStreamsFollow(asFollow)
	reject.SetActivityStreamsObject(rejectObject)

	// Address the Reject To the Follow requester.
	rejectTo := streams.NewActivityStreamsToProperty()
	rejectTo.AppendIRI(requestingAccountIRI)
	reject.SetActivityStreamsTo(rejectTo)

	// Send the Reject via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, reject,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			reject, outboxIRI, err,
		)
	}

	return nil
}

// Like sends the given fave out to relevant
// recipients with the Outbox of the status creator.
//
// If the fave is pending approval, then it will be
// sent **ONLY** to the inbox of the account it faves,
// ignoring shared inboxes.
func (f *federate) Like(ctx context.Context, fave *gtsmodel.StatusFave) error {
	// Populate model.
	if err := f.state.DB.PopulateStatusFave(ctx, fave); err != nil {
		return gtserror.Newf("error populating fave: %w", err)
	}

	// Do nothing if both accounts are local.
	if fave.Account.IsLocal() &&
		fave.TargetAccount.IsLocal() {
		return nil
	}

	// Create the ActivityStreams Like.
	like, err := f.converter.FaveToAS(ctx, fave)
	if err != nil {
		return gtserror.Newf("error converting fave to AS Like: %w", err)
	}

	// If fave is pending approval,
	// deliver it **ONLY** to the account
	// it faves, on behalf of the faver.
	if util.PtrOrValue(fave.PendingApproval, false) {
		return f.deliverToInboxOnly(
			ctx,
			fave.Account,
			fave.TargetAccount,
			like,
		)
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(fave.Account.OutboxURI)
	if err != nil {
		return err
	}

	// Send the Like via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, like,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			like, outboxIRI, err,
		)
	}

	return nil
}

// Announce sends the given boost out to relevant
// recipients with the Outbox of the status creator.
//
// If the boost is pending approval, then it will be
// sent **ONLY** to the inbox of the account it boosts,
// ignoring shared inboxes.
func (f *federate) Announce(ctx context.Context, boost *gtsmodel.Status) error {
	// Populate model.
	if err := f.state.DB.PopulateStatus(ctx, boost); err != nil {
		return gtserror.Newf("error populating status: %w", err)
	}

	// Do nothing if boosting
	// account isn't ours.
	if !boost.Account.IsLocal() {
		return nil
	}

	// Create the ActivityStreams Announce.
	announce, err := f.converter.BoostToAS(
		ctx,
		boost,
		boost.Account,
		boost.BoostOfAccount,
	)
	if err != nil {
		return gtserror.Newf("error converting boost to AS: %w", err)
	}

	// If announce is pending approval,
	// deliver it **ONLY** to the account
	// it boosts, on behalf of the booster.
	if util.PtrOrValue(boost.PendingApproval, false) {
		return f.deliverToInboxOnly(
			ctx,
			boost.Account,
			boost.BoostOfAccount,
			announce,
		)
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(boost.Account.OutboxURI)
	if err != nil {
		return err
	}

	// Send the Announce via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, announce,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			announce, outboxIRI, err,
		)
	}

	return nil
}

// deliverToInboxOnly delivers the given Activity
// *only* to the inbox of targetAcct, on behalf of
// sendingAcct, regardless of the `to` and `cc` values
// set on the activity. This should be used specifically
// for sending "pending approval" activities.
func (f *federate) deliverToInboxOnly(
	ctx context.Context,
	sendingAcct *gtsmodel.Account,
	targetAcct *gtsmodel.Account,
	t vocab.Type,
) error {
	if targetAcct.IsLocal() {
		// If this is a local target,
		// they've already received it.
		return nil
	}

	toInbox, err := url.Parse(targetAcct.InboxURI)
	if err != nil {
		return gtserror.Newf(
			"error parsing target inbox uri: %w",
			err,
		)
	}

	tsport, err := f.TransportController().NewTransportForUsername(
		ctx,
		sendingAcct.Username,
	)
	if err != nil {
		return gtserror.Newf(
			"error getting transport to deliver activity %T to target inbox %s: %w",
			t, targetAcct.InboxURI, err,
		)
	}

	m, err := ap.Serialize(t)
	if err != nil {
		return err
	}

	if err := tsport.Deliver(ctx, m, toInbox); err != nil {
		return gtserror.Newf(
			"error delivering activity %T to target inbox %s: %w",
			t, targetAcct.InboxURI, err,
		)
	}

	return nil
}

func (f *federate) UpdateAccount(ctx context.Context, account *gtsmodel.Account) error {
	// Populate model.
	if err := f.state.DB.PopulateAccount(ctx, account); err != nil {
		return gtserror.Newf("error populating account: %w", err)
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(account.OutboxURI)
	if err != nil {
		return err
	}

	// Convert account to Accountable.
	accountable, err := f.converter.AccountToAS(ctx, account)
	if err != nil {
		return gtserror.Newf("error converting account to Person: %w", err)
	}

	// Use Accountable as Object of Update.
	update, err := f.converter.WrapAccountableInUpdate(accountable)
	if err != nil {
		return gtserror.Newf("error wrapping Person in Update: %w", err)
	}

	// Send the Update via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, update,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			update, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) Block(ctx context.Context, block *gtsmodel.Block) error {
	// Populate model.
	if err := f.state.DB.PopulateBlock(ctx, block); err != nil {
		return gtserror.Newf("error populating block: %w", err)
	}

	// Do nothing if both accounts are local.
	if block.Account.IsLocal() &&
		block.TargetAccount.IsLocal() {
		return nil
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(block.Account.OutboxURI)
	if err != nil {
		return err
	}

	// Convert block to ActivityStreams Block.
	asBlock, err := f.converter.BlockToAS(ctx, block)
	if err != nil {
		return gtserror.Newf("error converting block to AS: %w", err)
	}

	// Send the Block via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, asBlock,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			asBlock, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) UndoBlock(ctx context.Context, block *gtsmodel.Block) error {
	// Populate model.
	if err := f.state.DB.PopulateBlock(ctx, block); err != nil {
		return gtserror.Newf("error populating block: %w", err)
	}

	// Do nothing if both accounts are local.
	if block.Account.IsLocal() &&
		block.TargetAccount.IsLocal() {
		return nil
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(block.Account.OutboxURI)
	if err != nil {
		return err
	}

	targetAccountIRI, err := parseURI(block.TargetAccount.URI)
	if err != nil {
		return err
	}

	// Convert block to ActivityStreams Block.
	asBlock, err := f.converter.BlockToAS(ctx, block)
	if err != nil {
		return gtserror.Newf("error converting block to AS: %w", err)
	}

	// Create a new Undo.
	// todo: tc.BlockToASUndo
	undo := streams.NewActivityStreamsUndo()

	// Set the Actor for the Undo:
	// same as the actor for the Block.
	undo.SetActivityStreamsActor(asBlock.GetActivityStreamsActor())

	// Set Block as the 'object' property.
	//
	// For most AP implementations, it's not enough
	// to just send the URI of the original Block,
	// we have to send the whole object again.
	undoObject := streams.NewActivityStreamsObjectProperty()
	undoObject.AppendActivityStreamsBlock(asBlock)
	undo.SetActivityStreamsObject(undoObject)

	// Address the Undo To the target account.
	undoTo := streams.NewActivityStreamsToProperty()
	undoTo.AppendIRI(targetAccountIRI)
	undo.SetActivityStreamsTo(undoTo)

	// Send the Undo via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, undo,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			undo, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) Flag(ctx context.Context, report *gtsmodel.Report) error {
	// Populate model.
	if err := f.state.DB.PopulateReport(ctx, report); err != nil {
		return gtserror.Newf("error populating report: %w", err)
	}

	// Do nothing if report target
	// is not remote account.
	if report.TargetAccount.IsLocal() {
		return nil
	}

	// Get our instance account from the db:
	// to anonymize the report, we'll deliver
	// using the outbox of the instance account.
	instanceAcct, err := f.state.DB.GetInstanceAccount(ctx, "")
	if err != nil {
		return gtserror.Newf("error getting instance account: %w", err)
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(instanceAcct.OutboxURI)
	if err != nil {
		return err
	}

	targetAccountIRI, err := parseURI(report.TargetAccount.URI)
	if err != nil {
		return err
	}

	// Convert report to ActivityStreams Flag.
	flag, err := f.converter.ReportToASFlag(ctx, report)
	if err != nil {
		return gtserror.Newf("error converting report to AS: %w", err)
	}

	// To is not set explicitly on Flags. Instead,
	// address Flag BTo report target account URI.
	// This ensures that our federating actor still
	// knows where to send the report, but the BTo
	// property will be stripped before sending.
	//
	// Happily, BTo does not prevent federating
	// actor from using shared inbox to deliver.
	bTo := streams.NewActivityStreamsBtoProperty()
	bTo.AppendIRI(targetAccountIRI)
	flag.SetActivityStreamsBto(bTo)

	// Send the Flag via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, flag,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			flag, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) MoveAccount(ctx context.Context, account *gtsmodel.Account) error {
	// Do nothing if it's not our
	// account that's been moved.
	if !account.IsLocal() {
		return nil
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(account.OutboxURI)
	if err != nil {
		return err
	}

	// Actor doing the Move.
	actorIRI := account.Move.Origin

	// Destination Actor of the Move.
	targetIRI := account.Move.Target

	followersIRI, err := parseURI(account.FollowersURI)
	if err != nil {
		return err
	}

	// Create a new move.
	move := streams.NewActivityStreamsMove()

	// Set the Move ID.
	if err := ap.SetJSONLDIdStr(move, account.Move.URI); err != nil {
		return err
	}

	// Set the Actor for the Move.
	ap.AppendActorIRIs(move, actorIRI)

	// Set the account's IRI as the 'object' property.
	ap.AppendObjectIRIs(move, actorIRI)

	// Set the target's IRI as the 'target' property.
	ap.AppendTargetIRIs(move, targetIRI)

	// Address the move To followers.
	ap.AppendTo(move, followersIRI)

	// Address the move CC public.
	ap.AppendCc(move, ap.PublicURI())

	// Send the Move via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, move,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			move, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) AcceptInteraction(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
) error {
	// Populate model.
	if err := f.state.DB.PopulateInteractionRequest(ctx, req); err != nil {
		return gtserror.Newf("error populating request: %w", err)
	}

	// Bail if interacting account is ours:
	// we've already accepted internally and
	// shouldn't send an Accept to ourselves.
	if req.InteractingAccount.IsLocal() {
		return nil
	}

	// Bail if account isn't ours:
	// we can't Accept on another
	// instance's behalf. (This
	// should never happen but...)
	if req.TargetAccount.IsRemote() {
		return nil
	}

	// Parse outbox URI.
	outboxIRI, err := parseURI(req.TargetAccount.OutboxURI)
	if err != nil {
		return err
	}

	// Convert req to Accept.
	accept, err := f.converter.InteractionReqToASAccept(ctx, req)
	if err != nil {
		return gtserror.Newf("error converting request to Accept: %w", err)
	}

	// Send the Accept via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, accept,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T for %v via outbox %s: %w",
			accept, req.InteractionType, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) RejectInteraction(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
) error {
	// Populate model.
	if err := f.state.DB.PopulateInteractionRequest(ctx, req); err != nil {
		return gtserror.Newf("error populating request: %w", err)
	}

	// Bail if interacting account is ours:
	// we've already rejected internally and
	// shouldn't send an Reject to ourselves.
	if req.InteractingAccount.IsLocal() {
		return nil
	}

	// Bail if account isn't ours:
	// we can't Reject on another
	// instance's behalf. (This
	// should never happen but...)
	if req.TargetAccount.IsRemote() {
		return nil
	}

	// Parse outbox URI.
	outboxIRI, err := parseURI(req.TargetAccount.OutboxURI)
	if err != nil {
		return err
	}

	// Convert req to Reject.
	reject, err := f.converter.InteractionReqToASReject(ctx, req)
	if err != nil {
		return gtserror.Newf("error converting request to Reject: %w", err)
	}

	// Send the Reject via the Actor's outbox.
	if _, err := f.FederatingActor().Send(
		ctx, outboxIRI, reject,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T for %v via outbox %s: %w",
			reject, req.InteractionType, outboxIRI, err,
		)
	}

	return nil
}
