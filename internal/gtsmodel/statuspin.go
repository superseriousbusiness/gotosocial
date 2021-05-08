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

// StatusPin refers to a status 'pinned' to the top of an account
type StatusPin struct {
	// id of this pin in the database
	ID string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull,unique"`
	// when was this pin created
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// id of the account that created ('did') the pinning (this should always be the same as the author of the status)
	AccountID string `pg:",notnull"`
	// database id of the status that has been pinned
	StatusID string `pg:",notnull"`
}
