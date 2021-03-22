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

package gtsmodel

import "time"

// Status represents a user-created 'post' or 'status' in the database, either remote or local
type Status struct {
	// id of the status in the database
	ID             string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull"`
	// uri at which this status is reachable
	URI            string `pg:",unique"`
	// web url for viewing this status
	URL            string `pg:",unique"`
	// the html-formatted content of this status
	Content        string
	// when was this status created?
	CreatedAt      time.Time `pg:"type:timestamp,notnull,default:now()"`
	// when was this status updated?
	UpdatedAt      time.Time `pg:"type:timestamp,notnull,default:now()"`
	// is this status from a local account?
	Local          bool
	// which account posted this status?
	AccountID      string
	// id of the status this status is a reply to
	InReplyToID    string
	// id of the status this status is a boost of
	BoostOfID      string
	// cw string for this status
	ContentWarning string
	// visibility entry for this status
	Visibility     *Visibility
}

// Visibility represents the visibility granularity of a status. It is a combination of flags.
type Visibility struct {
	// Is this status viewable as a direct message?
	Direct    bool
	// Is this status viewable to followers?
	Followers bool
	// Is this status viewable on the local timeline?
	Local     bool
	// Is this status boostable but not shown on public timelines?
	Unlisted  bool
	// Is this status shown on public and federated timelines?
	Public    bool
}
