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
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// Delete removes the entry with the given id.
//
// Delete is only called for federated objects. Deletes from the Social
// Protocol instead call Update to create a Tombstone.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Delete(ctx context.Context, id *url.URL) error {
	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	// Extract receiving / requesting accounts.
	requesting := activityContext.requestingAcct
	receiving := activityContext.receivingAcct

	// Serialize ID URI.
	uriStr := id.String()

	var (
		ok  bool
		err error
	)

	// Attempt to delete account.
	ok, err = f.deleteAccount(ctx,
		requesting,
		receiving,
		uriStr,
	)
	if err != nil || ok { // handles success
		return err
	}

	// Attempt to delete status.
	ok, err = f.deleteStatus(ctx,
		requesting,
		receiving,
		uriStr,
	)
	if err != nil || ok { // handles success
		return err
	}

	// Log at warning level, as lots of these could indicate federation
	// issues between remote and this instance, or help with debugging.
	log.Warnf(ctx, "received delete for unknown target: %s", uriStr)
	return nil
}

func (f *federatingDB) deleteAccount(
	ctx context.Context,
	requesting *gtsmodel.Account,
	receiving *gtsmodel.Account,
	uri string, // target account
) (
	bool, // success?
	error, // any error
) {
	account, err := f.state.DB.GetAccountByURI(ctx, uri)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return false, gtserror.Newf("error getting account: %w", err)
	}

	if account != nil {
		if account.ID != requesting.ID {
			const text = "signing account does not match delete target"
			return false, gtserror.NewErrorForbidden(err, text)
		}

		log.Debugf(ctx, "deleting account: %s", account.URI)

		// Drop any outgoing queued AP requests to / from / targeting
		// this account, (stops queued likes, boosts, creates etc).
		f.state.Workers.Delivery.Queue.Delete("ObjectID", account.URI)
		f.state.Workers.Delivery.Queue.Delete("TargetID", account.URI)

		// Drop any incoming queued client messages to / from this
		// account, (stops processing of local origin data for acccount).
		f.state.Workers.Client.Queue.Delete("Target.ID", account.ID)
		f.state.Workers.Client.Queue.Delete("TargetURI", account.URI)

		// Drop any incoming queued federator messages to this account,
		// (stops processing of remote origin data targeting this account).
		f.state.Workers.Federator.Queue.Delete("Requesting.ID", account.ID)
		f.state.Workers.Federator.Queue.Delete("TargetURI", account.URI)

		// Only AFTER we have finished purging queues do we enqueue,
		// otherwise we risk purging our own delete message from queue!
		f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
			APObjectType:   ap.ObjectProfile,
			APActivityType: ap.ActivityDelete,
			GTSModel:       account,
			Receiving:      receiving,
			Requesting:     requesting,
		})

		return true, nil
	}

	return false, nil
}

func (f *federatingDB) deleteStatus(
	ctx context.Context,
	requesting *gtsmodel.Account,
	receiving *gtsmodel.Account,
	uri string, // target status
) (
	bool, // success?
	error, // any error
) {
	status, err := f.state.DB.GetStatusByURI(ctx, uri)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return false, gtserror.Newf("error getting status: %w", err)
	}

	if status != nil {
		if status.AccountID != requesting.ID {
			const text = "signing account does not match delete target owner"
			return false, gtserror.NewErrorForbidden(err, text)
		}

		log.Debugf(ctx, "deleting status: %s", status.URI)

		// Drop any outgoing queued AP requests about / targeting
		// this status, (stops queued likes, boosts, creates etc).
		f.state.Workers.Delivery.Queue.Delete("ObjectID", status.URI)
		f.state.Workers.Delivery.Queue.Delete("TargetID", status.URI)

		// Drop any incoming queued client messages about / targeting
		// status, (stops processing of local origin data for status).
		f.state.Workers.Client.Queue.Delete("TargetURI", status.URI)

		// Drop any incoming queued federator messages targeting status,
		// (stops processing of remote origin data targeting this status).
		f.state.Workers.Federator.Queue.Delete("TargetURI", status.URI)

		// Only AFTER we have finished purging queues do we enqueue,
		// otherwise we risk purging our own delete message from queue!
		f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityDelete,
			GTSModel:       status,
			Receiving:      receiving,
			Requesting:     requesting,
		})

		return true, nil
	}

	return false, nil
}
