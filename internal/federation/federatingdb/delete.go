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

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
)

// Delete removes the entry with the given id.
//
// Delete is only called for federated objects. Deletes from the Social
// Protocol instead call Update to create a Tombstone.
//
// The library makes this call only after acquiring a lock first.
func (f *DB) Delete(ctx context.Context, id *url.URL) error {
	log.DebugKV(ctx, "id", id)

	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	// Extract receiving / requesting accounts.
	requesting := activityContext.requestingAcct
	receiving := activityContext.receivingAcct

	// Serialize deleted ID URI.
	// (may be status OR account)
	uriStr := id.String()

	var (
		ok  bool
		err error
	)

	// Try delete as an account URI.
	ok, err = f.deleteAccount(ctx,
		requesting,
		receiving,
		uriStr,
	)
	if err != nil {
		return err
	} else if ok {
		// success!
		return nil
	}

	// Try delete as a status URI.
	ok, err = f.deleteStatus(ctx,
		requesting,
		receiving,
		uriStr,
	)
	if err != nil {
		return err
	} else if ok {
		// success!
		return nil
	}

	log.Debugf(ctx, "unknown iri: %s", uriStr)
	return nil
}

func (f *DB) deleteAccount(
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
		// Ensure requesting account is
		// only trying to delete itself.
		if account.ID != requesting.ID {

			// TODO: handled forwarded deletes,
			// for now we silently drop this.
			return true, nil
		}

		log.Debugf(ctx, "deleting account: %s", account.URI)
		f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
			APObjectType:   ap.ActorPerson,
			APActivityType: ap.ActivityDelete,
			GTSModel:       account,
			Receiving:      receiving,
			Requesting:     requesting,
		})

		return true, nil
	}

	return false, nil
}

func (f *DB) deleteStatus(
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
		// Ensure requesting account is only
		// trying to delete its own statuses.
		if status.AccountID != requesting.ID {

			// TODO: handled forwarded deletes,
			// for now we silently drop this.
			return true, nil
		}

		log.Debugf(ctx, "deleting status: %s", status.URI)
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
