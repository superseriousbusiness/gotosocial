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
	ID string `bun:"type:CHAR(26),pk,notnull,unique"`
	// ID of the status this mention originates from
	StatusID string  `bun:"type:CHAR(26),notnull,nullzero"`
	Status   *Status `bun:"rel:belongs-to"`
	// When was this mention created?
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// When was this mention last updated?
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// What's the internal account ID of the originator of the mention?
	OriginAccountID string   `bun:"type:CHAR(26),notnull,nullzero"`
	OriginAccount   *Account `bun:"rel:belongs-to"`
	// What's the AP URI of the originator of the mention?
	OriginAccountURI string `bun:",notnull"`
	// What's the internal account ID of the mention target?
	TargetAccountID string   `bun:"type:CHAR(26),notnull,nullzero"`
	TargetAccount   *Account `bun:"rel:belongs-to"`
	// Prevent this mention from generating a notification?
	Silent bool

	/*
		NON-DATABASE CONVENIENCE FIELDS
		These fields are just for convenience while passing the mention
		around internally, to make fewer database calls and whatnot. They're
		not meant to be put in the database!
	*/

	// NameString is for putting in the namestring of the mentioned user
	// before the mention is dereferenced. Should be in a form along the lines of:
	// @whatever_username@example.org
	//
	// This will not be put in the database, it's just for convenience.
	NameString string `bun:"-"`
	// TargetAccountURI is the AP ID (uri) of the user mentioned.
	//
	// This will not be put in the database, it's just for convenience.
	TargetAccountURI string `bun:"-"`
	// TargetAccountURL is the web url of the user mentioned.
	//
	// This will not be put in the database, it's just for convenience.
	TargetAccountURL string `bun:"-"`
	// A pointer to the gtsmodel account of the mentioned account.

}
