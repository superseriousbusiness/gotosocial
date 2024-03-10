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

// Package gtsmodel contains types used *internally* by GoToSocial and added/removed/selected from the database.
// These types should never be serialized and/or sent out via public APIs, as they contain sensitive information.
// The annotation used on these structs is for handling them via the bun-db ORM.
// See here for more info on bun model annotations: https://bun.uptrace.dev/guide/models.html

package federatingdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (f *federatingDB) Move(ctx context.Context, move vocab.ActivityStreamsMove) error {
	if log.Level() >= level.DEBUG {
		i, err := marshalItem(move)
		if err != nil {
			return err
		}
		l := log.WithContext(ctx).
			WithField("move", i)
		l.Debug("entering Move")
	}

	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		// Already processed.
		return nil
	}

	requestingAcct := activityContext.requestingAcct
	receivingAcct := activityContext.receivingAcct

	if requestingAcct.IsLocal() {
		// We should not be processing
		// a Move sent from our own
		// instance in the federatingDB.
		return nil
	}

	// Basic Move requirements we can
	// check at this point already:
	//
	//   - Move must have ID/URI set.
	//   - Move `object` and `actor` must
	//     be set, and must be the same
	//     as requesting account.
	//   - Move `target` must be set, and
	//     must *not* be the same as
	//     requesting account.
	//   - Move `target` and `object` must
	//     not have been involved in a
	//     successful Move within the
	//     last 7 days.
	//
	// If the Move looks OK at this point,
	// additional requirements and checks
	// will be processed in FromFediAPI.

	// Ensure ID/URI set.
	moveURI := ap.GetJSONLDId(move)
	if moveURI == nil {
		err := errors.New("Move ID/URI was nil")
		return gtserror.SetMalformed(err)
	}
	moveURIStr := moveURI.String()

	// Check `object` property.
	objects := ap.GetObjectIRIs(move)
	if l := len(objects); l != 1 {
		err := fmt.Errorf("Move requires exactly 1 object, had %d", l)
		return gtserror.SetMalformed(err)
	}
	object := objects[0]
	objectStr := object.String()

	if objectStr != requestingAcct.URI {
		err := fmt.Errorf(
			"Move was signed by %s but object was %s",
			requestingAcct.URI, objectStr,
		)
		return gtserror.SetMalformed(err)
	}

	// Check `actor` property.
	actors := ap.GetActorIRIs(move)
	if l := len(actors); l != 1 {
		err := fmt.Errorf("Move requires exactly 1 actor, had %d", l)
		return gtserror.SetMalformed(err)
	}
	actor := actors[0]
	actorStr := actor.String()

	if actorStr != requestingAcct.URI {
		err := fmt.Errorf(
			"Move was signed by %s but actor was %s",
			requestingAcct.URI, actorStr,
		)
		return gtserror.SetMalformed(err)
	}

	// Check `target` property.
	targets := ap.GetTargetIRIs(move)
	if l := len(targets); l != 1 {
		err := fmt.Errorf("Move requires exactly 1 target, had %d", l)
		return gtserror.SetMalformed(err)
	}
	target := targets[0]
	targetStr := target.String()

	if targetStr == requestingAcct.URI {
		err := fmt.Errorf(
			"Move target and origin were the same (%s)",
			targetStr,
		)
		return gtserror.SetMalformed(err)
	}

	// If a Move has been *attempted* within last 5m,
	// that involved the origin and target in any way,
	// then we shouldn't try to reprocess immediately.
	//
	// This avoids the potential DDOS vector of a given
	// origin account spamming out moves to various
	// target accounts, causing loads of dereferences.
	latestMoveAttempt, err := f.state.DB.GetLatestMoveAttemptInvolvingURIs(
		ctx, objectStr, targetStr,
	)
	if err != nil {
		return gtserror.Newf(
			"error checking latest Move attempt involving object %s and target %s: %w",
			objectStr, targetStr, err,
		)
	}

	if !latestMoveAttempt.IsZero() &&
		time.Since(latestMoveAttempt) < 5*time.Minute {
		log.Infof(ctx,
			"object %s or target %s have been involved in a Move attempt within the last 5 minutes, will not process Move",
			objectStr, targetStr,
		)
		return nil
	}

	// If a Move has *succeeded* within the last week
	// that involved the origin and target in any way,
	// then we shouldn't process again for a while.
	latestMoveSuccess, err := f.state.DB.GetLatestMoveSuccessInvolvingURIs(
		ctx, objectStr, targetStr,
	)
	if err != nil {
		return gtserror.Newf(
			"error checking latest Move success involving object %s and target %s: %w",
			objectStr, targetStr, err,
		)
	}

	if !latestMoveSuccess.IsZero() &&
		time.Since(latestMoveSuccess) < 168*time.Hour {
		log.Infof(ctx,
			"object %s or target %s have been involved in a successful Move within the last 7 days, will not process Move",
			objectStr, targetStr,
		)
		return nil
	}

	// This Move looks surface-level legit,
	// and passes our rate-limiting requirements.
	//
	// Create it in the db (or retrieve it) and
	// then process side effects asynchronously.
	var gtsMove *gtsmodel.Move

	// See if we have a move with
	// this ID/URI stored already.
	gtsMove, err = f.state.DB.GetMoveByURI(ctx, moveURIStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := fmt.Errorf("db error retrieving move with URI %s: %w", moveURIStr, err)
		return gtserror.NewErrorInternalError(err)
	}

	if gtsMove != nil {
		// We had a Move with this ID/URI.
		//
		// Make sure the Move we already had
		// stored has the same origin + target.
		if gtsMove.OriginURI != objectStr ||
			gtsMove.TargetURI != targetStr {
			err := fmt.Errorf(
				"Move object %s and/or target %s differ from stored object and target for this ID (%s)",
				objectStr, targetStr, moveURIStr,
			)
			return gtserror.SetMalformed(err)
		}
	}

	// If we didn't have a move stored for
	// this ID/URI, then see if we have a
	// Move with this origin and target
	// already (but a different ID/URI).
	if gtsMove == nil {
		gtsMove, err = f.state.DB.GetMoveByOriginTarget(ctx, objectStr, targetStr)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err := fmt.Errorf(
				"db error retrieving Move with object %s and target %s: %w",
				objectStr, targetStr, err,
			)
			return gtserror.NewErrorInternalError(err)
		}

		if gtsMove != nil {
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
			gtsMove.URI = moveURIStr
			if err := f.state.DB.UpdateMove(ctx, gtsMove, "uri"); err != nil {
				err := fmt.Errorf(
					"db error updating Move with object %s and target %s: %w",
					objectStr, targetStr, err,
				)
				return gtserror.NewErrorInternalError(err)
			}
		}
	}

	if gtsMove == nil {
		// If Move is still nil then
		// we didn't have this Move
		// stored yet, so it's new.
		// Store it now!
		gtsMove = &gtsmodel.Move{
			ID:          id.NewULID(),
			AttemptedAt: time.Now(),
			OriginURI:   actorStr,
			Origin:      actor,
			TargetURI:   targetStr,
			Target:      target,
			URI:         moveURIStr,
		}
		if err := f.state.DB.PutMove(ctx, gtsMove); err != nil {
			err := fmt.Errorf("db error storing move %s: %w", moveURIStr, err)
			return gtserror.NewErrorInternalError(err)
		}
	}

	// If move_id isn't set on the requesting
	// account yet, set it so other processes
	// know there's a Move in progress.
	if requestingAcct.MoveID != gtsMove.ID {
		requestingAcct.Move = gtsMove
		requestingAcct.MoveID = gtsMove.ID
		if err := f.state.DB.UpdateAccount(ctx,
			requestingAcct, "move_id",
		); err != nil {
			err := fmt.Errorf("db error updating move_id on account: %w", err)
			return gtserror.NewErrorInternalError(err)
		}
	}

	// We had a Move already or stored a new Move.
	// Pass back to a worker for async processing.
	f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
		APObjectType:     ap.ObjectProfile,
		APActivityType:   ap.ActivityMove,
		GTSModel:         gtsMove,
		ReceivingAccount: receivingAcct,
	})

	return nil
}
