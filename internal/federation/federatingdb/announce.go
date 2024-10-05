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
	"net/url"
	"slices"

	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (f *federatingDB) Announce(ctx context.Context, announce vocab.ActivityStreamsAnnounce) error {
	log.DebugKV(ctx, "announce", serialize{announce})

	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	requestingAcct := activityContext.requestingAcct
	receivingAcct := activityContext.receivingAcct

	if requestingAcct.IsMoving() {
		// A Moving account
		// can't do this.
		return nil
	}

	// Ensure requestingAccount is among
	// the Actors doing the Announce.
	//
	// We don't support Announce forwards.
	actorIRIs := ap.GetActorIRIs(announce)
	if !slices.ContainsFunc(actorIRIs, func(actorIRI *url.URL) bool {
		return actorIRI.String() == requestingAcct.URI
	}) {
		return gtserror.Newf(
			"requestingAccount %s was not among Announce Actors",
			requestingAcct.URI,
		)
	}

	boost, isNew, err := f.converter.ASAnnounceToStatus(ctx, announce)
	if err != nil {
		return gtserror.Newf("error converting announce to boost: %w", err)
	}

	if !isNew {
		// We've already seen this boost;
		// nothing else to do here.
		return nil
	}

	// This is a new boost. Process side effects asynchronously.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActivityAnnounce,
		APActivityType: ap.ActivityCreate,
		GTSModel:       boost,
		Receiving:      receivingAcct,
		Requesting:     requestingAcct,
	})

	return nil
}
