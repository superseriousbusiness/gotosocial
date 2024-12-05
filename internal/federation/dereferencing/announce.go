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

package dereferencing

import (
	"context"
	"errors"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// EnrichAnnounce enriches the given boost wrapper status
// by either fetching from the DB or dereferencing the target
// status, populating the boost wrapper's fields based on the
// target status, and then storing the wrapper in the database.
// The wrapper is then returned to the caller.
//
// The provided boost wrapper status must have BoostOfURI set.
func (d *Dereferencer) EnrichAnnounce(
	ctx context.Context,
	boost *gtsmodel.Status,
	requestUser string,
) (*gtsmodel.Status, error) {
	targetURI := boost.BoostOfURI
	if targetURI == "" {
		// We can't do anything.
		return nil, gtserror.Newf("no URI to dereference")
	}

	// Parse the boost target status URI.
	targetURIObj, err := url.Parse(targetURI)
	if err != nil {
		return nil, gtserror.Newf(
			"couldn't parse boost target status URI %s: %w",
			targetURI, err,
		)
	}

	// Fetch and dereference status being boosted, noting that
	// d.GetStatusByURI handles domain blocks and local statuses.
	target, _, err := d.GetStatusByURI(ctx, requestUser, targetURIObj)
	if err != nil {
		return nil, gtserror.Newf("error fetching boost target %s: %w", targetURI, err)
	}

	if target.BoostOfID != "" {
		// Ensure that the target is not a boost (should not be possible).
		err := gtserror.Newf("target status %s is a boost", targetURI)
		return nil, err
	}

	// Set boost_of_uri again in case the
	// original URI was an indirect link.
	boost.BoostOfURI = target.URI

	// Boosts are not considered sensitive even if their target is.
	boost.Sensitive = util.Ptr(false)

	// Populate remaining fields on
	// the boost wrapper using target.
	boost.ActivityStreamsType = target.ActivityStreamsType
	boost.BoostOfID = target.ID
	boost.BoostOf = target
	boost.BoostOfAccountID = target.AccountID
	boost.BoostOfAccount = target.Account
	boost.Visibility = target.Visibility
	boost.Federated = target.Federated

	// Ensure this Announce is permitted by the Announcee.
	permit, err := d.isPermittedStatus(ctx, requestUser, nil, boost, true)
	if err != nil {
		return nil, gtserror.Newf("error checking permitted status %s: %w", boost.URI, err)
	}

	if !permit {
		// Return a checkable error type that can be ignored.
		err := gtserror.Newf("dropping unpermitted status: %s", boost.URI)
		return nil, gtserror.SetNotPermitted(err)
	}

	// Generate an ID for the boost wrapper status.
	boost.ID = id.NewULIDFromTime(boost.CreatedAt)

	// Store the boost wrapper status in database.
	switch err = d.state.DB.PutStatus(ctx, boost); {
	case err == nil:
		// all groovy.

	case errors.Is(err, db.ErrAlreadyExists):
		uri := boost.URI

		// DATA RACE! We likely lost out to another goroutine
		// in a call to db.Put(Status). Look again in DB by URI.
		boost, err = d.state.DB.GetStatusByURI(ctx, uri)
		if err != nil {
			return nil, gtserror.Newf(
				"error getting boost wrapper status %s from database after race: %w",
				uri, err,
			)
		}

	default: // Proper database error.
		return nil, gtserror.Newf("db error inserting status: %w", err)
	}

	return boost, err
}
