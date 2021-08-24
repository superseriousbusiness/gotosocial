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

// StatusFave refers to a 'fave' or 'like' in the database, from one account, targeting the status of another account
type StatusFave struct {
	// id of this fave in the database
	ID string `bun:"type:CHAR(26),pk,notnull,unique"`
	// when was this fave created
	CreatedAt time.Time `bun:"type:timestamp,notnull,default:current_timestamp"`
	// id of the account that created ('did') the fave
	AccountID string   `bun:"type:CHAR(26),notnull"`
	Account   *Account `bun:"rel:belongs-to"`
	// id the account owning the faved status
	TargetAccountID string   `bun:"type:CHAR(26),notnull"`
	TargetAccount   *Account `bun:"rel:belongs-to"`
	// database id of the status that has been 'faved'
	StatusID string  `bun:"type:CHAR(26),notnull"`
	Status   *Status `bun:"rel:belongs-to"`
	// ActivityPub URI of this fave
	URI string `bun:",notnull"`
}
