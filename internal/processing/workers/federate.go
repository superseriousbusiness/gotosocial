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
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// federate wraps functions for federating
// something out via ActivityPub in response
// to message processing.
type federate struct {
	state *state.State
	tc    typeutils.TypeConverter

	// send matches the signature of the
	// go-fed FederatingActor's Send function.
	// It can be used for sending the given
	// activity via the given outbox URI.
	send func(context.Context, *url.URL, vocab.Type) (pub.Activity, error)
}

// parseURI is a cheeky little
// shortcut to wrap parsing errors.
func parseURI(s string) (*url.URL, error) {
	uri, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("error parsing uri %s: %w", s, err)
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
		return gtserror.Newf("%w", err)
	}

	actorIRI, err := parseURI(account.URI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	followersIRI, err := parseURI(account.FollowersURI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	publicIRI, err := parseURI(pub.PublicActivityPubIRI)
	if err != nil {
		return gtserror.Newf("%w", err)
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
	deleteCC.AppendIRI(publicIRI)
	delete.SetActivityStreamsCc(deleteCC)

	// Send the Delete via the Actor's outbox.
	if _, err := f.send(
		ctx, outboxIRI, delete,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			delete, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) CreateStatus(ctx context.Context, status *gtsmodel.Status) error {
	// Do nothing if the status
	// shouldn't be federated.
	if !*status.Federated {
		return nil
	}

	// Do nothing if this
	// isn't our status.
	if !*status.Local {
		return nil
	}

	// Populate model.
	if err := f.state.DB.PopulateStatus(ctx, status); err != nil {
		return gtserror.Newf("error populating status: %w", err)
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(status.Account.OutboxURI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	// Convert status to an ActivityStreams
	// Note, wrapped in a Create activity.
	asStatus, err := f.tc.StatusToAS(ctx, status)
	if err != nil {
		return gtserror.Newf("error converting status to AS: %w", err)
	}

	create, err := f.tc.WrapNoteInCreate(asStatus, false)
	if err != nil {
		return gtserror.Newf("error wrapping status in create: %w", err)
	}

	// Send the Create via the Actor's outbox.
	if _, err := f.send(
		ctx, outboxIRI, create,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			create, outboxIRI, err,
		)
	}

	return nil
}

func (f *federate) DeleteStatus(ctx context.Context, status *gtsmodel.Status) error {
	// Do nothing if the status
	// shouldn't be federated.
	if !*status.Federated {
		return nil
	}

	// Do nothing if this
	// isn't our status.
	if !*status.Local {
		return nil
	}

	// Populate model.
	if err := f.state.DB.PopulateStatus(ctx, status); err != nil {
		return gtserror.Newf("error populating status: %w", err)
	}

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(status.Account.OutboxURI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	// Wrap the status URI in a Delete activity.
	delete, err := f.tc.StatusToASDelete(ctx, status)
	if err != nil {
		return gtserror.Newf("error creating Delete: %w", err)
	}

	// Send the Delete via the Actor's outbox.
	if _, err := f.send(
		ctx, outboxIRI, delete,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			delete, outboxIRI, err,
		)
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
		return gtserror.Newf("%w", err)
	}

	// Convert follow to ActivityStreams Follow.
	asFollow, err := f.tc.FollowToAS(ctx, follow)
	if err != nil {
		return gtserror.Newf("error converting follow to AS: %s", err)
	}

	// Send the Follow via the Actor's outbox.
	if _, err := f.send(
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
		return gtserror.Newf("%w", err)
	}

	targetAccountIRI, err := parseURI(follow.TargetAccount.URI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	// Recreate the ActivityStreams Follow.
	asFollow, err := f.tc.FollowToAS(ctx, follow)
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
	if _, err := f.send(
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
		return gtserror.Newf("%w", err)
	}

	targetAccountIRI, err := parseURI(fave.TargetAccount.URI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	// Recreate the ActivityStreams Like.
	like, err := f.tc.FaveToAS(ctx, fave)
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
	if _, err := f.send(
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
		return gtserror.Newf("%w", err)
	}

	// Recreate the ActivityStreams Announce.
	asAnnounce, err := f.tc.BoostToAS(
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
	if _, err := f.send(
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
		return gtserror.Newf("%w", err)
	}

	acceptingAccountIRI, err := parseURI(follow.TargetAccount.URI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	requestingAccountIRI, err := parseURI(follow.Account.URI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	// Recreate the ActivityStreams Follow.
	asFollow, err := f.tc.FollowToAS(ctx, follow)
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
	if _, err := f.send(
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
		return gtserror.Newf("%w", err)
	}

	rejectingAccountIRI, err := parseURI(follow.TargetAccount.URI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	requestingAccountIRI, err := parseURI(follow.Account.URI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	// Recreate the ActivityStreams Follow.
	asFollow, err := f.tc.FollowToAS(ctx, follow)
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
	if _, err := f.send(
		ctx, outboxIRI, reject,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			reject, outboxIRI, err,
		)
	}

	return nil
}

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

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(fave.Account.OutboxURI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	// Create the ActivityStreams Like.
	like, err := f.tc.FaveToAS(ctx, fave)
	if err != nil {
		return gtserror.Newf("error converting fave to AS Like: %w", err)
	}

	// Send the Like via the Actor's outbox.
	if _, err := f.send(
		ctx, outboxIRI, like,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			like, outboxIRI, err,
		)
	}

	return nil
}

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

	// Parse relevant URI(s).
	outboxIRI, err := parseURI(boost.Account.OutboxURI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	// Create the ActivityStreams Announce.
	announce, err := f.tc.BoostToAS(
		ctx,
		boost,
		boost.Account,
		boost.BoostOfAccount,
	)
	if err != nil {
		return gtserror.Newf("error converting boost to AS: %w", err)
	}

	// Send the Announce via the Actor's outbox.
	if _, err := f.send(
		ctx, outboxIRI, announce,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			announce, outboxIRI, err,
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
		return gtserror.Newf("%w", err)
	}

	// Convert account to ActivityStreams Person.
	person, err := f.tc.AccountToAS(ctx, account)
	if err != nil {
		return gtserror.Newf("error converting account to Person: %w", err)
	}

	// Use ActivityStreams Person as Object of Update.
	update, err := f.tc.WrapPersonInUpdate(person, account)
	if err != nil {
		return gtserror.Newf("error wrapping Person in Update: %w", err)
	}

	// Send the Update via the Actor's outbox.
	if _, err := f.send(
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
		return gtserror.Newf("%w", err)
	}

	// Convert block to ActivityStreams Block.
	asBlock, err := f.tc.BlockToAS(ctx, block)
	if err != nil {
		return gtserror.Newf("error converting block to AS: %w", err)
	}

	// Send the Block via the Actor's outbox.
	if _, err := f.send(
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
		return gtserror.Newf("%w", err)
	}

	targetAccountIRI, err := parseURI(block.TargetAccount.URI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	// Convert block to ActivityStreams Block.
	asBlock, err := f.tc.BlockToAS(ctx, block)
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
	if _, err := f.send(
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
		return gtserror.Newf("%w", err)
	}

	targetAccountIRI, err := parseURI(report.TargetAccount.URI)
	if err != nil {
		return gtserror.Newf("%w", err)
	}

	// Convert report to ActivityStreams Flag.
	flag, err := f.tc.ReportToASFlag(ctx, report)
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
	if _, err := f.send(
		ctx, outboxIRI, flag,
	); err != nil {
		return gtserror.Newf(
			"error sending activity %T via outbox %s: %w",
			flag, outboxIRI, err,
		)
	}

	return nil
}
