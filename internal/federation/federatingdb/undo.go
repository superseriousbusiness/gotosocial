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

	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (f *federatingDB) Undo(ctx context.Context, undo vocab.ActivityStreamsUndo) error {
	log.DebugKV(ctx, "undo", serialize{undo})

	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	requestingAcct := activityContext.requestingAcct
	receivingAcct := activityContext.receivingAcct

	for _, object := range ap.ExtractObjects(undo) {
		// Try to get object as vocab.Type,
		// else skip handling (likely) IRI.
		asType := object.GetType()
		if asType == nil {
			continue
		}

		// Check and handle any vocab.Type objects.
		switch name := asType.GetTypeName(); name {

		// UNDO FOLLOW
		case ap.ActivityFollow:
			if err := f.undoFollow(
				ctx,
				receivingAcct,
				requestingAcct,
				undo,
				asType,
			); err != nil {
				return err
			}

		// UNDO LIKE
		case ap.ActivityLike:
			if err := f.undoLike(
				ctx,
				receivingAcct,
				requestingAcct,
				undo,
				asType,
			); err != nil {
				return err
			}

		// UNDO BLOCK
		case ap.ActivityBlock:
			if err := f.undoBlock(
				ctx,
				receivingAcct,
				requestingAcct,
				undo,
				asType,
			); err != nil {
				return err
			}

		// UNDO ANNOUNCE
		case ap.ActivityAnnounce:
			if err := f.undoAnnounce(
				ctx,
				receivingAcct,
				requestingAcct,
				undo,
				asType,
			); err != nil {
				return err
			}

		// UNHANDLED
		default:
			log.Debugf(ctx, "unhandled object type: %s", name)
		}
	}

	return nil
}

func (f *federatingDB) undoFollow(
	ctx context.Context,
	receivingAcct *gtsmodel.Account,
	requestingAcct *gtsmodel.Account,
	undo vocab.ActivityStreamsUndo,
	t vocab.Type,
) error {
	asFollow, ok := t.(vocab.ActivityStreamsFollow)
	if !ok {
		err := fmt.Errorf("%T not parseable as vocab.ActivityStreamsFollow", t)
		return gtserror.SetMalformed(err)
	}

	// Make sure the Undo
	// actor owns the target.
	if !sameActor(
		undo.GetActivityStreamsActor(),
		asFollow.GetActivityStreamsActor(),
	) {
		// Ignore this Activity.
		return nil
	}

	// Convert AS Follow to barebones *gtsmodel.Follow,
	// retrieving origin + target accts from the db.
	follow, err := f.converter.ASFollowToFollow(
		gtscontext.SetBarebones(ctx),
		asFollow,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error converting AS Follow to follow: %w", err)
		return err
	}

	// We were missing origin or
	// target for this Follow, so
	// we cannot Undo anything.
	if follow == nil {
		return nil
	}

	// Ensure addressee is follow target.
	if follow.TargetAccountID != receivingAcct.ID {
		const text = "receivingAcct was not Follow target"
		return gtserror.NewErrorForbidden(errors.New(text), text)
	}

	// Ensure requester is follow origin.
	if follow.AccountID != requestingAcct.ID {
		const text = "requestingAcct was not Follow origin"
		return gtserror.NewErrorForbidden(errors.New(text), text)
	}

	// Delete any existing follow with this URI.
	err = f.state.DB.DeleteFollowByURI(ctx, follow.URI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error deleting follow: %w", err)
		return err
	}

	// Delete any existing follow request with this URI.
	err = f.state.DB.DeleteFollowRequestByURI(ctx, follow.URI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error deleting follow request: %w", err)
		return err
	}

	log.Debug(ctx, "Follow undone")
	return nil
}

func (f *federatingDB) undoLike(
	ctx context.Context,
	receivingAcct *gtsmodel.Account,
	requestingAcct *gtsmodel.Account,
	undo vocab.ActivityStreamsUndo,
	t vocab.Type,
) error {
	asLike, ok := t.(vocab.ActivityStreamsLike)
	if !ok {
		err := fmt.Errorf("%T not parseable as vocab.ActivityStreamsLike", t)
		return gtserror.SetMalformed(err)
	}

	// Make sure the Undo
	// actor owns the target.
	if !sameActor(
		undo.GetActivityStreamsActor(),
		asLike.GetActivityStreamsActor(),
	) {
		// Ignore this Activity.
		return nil
	}

	// Convert AS Like to barebones *gtsmodel.StatusFave,
	// retrieving liking acct and target status from the DB.
	fave, err := f.converter.ASLikeToFave(
		gtscontext.SetBarebones(ctx),
		asLike,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error converting AS Like to fave: %w", err)
		return err
	}

	// We were missing status, account,
	// or other for this Like, so we
	// cannot Undo anything.
	if fave == nil {
		return nil
	}

	// Ensure addressee is fave target.
	if fave.TargetAccountID != receivingAcct.ID {
		const text = "receivingAcct was not Fave target"
		return gtserror.NewErrorForbidden(errors.New(text), text)
	}

	// Ensure requester is fave origin.
	if fave.AccountID != requestingAcct.ID {
		const text = "requestingAcct was not Fave origin"
		return gtserror.NewErrorForbidden(errors.New(text), text)
	}

	// Fetch fave from the DB so we know the ID to delete it.
	//
	// Ignore URI on Likes, since we often get multiple Likes
	// with the same target and account ID, but differing URIs.
	// Instead, we'll select using account and target status.
	//
	// Regardless of the URI, we can read an Undo Like to mean
	// "I don't want to fave this post anymore".
	fave, err = f.state.DB.GetStatusFave(
		gtscontext.SetBarebones(ctx),
		fave.AccountID,
		fave.StatusID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf(
			"db error getting fave from %s targeting %s: %w",
			fave.AccountID, fave.StatusID, err,
		)
		return err
	}

	if fave == nil {
		// We didn't have this fave
		// stored anyway, so we can't
		// Undo it, just ignore.
		return nil
	}

	// Delete the fave.
	if err := f.state.DB.DeleteStatusFaveByID(ctx, fave.ID); err != nil {
		err := gtserror.Newf("db error deleting fave %s: %w", fave.ID, err)
		return err
	}

	log.Debug(ctx, "Like undone")
	return nil
}

func (f *federatingDB) undoBlock(
	ctx context.Context,
	receivingAcct *gtsmodel.Account,
	requestingAcct *gtsmodel.Account,
	undo vocab.ActivityStreamsUndo,
	t vocab.Type,
) error {
	asBlock, ok := t.(vocab.ActivityStreamsBlock)
	if !ok {
		err := fmt.Errorf("%T not parseable as vocab.ActivityStreamsBlock", t)
		return gtserror.SetMalformed(err)
	}

	// Make sure the Undo
	// actor owns the target.
	if !sameActor(
		undo.GetActivityStreamsActor(),
		asBlock.GetActivityStreamsActor(),
	) {
		// Ignore this Activity.
		return nil
	}

	// Convert AS Block to barebones *gtsmodel.Block,
	// retrieving origin + target accts from the DB.
	block, err := f.converter.ASBlockToBlock(
		gtscontext.SetBarebones(ctx),
		asBlock,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error converting AS Block to block: %w", err)
		return err
	}

	// We were missing origin or
	// target for this Block, so
	// we cannot Undo anything.
	if block == nil {
		return nil
	}

	// Ensure addressee is block target.
	if block.TargetAccountID != receivingAcct.ID {
		const text = "receivingAcct was not Block target"
		return gtserror.NewErrorForbidden(errors.New(text), text)
	}

	// Ensure requester is block origin.
	if block.AccountID != requestingAcct.ID {
		const text = "requestingAcct was not Block origin"
		return gtserror.NewErrorForbidden(errors.New(text), text)
	}

	// Delete any existing block.
	err = f.state.DB.DeleteBlockByURI(ctx, block.URI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error deleting block: %w", err)
		return err
	}

	log.Debug(ctx, "Block undone")
	return nil
}

func (f *federatingDB) undoAnnounce(
	ctx context.Context,
	receivingAcct *gtsmodel.Account,
	requestingAcct *gtsmodel.Account,
	undo vocab.ActivityStreamsUndo,
	t vocab.Type,
) error {
	asAnnounce, ok := t.(vocab.ActivityStreamsAnnounce)
	if !ok {
		err := fmt.Errorf("%T not parseable as vocab.ActivityStreamsAnnounce", t)
		return gtserror.SetMalformed(err)
	}

	// Make sure the Undo actor owns the
	// Announce they're trying to undo.
	if !sameActor(
		undo.GetActivityStreamsActor(),
		asAnnounce.GetActivityStreamsActor(),
	) {
		// Ignore this Activity.
		return nil
	}

	// Convert AS Announce to *gtsmodel.Status,
	// retrieving origin account + target status.
	boost, isNew, err := f.converter.ASAnnounceToStatus(
		// Use barebones as we don't
		// need to populate attachments
		// on boosted status, mentions, etc.
		gtscontext.SetBarebones(ctx),
		asAnnounce,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error converting AS Announce to boost: %w", err)
		return err
	}

	if boost == nil {
		// We were missing origin or
		// target(s) for this Announce,
		// so we cannot Undo anything.
		return nil
	}

	if isNew {
		// We hadn't seen this boost
		// before anyway, so there's
		// nothing to Undo.
		return nil
	}

	// Ensure requester == announcer.
	if boost.AccountID != requestingAcct.ID {
		const text = "requestingAcct was not Block origin"
		return gtserror.NewErrorForbidden(errors.New(text), text)
	}

	// Looks valid. Process side effects asynchronously.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActivityAnnounce,
		APActivityType: ap.ActivityUndo,
		GTSModel:       boost,
		Receiving:      receivingAcct,
		Requesting:     requestingAcct,
	})

	return nil
}
