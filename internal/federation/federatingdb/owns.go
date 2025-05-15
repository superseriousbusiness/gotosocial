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
	"net/url"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
)

// Owns returns true if the IRI belongs to this instance, and if
// the database has an entry for the IRI.
// The library makes this call only after acquiring a lock first.
func (f *DB) Owns(ctx context.Context, id *url.URL) (bool, error) {
	log.DebugKV(ctx, "id", id)

	// if the id host isn't this instance host, we don't own this IRI
	if host := config.GetHost(); id.Host != host {
		log.Tracef(ctx, "we DO NOT own activity because the host is %s not %s", id.Host, host)
		return false, nil
	}

	// todo: refactor the below; make sure we use
	// proper db functions for everything, and
	// preferably clean up by calling subfuncs
	// (like we now do for ownsLike).

	// apparently it belongs to this host, so what *is* it?
	// check if it's a status, eg /users/example_username/statuses/SOME_UUID_OF_A_STATUS
	if uris.IsStatusesPath(id) {
		_, uid, err := uris.ParseStatusesPath(id)
		if err != nil {
			return false, fmt.Errorf("error parsing statuses path for url %s: %s", id.String(), err)
		}
		status, err := f.state.DB.GetStatusByURI(ctx, uid)
		if err != nil {
			if err == db.ErrNoEntries {
				// there are no entries for this status
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching status with id %s: %s", uid, err)
		}
		return *status.Local, nil
	}

	if uris.IsUserPath(id) {
		username, err := uris.ParseUserPath(id)
		if err != nil {
			return false, fmt.Errorf("error parsing statuses path for url %s: %s", id.String(), err)
		}
		if _, err := f.state.DB.GetAccountByUsernameDomain(ctx, username, ""); err != nil {
			if err == db.ErrNoEntries {
				// there are no entries for this username
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching account with username %s: %s", username, err)
		}
		log.Debugf(ctx, "we own url %s", id)
		return true, nil
	}

	if uris.IsFollowersPath(id) {
		username, err := uris.ParseFollowersPath(id)
		if err != nil {
			return false, fmt.Errorf("error parsing statuses path for url %s: %s", id.String(), err)
		}
		if _, err := f.state.DB.GetAccountByUsernameDomain(ctx, username, ""); err != nil {
			if err == db.ErrNoEntries {
				// there are no entries for this username
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching account with username %s: %s", username, err)
		}
		log.Debugf(ctx, "we own url %s", id)
		return true, nil
	}

	if uris.IsFollowingPath(id) {
		username, err := uris.ParseFollowingPath(id)
		if err != nil {
			return false, fmt.Errorf("error parsing statuses path for url %s: %s", id.String(), err)
		}
		if _, err := f.state.DB.GetAccountByUsernameDomain(ctx, username, ""); err != nil {
			if err == db.ErrNoEntries {
				// there are no entries for this username
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching account with username %s: %s", username, err)
		}
		log.Debugf(ctx, "we own url %s", id)
		return true, nil
	}

	if uris.IsLikePath(id) {
		return f.ownsLike(ctx, id)
	}

	if uris.IsBlockPath(id) {
		username, blockID, err := uris.ParseBlockPath(id)
		if err != nil {
			return false, fmt.Errorf("error parsing block path for url %s: %s", id.String(), err)
		}
		if _, err := f.state.DB.GetAccountByUsernameDomain(ctx, username, ""); err != nil {
			if err == db.ErrNoEntries {
				// there are no entries for this username
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching account with username %s: %s", username, err)
		}
		if err := f.state.DB.GetByID(ctx, blockID, &gtsmodel.Block{}); err != nil {
			if err == db.ErrNoEntries {
				// there are no entries
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching block with id %s: %s", blockID, err)
		}
		log.Debugf(ctx, "we own url %s", id)
		return true, nil
	}

	return false, fmt.Errorf("could not match activityID: %s", id.String())
}

func (f *DB) ownsLike(ctx context.Context, uri *url.URL) (bool, error) {
	username, id, err := uris.ParseLikedPath(uri)
	if err != nil {
		return false, fmt.Errorf("error parsing Like path for url %s: %w", uri.String(), err)
	}

	// We're only checking for existence,
	// so use barebones context.
	bbCtx := gtscontext.SetBarebones(ctx)

	if _, err := f.state.DB.GetAccountByUsernameDomain(bbCtx, username, ""); err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// No entries for this acct,
			// we don't own this item.
			return false, nil
		}

		// Actual error.
		return false, fmt.Errorf("database error fetching account with username %s: %w", username, err)
	}

	if _, err := f.state.DB.GetStatusFaveByID(bbCtx, id); err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// No entries for this ID,
			// we don't own this item.
			return false, nil
		}

		// Actual error.
		return false, fmt.Errorf("database error fetching status fave with id %s: %w", id, err)
	}

	log.Tracef(ctx, "we own Like %s", uri.String())
	return true, nil
}
