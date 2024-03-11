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

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// MoveAccount processes the given
// Move FromFediAPI message:
//
//	APObjectType:     "Profile"
//	APActivityType:   "Move"
//	GTSModel:         *gtsmodel.Move.
//	ReceivingAccount: Account of inbox owner receiving the Move.
func (p *fediAPI) MoveAccount(ctx context.Context, fMsg messages.FromFediAPI) error {
	// The account who received the Move message.
	receiver := fMsg.ReceivingAccount

	// gtsmodel Move activity.
	move, ok := fMsg.GTSModel.(*gtsmodel.Move)
	if !ok {
		return gtserror.Newf(
			"%T not parseable as *gtsmodel.Move",
			fMsg.GTSModel,
		)
	}

	// Move origin and target info.
	var (
		originAcctURIStr = move.OriginURI
		originAcct       = fMsg.RequestingAccount
		targetAcctURIStr = move.TargetURI
		targetAcctURI    = move.Target
	)

	// Assemble log context.
	l := log.
		WithContext(ctx).
		WithField("originAcct", originAcctURIStr).
		WithField("targetAcct", targetAcctURIStr)

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

	// If movedToURI is set on originAcct, make
	// sure it's actually to the intended target.
	//
	// If it's not set, that's fine, we don't
	// really need it. We know by now that the
	// Move was really sent to us by originAcct.
	movedToURI := originAcct.MovedToURI
	if movedToURI != "" &&
		movedToURI != targetAcctURIStr {
		l.Infof(
			"origin account movedTo is set to %s, which differs from Move target; will not process Move",
			movedToURI,
		)
		return nil
	}

	/*
		At this point we have an up-to-date
		model of the Move origin account.

		Now we need to get the target account.
	*/

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

	// Account to which the Move is taking place.
	var (
		targetAcct     *gtsmodel.Account
		targetAcctable ap.Accountable
	)

	if targetAcctURI.Host == config.GetHost() {
		// Target account is ours,
		// just get from the db.
		targetAcct, err = p.state.DB.GetAccountByURI(
			ctx,
			targetAcctURIStr,
		)
	} else {
		// Target account is not ours;
		// try to get from db but deref
		// from remote instance if necessary.
		targetAcct, targetAcctable, err = p.federate.GetAccountByURI(
			ctx,
			receiver.Username,
			targetAcctURI,
		)
	}
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
			receiver.Username,
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

	/*
		At this point we have an up-to-date version
		of the Move origin account (not ours), and
		also the Move target account (possibly ours).
	*/

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

	var errs gtserror.MultiError

	// Transfer originAcct's followers
	// on this instance to targetAcct.
	if err := p.RedirectAccountFollowers(
		ctx,
		originAcct,
		targetAcct,
	); err != nil {
		errs.Append(err)
	}

	// Remove follows on this
	// instance owned by originAcct.
	if err := p.RemoveAccountFollowing(
		ctx,
		originAcct,
		targetAcct,
	); err != nil {
		errs.Append(err)
	}

	// Whatever happened above, error or
	// not, we've just at least attempted
	// the Move so we'll need to update it.
	move.AttemptedAt = time.Now()
	updateColumns := []string{"attempted_at"}

	if err := errs.Combine(); err != nil {
		// We tried to process but
		// didn't succeed completely.
		l.Infof("one or more errors procesing Move side effects: %v", err)
	} else {
		// No errors means we can mark the
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

// RedirectAccountFollowers redirects all local
// followers of originAcct to targetAcct.
//
// Both accounts must be fully dereferenced
// already, and the Move must be valid.
//
// Callers to this function MUST have obtained
// a lock already by calling FedLocks.Lock.
func (p *fediAPI) RedirectAccountFollowers(
	ctx context.Context,
	originAcct *gtsmodel.Account,
	targetAcct *gtsmodel.Account,
) error {
	var errs gtserror.MultiError

	// Any local followers of originAcct should
	// send follow requests to targetAcct instead,
	// and have followers of originAcct removed.
	//
	// Select local followers with barebones, since
	// we only need follow.Account and we can get
	// that ourselves.
	followers, err := p.state.DB.GetAccountLocalFollowers(
		gtscontext.SetBarebones(ctx),
		originAcct.ID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		errs.Appendf(
			"db error getting follows targeting originAcct: %w",
			err,
		)

		// Shouldn't do anything
		// else if this happens.
		return errs.Combine()
	}

	for _, follow := range followers {
		// Populate the local account that
		// owns the follow targeting originAcct.
		if follow.Account, err = p.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			follow.AccountID,
		); err != nil {
			errs.Appendf(
				"db error getting follow account %s: %w",
				follow.AccountID, err,
			)

			// Skip this one.
			continue
		}

		// Use the account processor FollowCreate
		// function to send off the new follow,
		// carrying over the Reblogs and Notify
		// values from the old follow to the new.
		//
		// This will also handle cases where our
		// account has already followed the target
		// account, by just updating the existing
		// follow of target account.
		if _, err := p.account.FollowCreate(
			ctx,
			follow.Account,
			&apimodel.AccountFollowRequest{
				ID:      targetAcct.ID,
				Reblogs: follow.ShowReblogs,
				Notify:  follow.Notify,
			},
		); err != nil {
			errs.Appendf(
				"error creating new follow for account %s: %w",
				follow.AccountID, err,
			)

			// Skip this one.
			continue
		}

		// New follow is in the process of
		// sending, remove the existing follow.
		// This will send out an Undo Activity for each Follow.
		if _, err := p.account.FollowRemove(
			ctx,
			follow.Account,
			follow.TargetAccountID,
		); err != nil {
			errs.Appendf(
				"error removing old follow for account %s: %w",
				follow.AccountID, err,
			)
		}
	}

	return errs.Combine()
}

func (p *fediAPI) RemoveAccountFollowing(
	ctx context.Context,
	originAcct *gtsmodel.Account,
	targetAcct *gtsmodel.Account,
) error {
	var errs gtserror.MultiError

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
		errs.Appendf(
			"db error getting follows owned by originAcct: %w",
			err,
		)

		// Shouldn't do anything
		// else if this happens.
		return errs.Combine()
	}

	for _, follow := range following {
		// Ditch it. This is a one-way action
		// from our side so we don't need to
		// send any messages this time.
		if err := p.state.DB.DeleteFollowByID(ctx, follow.ID); err != nil {
			errs.Appendf(
				"error removing old follow owned by account %s: %w",
				follow.AccountID, err,
			)
		}
	}

	// Finally delete any follow requests
	// owned by or targeting the originAcct.
	if err := p.state.DB.DeleteAccountFollowRequests(
		ctx, originAcct.ID,
	); err != nil {
		errs.Appendf(
			"db error deleting follow requests involving originAcct %s: %w",
			originAcct.URI, err,
		)
	}

	return errs.Combine()
}
