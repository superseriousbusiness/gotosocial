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

// Follow represents one account following another, and the metadata around that follow.
type Follow struct {
	// id of this follow in the database
	ID string `bun:"type:CHAR(26),pk,notnull,unique"`
	// When was this follow created?
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// When was this follow last updated?
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// Who does this follow belong to?
	AccountID string   `bun:"type:CHAR(26),unique:srctarget,notnull"`
	Account   *Account `bun:"rel:belongs-to"`
	// Who does AccountID follow?
	TargetAccountID string   `bun:"type:CHAR(26),unique:srctarget,notnull"`
	TargetAccount   *Account `bun:"rel:belongs-to"`
	// Does this follow also want to see reblogs and not just posts?
	ShowReblogs bool `bun:"default:true"`
	// What is the activitypub URI of this follow?
	URI string `bun:",unique,nullzero"`
	// does the following account want to be notified when the followed account posts?
	Notify bool
}
