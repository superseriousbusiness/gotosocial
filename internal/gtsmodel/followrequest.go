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

// FollowRequest represents one account requesting to follow another, and the metadata around that request.
type FollowRequest struct {
	// id of this follow request in the database
	ID string `bun:"type:CHAR(26),pk,notnull,unique"`
	// When was this follow request created?
	CreatedAt time.Time `bun:"type:timestamp,notnull,default:current_timestamp"`
	// When was this follow request last updated?
	UpdatedAt time.Time `bun:"type:timestamp,notnull,default:current_timestamp"`
	// Who does this follow request originate from?
	AccountID string  `bun:"type:CHAR(26),unique:frsrctarget,notnull"`
	Account   Account `bun:"-"`
	// Who is the target of this follow request?
	TargetAccountID string  `bun:"type:CHAR(26),unique:frsrctarget,notnull"`
	TargetAccount   Account `bun:"-"`
	// Does this follow also want to see reblogs and not just posts?
	ShowReblogs bool `bun:"default:true"`
	// What is the activitypub URI of this follow request?
	URI string `bun:",unique"`
	// does the following account want to be notified when the followed account posts?
	Notify bool
}
