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

	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// Update sets an existing entry to the database based on the value's
// id.
//
// Note that Activity values received from federated peers may also be
// updated in the database this way if the Federating Protocol is
// enabled. The client may freely decide to store only the id instead of
// the entire value.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Update(ctx context.Context, asType vocab.Type) error {
	log.DebugKV(ctx, "update", serialize{asType})

	// Mark activity as handled.
	f.storeActivityID(asType)

	// Extract relevant values from passed ctx.
	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	requesting := activityContext.requestingAcct
	receiving := activityContext.receivingAcct

	if accountable, ok := ap.ToAccountable(asType); ok {
		return f.updateAccountable(ctx, receiving, requesting, accountable)
	}

	if statusable, ok := ap.ToStatusable(asType); ok {
		return f.updateStatusable(ctx, receiving, requesting, statusable)
	}

	log.Debugf(ctx, "unhandled object type: %T", asType)
	return nil
}

func (f *federatingDB) updateAccountable(ctx context.Context, receivingAcct *gtsmodel.Account, requestingAcct *gtsmodel.Account, accountable ap.Accountable) error {
	// Extract AP URI of the updated Accountable model.
	idProp := accountable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return gtserror.New("invalid id prop")
	}

	// Get the account URI string for checks
	accountURI := idProp.GetIRI()
	accountURIStr := accountURI.String()

	// Don't try to update local accounts.
	if accountURI.Host == config.GetHost() {
		return nil
	}

	// Check that update was by the account themselves.
	if accountURIStr != requestingAcct.URI {
		return gtserror.Newf("update for %s was not requested by owner", accountURIStr)
	}

	// Pass in to the processor the existing version of the requesting
	// account that we have, plus the Accountable representation that
	// was delivered along with the Update, for further asynchronous
	// updating of eg., avatar/header, emojis, etc. The actual db
	// inserts/updates will take place there.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActorPerson,
		APActivityType: ap.ActivityUpdate,
		GTSModel:       requestingAcct,
		APObject:       accountable,
		Receiving:      receivingAcct,
		Requesting:     requestingAcct,
	})

	return nil
}

func (f *federatingDB) updateStatusable(ctx context.Context, receivingAcct *gtsmodel.Account, requestingAcct *gtsmodel.Account, statusable ap.Statusable) error {
	// Extract AP URI of the updated model.
	idProp := statusable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return gtserror.New("invalid id prop")
	}

	// Get the status URI string for lookups.
	statusURI := idProp.GetIRI()
	statusURIStr := statusURI.String()

	// Don't try to update local statuses.
	if statusURI.Host == config.GetHost() {
		return nil
	}

	// Check if this is a forwarded object, i.e. did
	// the account making the request also create this?
	forwarded := !isSender(statusable, requestingAcct)

	// Get the status we have on file for this URI string.
	status, err := f.state.DB.GetStatusByURI(ctx, statusURIStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("error fetching status from db: %w", err)
	}

	if status == nil {
		// We haven't seen this status before, be
		// lenient and handle as a CREATE event.
		return f.createStatusable(ctx,
			receivingAcct,
			requestingAcct,
			statusable,
			forwarded,
		)
	}

	if forwarded {
		// For forwarded updates, set a nil AS
		// status to force refresh from remote.
		statusable = nil
	}

	// Queue an UPDATE NOTE activity to our fedi API worker,
	// this will handle necessary database insertions, etc.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityUpdate,
		GTSModel:       status, // original status
		APObject:       (ap.Statusable)(statusable),
		Receiving:      receivingAcct,
		Requesting:     requestingAcct,
	})

	return nil
}
