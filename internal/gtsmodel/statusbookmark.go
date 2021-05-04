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

// StatusBookmark refers to one account having a 'bookmark' of the status of another account
type StatusBookmark struct {
	// id of this bookmark in the database
	ID string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull,unique"`
	// when was this bookmark created
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// id of the account that created ('did') the bookmarking
	AccountID string `pg:",notnull"`
	// id the account owning the bookmarked status
	TargetAccountID string `pg:",notnull"`
	// database id of the status that has been bookmarked
	StatusID string `pg:",notnull"`
}
