/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

// Application represents an application that can perform actions on behalf of a user.
// It is used to authorize tokens etc, and is associated with an oauth client id in the database.
type Application struct {
	ID           string    `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`        // id of this item in the database
	CreatedAt    time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt    time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	Name         string    `validate:"required" bun:",notnull"`                                             // name of the application given when it was created (eg., 'tusky')
	Website      string    `validate:"omitempty,url" bun:",nullzero"`                                       // website for the application given when it was created (eg., 'https://tusky.app')
	RedirectURI  string    `validate:"required,uri" bun:",nullzero,notnull"`                                // redirect uri requested by the application for oauth2 flow
	ClientID     string    `validate:"required,ulid" bun:"type:CHAR(26),nullzero,notnull"`                  // id of the associated oauth client entity in the db
	ClientSecret string    `validate:"required,uuid" bun:",nullzero,notnull"`                               // secret of the associated oauth client entity in the db
	Scopes       string    `validate:"required" bun:",notnull"`                                             // scopes requested when this app was created
}
