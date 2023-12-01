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

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
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

	// Ensure target status isn't from a blocked host.
	if blocked, err := d.state.DB.IsDomainBlocked(ctx, targetURIObj.Host); err != nil {
		return nil, gtserror.Newf("error checking blocked domain: %w", err)
	} else if blocked {
		err = gtserror.Newf("%s is blocked", targetURIObj.Host)
		return nil, gtserror.SetUnretrievable(err)
	}

	// Fetch/deref status being boosted.
	var target *gtsmodel.Status

	if targetURIObj.Host == config.GetHost() {
		// This is a local status, fetch from the database
		target, err = d.state.DB.GetStatusByURI(ctx, targetURI)
	} else {
		// This is a remote status, we need to dereference it.
		target, _, err = d.GetStatusByURI(ctx, requestUser, targetURIObj)
	}

	if err != nil {
		return nil, gtserror.Newf(
			"error getting boost target status %s: %w",
			targetURI, err,
		)
	}

	// Generate an ID for the boost wrapper status.
	boost.ID, err = id.NewULIDFromTime(boost.CreatedAt)
	if err != nil {
		return nil, gtserror.Newf("error generating id: %w", err)
	}

	// Populate remaining fields on
	// the boost wrapper using target.
	boost.Content = target.Content
	boost.ContentWarning = target.ContentWarning
	boost.ActivityStreamsType = target.ActivityStreamsType
	boost.Sensitive = target.Sensitive
	boost.Language = target.Language
	boost.Text = target.Text
	boost.BoostOfID = target.ID
	boost.BoostOf = target
	boost.BoostOfAccountID = target.AccountID
	boost.BoostOfAccount = target.Account
	boost.Visibility = target.Visibility
	boost.Federated = target.Federated
	boost.Boostable = target.Boostable
	boost.Replyable = target.Replyable
	boost.Likeable = target.Likeable

	// Store the boost wrapper status.
	switch err = d.state.DB.PutStatus(ctx, boost); {
	case err == nil:
		// All good baby.

	case errors.Is(err, db.ErrAlreadyExists):
		// DATA RACE! We likely lost out to another goroutine
		// in a call to db.Put(Status). Look again in DB by URI.
		boost, err = d.state.DB.GetStatusByURI(ctx, boost.URI)
		if err != nil {
			err = gtserror.Newf(
				"error getting boost wrapper status %s from database after race: %w",
				boost.URI, err,
			)
		}

	default:
		// Proper database error.
		err = gtserror.Newf("db error inserting status: %w", err)
	}

	return boost, err
}
