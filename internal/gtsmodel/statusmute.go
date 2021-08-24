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

// StatusMute refers to one account having muted the status of another account or its own
type StatusMute struct {
	// id of this mute in the database
	ID string `bun:"type:CHAR(26),pk,notnull,unique"`
	// when was this mute created
	CreatedAt time.Time `bun:"type:timestamp,notnull,default:now()"`
	// id of the account that created ('did') the mute
	AccountID string   `bun:"type:CHAR(26),notnull"`
	Account   *Account `bun:"rel:belongs-to"`
	// id the account owning the muted status (can be the same as accountID)
	TargetAccountID string   `bun:"type:CHAR(26),notnull"`
	TargetAccount   *Account `bun:"rel:belongs-to"`
	// database id of the status that has been muted
	StatusID string  `bun:"type:CHAR(26),notnull"`
	Status   *Status `bun:"rel:belongs-to"`
}
