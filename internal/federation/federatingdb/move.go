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

	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (f *federatingDB) Move(ctx context.Context, move vocab.ActivityStreamsMove) error {
	log.DebugKV(ctx, "move", serialize{move})

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

	// If movedToURI is set on requestingAcct,
	// make sure it points to the intended target.
	//
	// If it's not set, that's fine, we don't
	// need it right now. We know by now that the
	// Move was really sent to us by requestingAcct.
	movedToURI := receivingAcct.MovedToURI
	if movedToURI != "" &&
		movedToURI != targetStr {
		err := fmt.Errorf(
			"origin account movedTo is set to %s, which differs from Move target; will not process Move",
			movedToURI,
		)
		return gtserror.SetMalformed(err)
	}

	// Create a stub *gtsmodel.Move with relevant
	// values. This will be updated / stored by the
	// fedi api worker as necessary.
	stubMove := &gtsmodel.Move{
		OriginURI: objectStr,
		Origin:    object,
		TargetURI: targetStr,
		Target:    target,
		URI:       moveURIStr,
	}

	// We had a Move already or stored a new Move.
	// Pass back to a worker for async processing.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActorPerson,
		APActivityType: ap.ActivityMove,
		GTSModel:       stubMove,
		Requesting:     requestingAcct,
		Receiving:      receivingAcct,
	})

	return nil
}
