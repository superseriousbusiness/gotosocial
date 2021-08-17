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

package db

import "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

type Timeline interface {
	// GetHomeTimelineForAccount returns a slice of statuses from accounts that are followed by the given account id.
	//
	// Statuses should be returned in descending order of when they were created (newest first).
	GetHomeTimelineForAccount(accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, DBError)

	// GetPublicTimelineForAccount fetches the account's PUBLIC timeline -- ie., posts and replies that are public.
	// It will use the given filters and try to return as many statuses as possible up to the limit.
	//
	// Statuses should be returned in descending order of when they were created (newest first).
	GetPublicTimelineForAccount(accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, DBError)

	// GetFavedTimelineForAccount fetches the account's FAVED timeline -- ie., posts and replies that the requesting account has faved.
	// It will use the given filters and try to return as many statuses as possible up to the limit.
	//
	// Note that unlike the other GetTimeline functions, the returned statuses will be arranged by their FAVE id, not the STATUS id.
	// In other words, they'll be returned in descending order of when they were faved by the requesting user, not when they were created.
	//
	// Also note the extra return values, which correspond to the nextMaxID and prevMinID for building Link headers.
	GetFavedTimelineForAccount(accountID string, maxID string, minID string, limit int) ([]*gtsmodel.Status, string, string, DBError)
}
