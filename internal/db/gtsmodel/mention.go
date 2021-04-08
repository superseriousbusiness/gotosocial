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

// Mention refers to the 'tagging' or 'mention' of a user within a status.
type Mention struct {
	// ID of this mention in the database
	ID string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull,unique"`
	// ID of the status this mention originates from
	StatusID string
	// When was this mention created?
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// When was this mention last updated?
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// Who created this mention?
	OriginAccountID string
	// Who does this mention target?
	TargetAccountID string
	// Prevent this mention from generating a notification?
	Silent bool
}
