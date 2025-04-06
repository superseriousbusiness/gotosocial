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
	"errors"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// ShouldProcessMove checks whether we should attempt
// to process a move with the given object and target,
// based on whether or not a move with those values
// was attempted or succeeded recently.
func (p *fediAPI) ShouldProcessMove(
	ctx context.Context,
	object string,
	target string,
) (bool, error) {
	// If a Move has been *attempted* within last 5m,
	// that involved the origin and target in any way,
	// then we shouldn't try to reprocess immediately.
	//
	// This avoids the potential DDOS vector of a given
	// origin account spamming out moves to various
	// target accounts, causing loads of dereferences.
	latestMoveAttempt, err := p.state.DB.GetLatestMoveAttemptInvolvingURIs(
		ctx, object, target,
	)
	if err != nil {
		return false, gtserror.Newf(
			"error checking latest Move attempt involving object %s and target %s: %w",
			object, target, err,
		)
	}

	if !latestMoveAttempt.IsZero() &&
		time.Since(latestMoveAttempt) < 5*time.Minute {
		log.Infof(ctx,
			"object %s or target %s have been involved in a Move attempt within the last 5 minutes, will not process Move",
			object, target,
		)
		return false, nil
	}

	// If a Move has *succeeded* within the last week
	// that involved the origin and target in any way,
	// then we shouldn't process again for a while.
	latestMoveSuccess, err := p.state.DB.GetLatestMoveSuccessInvolvingURIs(
		ctx, object, target,
	)
	if err != nil {
		return false, gtserror.Newf(
			"error checking latest Move success involving object %s and target %s: %w",
			object, target, err,
		)
	}

	if !latestMoveSuccess.IsZero() &&
		time.Since(latestMoveSuccess) < 168*time.Hour {
		log.Infof(ctx,
			"object %s or target %s have been involved in a successful Move within the last 7 days, will not process Move",
			object, target,
		)
		return false, nil
	}

	return true, nil
}

// GetOrCreateMove takes a stub move created by the
// requesting account, and either retrieves or creates
// a corresponding move in the database. If a move is
// created in this way, requestingAcct will be updated
// with the correct moveID.
func (p *fediAPI) GetOrCreateMove(
	ctx context.Context,
	requestingAcct *gtsmodel.Account,
	stubMove *gtsmodel.Move,
) (*gtsmodel.Move, error) {
	var (
		moveURIStr = stubMove.URI
		objectStr  = stubMove.OriginURI
		object     = stubMove.Origin
		targetStr  = stubMove.TargetURI
		target     = stubMove.Target

		move *gtsmodel.Move
		err  error
	)

	// See if we have a move with
	// this ID/URI stored already.
	move, err = p.state.DB.GetMoveByURI(ctx, moveURIStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.Newf(
			"db error retrieving move with URI %s: %w",
			moveURIStr, err,
		)
	}

	if move != nil {
		// We had a Move with this ID/URI.
		//
		// Make sure the Move we already had
		// stored has the same origin + target.
		if move.OriginURI != objectStr ||
			move.TargetURI != targetStr {
			return nil, gtserror.Newf(
				"Move object %s and/or target %s differ from stored object and target for this ID (%s)",
				objectStr, targetStr, moveURIStr,
			)
		}
	}

	// If we didn't have a move stored for
	// this ID/URI, then see if we have a
	// Move with this origin and target
	// already (but a different ID/URI).
	if move == nil {
		move, err = p.state.DB.GetMoveByOriginTarget(ctx, objectStr, targetStr)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.Newf(
				"db error retrieving Move with object %s and target %s: %w",
				objectStr, targetStr, err,
			)
		}

		if move != nil {
			// We had a move for this object and
			// target, but the ID/URI has changed.
			// Update the Move's URI in the db to
			// reflect that this is but the latest
			// attempt with this origin + target.
			//
			// The remote may be trying to retry
			// the Move but their server might
			// not reuse the same Activity URIs,
			// and we don't want to store a brand
			// new Move for each attempt!
			move.URI = moveURIStr
			if err := p.state.DB.UpdateMove(ctx, move, "uri"); err != nil {
				return nil, gtserror.Newf(
					"db error updating Move with object %s and target %s: %w",
					objectStr, targetStr, err,
				)
			}
		}
	}

	if move == nil {
		// If Move is still nil then
		// we didn't have this Move
		// stored yet, so it's new.
		// Store it now!
		move = &gtsmodel.Move{
			ID:          id.NewULID(),
			AttemptedAt: time.Now(),
			OriginURI:   objectStr,
			Origin:      object,
			TargetURI:   targetStr,
			Target:      target,
			URI:         moveURIStr,
		}
		if err := p.state.DB.PutMove(ctx, move); err != nil {
			return nil, gtserror.Newf(
				"db error storing move %s: %w",
				moveURIStr, err,
			)
		}
	}

	// If move_id isn't set on the requesting
	// account yet, set it so other processes
	// know there's a Move in progress.
	if requestingAcct.MoveID != move.ID {
		requestingAcct.Move = move
		requestingAcct.MoveID = move.ID
		if err := p.state.DB.UpdateAccount(ctx,
			requestingAcct, "move_id",
		); err != nil {
			return nil, gtserror.Newf(
				"db error updating move_id on account: %w",
				err,
			)
		}
	}

	return move, nil
}

// MoveAccount processes the given
// Move FromFediAPI message:
//
//	APObjectType:     "Profile"
//	APActivityType:   "Move"
//	GTSModel:         stub *gtsmodel.Move.
//	ReceivingAccount: Account of inbox owner receiving the Move.
func (p *fediAPI) MoveAccount(ctx context.Context, fMsg *messages.FromFediAPI) error {
	// *gtsmodel.Move activity.
	stubMove, ok := fMsg.GTSModel.(*gtsmodel.Move)
	if !ok {
		return gtserror.Newf(
			"%T not parseable as *gtsmodel.Move",
			fMsg.GTSModel,
		)
	}

	// Move origin and target info.
	var (
		originAcctURIStr = stubMove.OriginURI
		originAcct       = fMsg.Requesting
		targetAcctURIStr = stubMove.TargetURI
		targetAcctURI    = stubMove.Target
	)

	// Assemble log context.
	l := log.
		WithContext(ctx).
		WithField("originAcct", originAcctURIStr).
		WithField("targetAcct", targetAcctURIStr)

	// We can't/won't validate Move activities
	// to domains we have blocked, so check this.
	targetDomainBlocked, err := p.state.DB.IsDomainBlocked(ctx, targetAcctURI.Host)
	if err != nil {
		return gtserror.Newf(
			"db error checking if target domain %s blocked: %w",
			targetAcctURI.Host, err,
		)
	}

	if targetDomainBlocked {
		l.Info("target domain is blocked, will not process Move")
		return nil
	}

	// Next steps require making calls to remote +
	// setting values that may be attempted by other
	// in-process Moves. To avoid race conditions,
	// ensure we're only trying to process this
	// Move combo one attempt at a time.
	//
	// We use a custom lock because remotes might
	// try to send the same Move several times with
	// different IDs (you never know), but we only
	// want to process them based on origin + target.
	unlock := p.state.FedLocks.Lock(
		"move:" + originAcctURIStr + ":" + targetAcctURIStr,
	)
	defer unlock()

	// Check if Move is rate limited based
	// on previous attempts / successes.
	shouldProcess, err := p.ShouldProcessMove(ctx,
		originAcctURIStr, targetAcctURIStr,
	)
	if err != nil {
		return gtserror.Newf(
			"error checking if Move should be processed now: %w",
			err,
		)
	}

	if !shouldProcess {
		// Move is rate limited, so don't process.
		// Reason why should already be logged.
		return nil
	}

	// Store new or retrieve existing Move. This will
	// also update moveID on originAcct if necessary.
	move, err := p.GetOrCreateMove(ctx, originAcct, stubMove)
	if err != nil {
		return gtserror.Newf(
			"error refreshing target account %s: %w",
			targetAcctURIStr, err,
		)
	}

	// Account to which the Move is taking place.
	//
	// Match by uri only.
	targetAcct, targetAcctable, err := p.federate.GetAccountByURI(
		ctx,
		fMsg.Receiving.Username,
		targetAcctURI,
		false,
	)
	if err != nil {
		return gtserror.Newf(
			"error getting target account %s: %w",
			targetAcctURIStr, err,
		)
	}

	// If target is suspended from this instance,
	// then we can't/won't process any move side
	// effects to that account, because:
	//
	//   1. We can't verify that it's aliased correctly
	//      back to originAcct without dereferencing it.
	//   2. We can't/won't forward follows to a suspended
	//      account, since suspension would remove follows
	//      etc. targeting the new account anyways.
	//   3. If someone is moving to a suspended account
	//      they probably totally suck ass (according to
	//      the moderators of this instance, anyway) so
	//      to hell with it.
	if targetAcct.IsSuspended() {
		l.Info("target account is suspended, will not process Move")
		return nil
	}

	if targetAcct.IsRemote() {
		// Force refresh Move target account
		// to ensure we have up-to-date version.
		targetAcct, _, err = p.federate.RefreshAccount(ctx,
			fMsg.Receiving.Username,
			targetAcct,
			targetAcctable,
			dereferencing.Freshest,
		)
		if err != nil {
			return gtserror.Newf(
				"error refreshing target account %s: %w",
				targetAcctURIStr, err,
			)
		}
	}

	// Target must not itself have moved somewhere.
	// You can't move to an already-moved account.
	targetAcctMovedTo := targetAcct.MovedToURI
	if targetAcctMovedTo != "" {
		l.Infof(
			"target account has, itself, already moved to %s, will not process Move",
			targetAcctMovedTo,
		)
		return nil
	}

	// Target must be aliased back to origin account.
	// Ie., its alsoKnownAs values must include the
	// origin account, so we know it's for real.
	if !targetAcct.IsAliasedTo(originAcctURIStr) {
		l.Info("target account is not aliased back to origin account, will not process Move")
		return nil
	}

	/*
		At this point we know that the move
		looks valid and we should process it.
	*/

	// Transfer originAcct's followers
	// on this instance to targetAcct.
	redirectOK := p.utils.redirectFollowers(
		ctx,
		originAcct,
		targetAcct,
	)

	// Remove follows on this
	// instance owned by originAcct.
	removeFollowingOK := p.RemoveAccountFollowing(
		ctx,
		originAcct,
	)

	// Whatever happened above, error or
	// not, we've just at least attempted
	// the Move so we'll need to update it.
	move.AttemptedAt = time.Now()
	updateColumns := []string{"attempted_at"}

	if redirectOK && removeFollowingOK {
		// All OK means we can mark the
		// Move as definitively succeeded.
		//
		// Take same time so SucceededAt
		// isn't 0.0001s later or something.
		move.SucceededAt = move.AttemptedAt
		updateColumns = append(updateColumns, "succeeded_at")
	}

	// Update whatever columns we need to update.
	if err := p.state.DB.UpdateMove(ctx,
		move, updateColumns...,
	); err != nil {
		return gtserror.Newf(
			"db error updating Move %s: %w",
			move.URI, err,
		)
	}

	return nil
}

// RemoveAccountFollowing removes all
// follows owned by the move originAcct.
//
// originAcct must be fully dereferenced
// already, and the Move must be valid.
//
// Callers to this function MUST have obtained
// a lock already by calling FedLocks.Lock.
//
// Return bool will be true if all goes OK.
func (p *fediAPI) RemoveAccountFollowing(
	ctx context.Context,
	originAcct *gtsmodel.Account,
) bool {
	// Any follows owned by originAcct which target
	// accounts on our instance should be removed.
	//
	// We should rely on the target instance
	// to send out new follows from targetAcct.
	following, err := p.state.DB.GetAccountLocalFollows(
		gtscontext.SetBarebones(ctx),
		originAcct.ID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf(ctx,
			"db error getting follows owned by originAcct: %v",
			err,
		)
		return false
	}

	for _, follow := range following {
		// Ditch it. This is a one-way action
		// from our side so we don't need to
		// send any messages this time.
		if err := p.state.DB.DeleteFollowByID(ctx, follow.ID); err != nil {
			log.Errorf(ctx,
				"error removing old follow owned by account %s: %v",
				follow.AccountID, err,
			)
			return false
		}
	}

	// Finally delete any follow requests
	// owned by or targeting the originAcct.
	if err := p.state.DB.DeleteAccountFollowRequests(
		ctx, originAcct.ID,
	); err != nil {
		log.Errorf(ctx,
			"db error deleting follow requests involving originAcct %s: %v",
			originAcct.URI, err,
		)
		return false
	}

	return true
}
