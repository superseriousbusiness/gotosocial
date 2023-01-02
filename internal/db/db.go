/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

const (
	// DBTypePostgres represents an underlying POSTGRES database type.
	DBTypePostgres string = "POSTGRES"
)

// DB provides methods for interacting with an underlying database or other storage mechanism.
type DB interface {
	Account
	Admin
	Basic
	Domain
	Emoji
	Instance
	Media
	Mention
	Notification
	Relationship
	Session
	Status
	Timeline
	User
	Tombstone

	/*
		USEFUL CONVERSION FUNCTIONS
	*/

	// TagStringsToTags takes a slice of deduplicated, lowercase tags in the form "somehashtag", which have been
	// used in a status. It takes the id of the account that wrote the status, and the id of the status itself, and then
	// returns a slice of *apimodel.Tag corresponding to the given tags. If the tag already exists in database, that tag
	// will be returned. Otherwise a pointer to a new tag struct will be created and returned.
	//
	// Note: this func doesn't/shouldn't do any manipulation of the tags in the DB, it's just for checking
	// if they exist in the db already, and conveniently returning them, or creating new tag structs.
	TagStringsToTags(ctx context.Context, tags []string, originAccountID string) ([]*gtsmodel.Tag, error)
}
