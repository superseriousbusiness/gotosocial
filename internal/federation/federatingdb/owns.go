/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package federatingdb

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Owns returns true if the IRI belongs to this instance, and if
// the database has an entry for the IRI.
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Owns(ctx context.Context, id *url.URL) (bool, error) {
	l := logrus.WithFields(
		logrus.Fields{
			"func": "Owns",
			"id":   id,
		},
	)
	l.Debug("entering Owns")

	// if the id host isn't this instance host, we don't own this IRI
	if id.Host != f.config.Host {
		l.Tracef("we DO NOT own activity because the host is %s not %s", id.Host, f.config.Host)
		return false, nil
	}

	// apparently it belongs to this host, so what *is* it?
	// check if it's a status, eg /users/example_username/statuses/SOME_UUID_OF_A_STATUS
	if util.IsStatusesPath(id) {
		_, uid, err := util.ParseStatusesPath(id)
		if err != nil {
			return false, fmt.Errorf("error parsing statuses path for url %s: %s", id.String(), err)
		}
		status, err := f.db.GetStatusByURI(ctx, uid)
		if err != nil {
			if err == db.ErrNoEntries {
				// there are no entries for this status
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching status with id %s: %s", uid, err)
		}
		return status.Local, nil
	}

	if util.IsUserPath(id) {
		username, err := util.ParseUserPath(id)
		if err != nil {
			return false, fmt.Errorf("error parsing statuses path for url %s: %s", id.String(), err)
		}
		if _, err := f.db.GetLocalAccountByUsername(ctx, username); err != nil {
			if err == db.ErrNoEntries {
				// there are no entries for this username
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching account with username %s: %s", username, err)
		}
		l.Debugf("we own url %s", id.String())
		return true, nil
	}

	if util.IsFollowersPath(id) {
		username, err := util.ParseFollowersPath(id)
		if err != nil {
			return false, fmt.Errorf("error parsing statuses path for url %s: %s", id.String(), err)
		}
		if _, err := f.db.GetLocalAccountByUsername(ctx, username); err != nil {
			if err == db.ErrNoEntries {
				// there are no entries for this username
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching account with username %s: %s", username, err)
		}
		l.Debugf("we own url %s", id.String())
		return true, nil
	}

	if util.IsFollowingPath(id) {
		username, err := util.ParseFollowingPath(id)
		if err != nil {
			return false, fmt.Errorf("error parsing statuses path for url %s: %s", id.String(), err)
		}
		if _, err := f.db.GetLocalAccountByUsername(ctx, username); err != nil {
			if err == db.ErrNoEntries {
				// there are no entries for this username
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching account with username %s: %s", username, err)
		}
		l.Debugf("we own url %s", id.String())
		return true, nil
	}

	if util.IsLikePath(id) {
		username, likeID, err := util.ParseLikedPath(id)
		if err != nil {
			return false, fmt.Errorf("error parsing like path for url %s: %s", id.String(), err)
		}
		if _, err := f.db.GetLocalAccountByUsername(ctx, username); err != nil {
			if err == db.ErrNoEntries {
				// there are no entries for this username
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching account with username %s: %s", username, err)
		}
		if err := f.db.GetByID(ctx, likeID, &gtsmodel.StatusFave{}); err != nil {
			if err == db.ErrNoEntries {
				// there are no entries
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching like with id %s: %s", likeID, err)
		}
		l.Debugf("we own url %s", id.String())
		return true, nil
	}

	if util.IsBlockPath(id) {
		username, blockID, err := util.ParseBlockPath(id)
		if err != nil {
			return false, fmt.Errorf("error parsing block path for url %s: %s", id.String(), err)
		}
		if _, err := f.db.GetLocalAccountByUsername(ctx, username); err != nil {
			if err == db.ErrNoEntries {
				// there are no entries for this username
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching account with username %s: %s", username, err)
		}
		if err := f.db.GetByID(ctx, blockID, &gtsmodel.Block{}); err != nil {
			if err == db.ErrNoEntries {
				// there are no entries
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching block with id %s: %s", blockID, err)
		}
		l.Debugf("we own url %s", id.String())
		return true, nil
	}

	return false, fmt.Errorf("could not match activityID: %s", id.String())
}
