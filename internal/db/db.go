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

import (
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

const (
	// DBTypePostgres represents an underlying POSTGRES database type.
	DBTypePostgres string = "POSTGRES"
)

// DB provides methods for interacting with an underlying database or other storage mechanism (for now, just postgres).
// Note that in all of the functions below, the passed interface should be a pointer or a slice, which will then be populated
// by whatever is returned from the database.
type DB interface {
	Account
	Admin
	Basic
	Instance
	Mention
	Notification
	Relationship
	Status
	Timeline

	/*
		USEFUL CONVERSION FUNCTIONS
	*/

	// MentionStringsToMentions takes a slice of deduplicated, lowercase account names in the form "@test@whatever.example.org" for a remote account,
	// or @test for a local account, which have been mentioned in a status.
	// It takes the id of the account that wrote the status, and the id of the status itself, and then
	// checks in the database for the mentioned accounts, and returns a slice of mentions generated based on the given parameters.
	//
	// Note: this func doesn't/shouldn't do any manipulation of the accounts in the DB, it's just for checking
	// if they exist in the db and conveniently returning them if they do.
	MentionStringsToMentions(targetAccounts []string, originAccountID string, statusID string) ([]*gtsmodel.Mention, error)

	// TagStringsToTags takes a slice of deduplicated, lowercase tags in the form "somehashtag", which have been
	// used in a status. It takes the id of the account that wrote the status, and the id of the status itself, and then
	// returns a slice of *model.Tag corresponding to the given tags. If the tag already exists in database, that tag
	// will be returned. Otherwise a pointer to a new tag struct will be created and returned.
	//
	// Note: this func doesn't/shouldn't do any manipulation of the tags in the DB, it's just for checking
	// if they exist in the db already, and conveniently returning them, or creating new tag structs.
	TagStringsToTags(tags []string, originAccountID string, statusID string) ([]*gtsmodel.Tag, error)

	// EmojiStringsToEmojis takes a slice of deduplicated, lowercase emojis in the form ":emojiname:", which have been
	// used in a status. It takes the id of the account that wrote the status, and the id of the status itself, and then
	// returns a slice of *model.Emoji corresponding to the given emojis.
	//
	// Note: this func doesn't/shouldn't do any manipulation of the emoji in the DB, it's just for checking
	// if they exist in the db and conveniently returning them if they do.
	EmojiStringsToEmojis(emojis []string, originAccountID string, statusID string) ([]*gtsmodel.Emoji, error)
}
